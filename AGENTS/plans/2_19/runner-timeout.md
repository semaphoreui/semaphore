# Plan — Fail Hung Tasks When a Runner Dies or Restarts

## Goal

Today a task dispatched to a remote runner can hang **forever** in a
non-finished state when its runner disappears. Two operator-visible scenarios:

- **The runner fell off** (process killed, host gone, network partition). It
  stops polling and never reports a terminal status. The task stays
  `running`/`starting` indefinitely.
- **The runner restarted.** It comes back and resumes polling (so it *looks*
  healthy — `touched` is fresh), but it lost its in-memory job pool
  (`JobPool.runningJobs` is empty after a restart). It no longer reports the
  task it was executing and never will. The task stays `running` forever.

In both cases the work did **not** complete, so the task must be failed with a
clear message, its `End` set, and the usual finalization (finish webhook,
autorun children, state/Redis cleanup) run — exactly as if the task had
errored.

## Why this happens today

The dispatch path deliberately hands ownership of completion to the runner and
returns:

- `RemoteJob.Run` (`services/tasks/RemoteJob.go:204-213`) assigns
  `task.RunnerID`, persists, and returns. No node-local goroutine owns the
  task's completion — by design, so the task survives the dispatching node's
  death.
- The only backstop is `scheduleTimeout` (`RemoteJob.go:219-235`), which fires
  only when `util.Config.MaxTaskDurationSec > 0`. It is **opt-in** and
  **node-local**: if the duration is unset (the default) or the dispatching
  node restarts, nothing fails the task.
- The HA orphan cleaner (`pro_impl/services/ha/orphan_cleaner.go`) reconciles
  dead **nodes**, not dead **runners**. Its own comment is explicit
  (`orphan_cleaner.go:174-175`): *"dispatched to a runner that keeps executing
  it independently of the dead node. Leave it."* It assumes the runner is
  alive. When the runner is the thing that died, nothing ever GCs the task.
- The progress handler `UpdateRunner` (`api/runners/runners.go:286-289`)
  returns early on `body.Jobs == nil`. A restarted runner sends exactly that —
  an empty job list (`services/runners/job_pool.go:284-296` builds `Jobs` from
  the now-empty `runningJobs`) — so the server gets no signal that the task is
  gone.

Crucially, the two unfinished statuses behave differently on a **restart**:

- `starting` — the `GetRunner` handler still returns the task in `NewJobs`
  (`runners.go:146-158` keys `NewJobs` on `waiting`/`starting`), so a restarted
  runner re-pulls and re-runs it. This case largely self-heals (but can still
  hang if the runner never returns — covered by the liveness timeout below).
- `running` — the handler returns it only in `CurrentJobs`
  (`runners.go:229-234`), which the runner uses for monitoring, not execution.
  A restarted runner never re-runs it. **This is the core hang to fix**, and
  failing (not resuming) is correct: a partially-run job must not silently
  restart.

## Scope

In scope:

- A server-side **runner-task reconciler** that fails non-finished tasks whose
  assigned runner is dead (stale heartbeat) or has restarted since the task
  started.
- A **runner liveness signal** strong enough to distinguish "alive",
  "dead/silent", and "restarted" — heartbeat staleness plus a per-process
  generation marker.
- Wiring the reconciler into both the single-node task pool and the HA orphan
  cleaner, funnelling every failure through one idempotent helper.
- A configuration knob for the death threshold, with a safe default.

Out of scope:

- Resuming or retrying a task on another runner. We fail; retry policy is a
  separate decision.
- Killing the actual remote process (a dead/restarted runner has none to
  kill).
- Changing how runners are *selected* for dispatch (`RemoteJob.Run` two-pass
  logic stays as is).
- The cosmetic runner Version/Platform/Uptime UI work — tracked in
  `runner-version-platform-uptime.md`. This plan *reuses* the `started_at`
  field proposed there (see "Dependency" below).

## Design Summary

Three layers, each independently valuable, sharing one failure path:

1. **Heartbeat-staleness death detection (handles "fell off").** Each runner
   already updates `runner.touched` on every `GetRunner` poll
   (`api/runners/runners.go:117`). A reconciler periodically scans every
   non-finished task with a `RunnerID` and loads its runner. If
   `now - runner.touched > RunnerDeadTimeoutSec`, the runner is presumed dead
   and the task is failed.

2. **Generation-based restart detection (handles "restarted").** The runner
   reports a marker that changes on every process start — its `started_at`
   timestamp (reused from `runner-version-platform-uptime.md`) or, if that
   field is not present, a per-process random `session_id`. The server stores
   the current marker on the runner row. The reconciler fails any non-finished
   task whose owning runner's current generation is **newer** than the task's
   start: `runner.started_at > task.Start` ⇒ the runner booted after the task
   began ⇒ it cannot still be running it ⇒ fail. This catches the restart case
   even though `touched` is fresh.

3. **Reported-jobs reconciliation (defense in depth, optional).** Treat an
   actively-polling runner's reported job set as authoritative. Remove the
   `body.Jobs == nil` early return in `UpdateRunner` and, for tasks the server
   believes are `running`/`starting` on this runner but absent from the
   runner's reported set *after a dispatch grace window*, fail them. This
   reinforces layer 2 without depending on clock comparison, at the cost of a
   grace-window subtlety (just-dispatched tasks legitimately aren't reported
   yet).

Recommended: ship **layers 1 + 2** as the core fix (they fully cover both
scenarios and are simple to reason about); treat **layer 3** as a follow-up
reinforcement.

All three converge on a single idempotent helper, `failTaskRunnerLost`, that:
- re-loads the task and returns immediately if `Status.IsFinished()` (guards
  the race where a real terminal status arrives concurrently),
- logs a clear line (`"Runner #X lost: marking task failed"`),
- sets `Status = TaskFailStatus`, `End = now`, a descriptive `Message`,
- runs the existing finalization (`TaskPool.FinalizeRemoteTask`, which already
  has a `finalizing sync.Map` guard at `services/tasks/TaskPool.go:71`) so
  webhooks/autorun/state cleanup happen exactly once.

## Steps

### 1. Runner liveness data

- **Heartbeat:** already present (`runner.touched`). No change.
- **Generation marker:** add `started_at` to the runner (shared with
  `runner-version-platform-uptime.md`). The runner captures `time.Now()` once
  at startup and sends it on every poll (header `X-Runner-Started-At`, RFC3339,
  per that plan). The server persists it next to `touched` in the same UPDATE
  (`db/sql/global_runner.go:138-154` `TouchRunner`, extended to
  `TouchRunnerWithInfo`). If that plan does not land first, introduce a minimal
  `runner.session_id` (random string per process start) here instead — the
  reconciler only needs "did this change since dispatch".
- Add a nullable `started_at` (or `session_id`) column to the `runner` table
  for MySQL/Postgres/SQLite and the Bolt model.

### 2. Stamp tasks with their runner's generation (only if using `session_id`)

- If we compare against `started_at`, **no per-task column is needed**:
  `task.Start` already records when the task began, and the comparison
  `runner.started_at > task.Start` is sufficient.
- If we instead use an opaque `session_id`, record the dispatching runner's
  session on the task at assignment time (`RemoteJob.Run`, right where
  `task.RunnerID` is set, `RemoteJob.go:192-197`) via a new nullable
  `task.runner_session_id` column, and compare current-vs-recorded in the
  reconciler.

> Decision: prefer the `started_at` comparison — zero new task columns, reuses
> a field we want for the UI anyway.

### 3. The reconciler core (shared helper)

In `services/tasks/` add a `reconcileRunnerTasks` routine and the
`failTaskRunnerLost(tsk, runner, reason)` helper described in the Design
Summary. The reconcile pass:

```
for each tsk in pool.GetRunningTasks():               // services/tasks/TaskPool.go:144
    if tsk.Task.Status.IsFinished():          continue
    if tsk.Task.RunnerID == nil:              continue  // not dispatched yet
    runner := load(tsk.Task.RunnerID)
    if runner missing/deleted:                fail("runner no longer exists")
    // Layer 1 — dead/silent runner
    if now - runner.Touched > deadTimeout:    fail("runner stopped responding")
    // Layer 2 — restarted runner
    if runner.StartedAt != nil && tsk.Task.Start != nil &&
       runner.StartedAt.After(*tsk.Task.Start):
                                              fail("runner restarted; task lost")
```

Apply a **dispatch grace period** before layer 1/2 can fire on a brand-new
task (e.g. skip tasks whose `Start`/assignment is younger than the grace
window) so a task dispatched to a runner that is briefly between polls is not
killed prematurely. Grace ≈ a small multiple of the poll interval.

### 4. Run the reconciler in both deployment modes

- **Single node:** start a background goroutine from `TaskPool` (alongside the
  existing queue loop, `services/tasks/TaskPool.go:212`) ticking every
  `reconcileInterval` (≈30s). No global variables — the ticker lives on the
  pool instance.
- **HA / cluster:** extend `RedisOrphanCleaner.cleanupRunning`
  (`pro_impl/services/ha/orphan_cleaner.go:93-177`). Today the branch at
  `:174-175` ("dispatched to a runner → leave it") is precisely the hole.
  Replace "leave it" with the runner liveness/generation check from step 3,
  failing the task via the same helper and calling `removeStaleState` to clear
  Redis. The cleaner already runs every 60s and already loads each task — this
  is a localized change, not a new loop.

### 5. (Optional, layer 3) Authoritative reported-jobs reconciliation

- In `UpdateRunner` (`api/runners/runners.go:271-345`) replace the
  `body.Jobs == nil` early return with logic that still reconciles: for every
  task the server believes is `running`/`starting` on this runner but **not**
  present in `body.Jobs`, and whose dispatch is older than the grace window,
  call `failTaskRunnerLost`. Keep the per-job ownership check that already
  exists (`runners.go:312-315`).
- This requires the runner to keep sending a poll even when it has no jobs
  (it already does — `sendProgress` runs on the request timer regardless),
  and the server to treat "absent from a fresh runner's set" as lost.

### 6. Configuration

- Add `RunnerDeadTimeoutSec` (default e.g. **60s**) and `ReconcileIntervalSec`
  (default **30s**) under `RunnerConfig` / top-level config
  (`util/config.go`), with env vars
  `SEMAPHORE_RUNNER_DEAD_TIMEOUT_SEC` / `SEMAPHORE_RUNNER_RECONCILE_INTERVAL_SEC`.
- Constraint: `RunnerDeadTimeoutSec` must be comfortably larger than the runner
  poll interval (a few multiples) so a healthy-but-slow runner is never killed.
  Document this. Regenerate `config.schema.yaml` via the config-schema skill.

### 7. Tests

- Unit-test `failTaskRunnerLost`: idempotent (no double-finalize when called
  twice), no-op on an already-finished task, sets `Status`/`End`/`Message`.
- Table-driven tests for the reconcile decision: alive runner (no-op), stale
  `touched` past threshold (fail), `started_at` after `task.Start` (fail),
  `started_at` before `task.Start` (no-op), task within grace window (no-op),
  runner deleted (fail), task already finished (no-op).
- Initialize `util.Config` / `util.Config.Runner` in a helper per the project
  test conventions; reset between tests.
- HA: a focused test of the modified `cleanupRunning` branch covering
  dead-runner-with-live-node (the previously-uncovered hole).

## Rollout

- Backend-only behavioral change plus one additive runner column
  (`started_at`/`session_id`). No backfill: a runner with a NULL marker simply
  skips layer 2 until it next polls; layer 1 still protects it.
- Defaults are conservative (60s death threshold) so the change is safe to
  enable by default. Operators with very long poll intervals can raise it.
- Single-node and HA reconcilers ship together and share the helper.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Killing a healthy runner's task during a transient network blip | `RunnerDeadTimeoutSec` is a multiple of the poll interval + a dispatch grace window; a single missed poll never trips it. |
| Race: runner reports completion at the same instant the reconciler fails the task | `failTaskRunnerLost` re-loads and bails on `IsFinished()`; `FinalizeRemoteTask`'s `finalizing` guard ensures single finalization. |
| Clock skew between runner-reported `started_at` and `task.Start` (server clock) | Compare with a small margin; or use an opaque `session_id` (step 2 fallback) which is skew-immune. |
| `starting` tasks being failed when they would have self-healed via `NewJobs` on restart | Re-running a `starting` task is acceptable; let layer 1's grace window cover the brief gap, and only fail if the runner is genuinely dead. Failing `running` tasks is the priority. |
| Layer 3 grace-window subtlety (just-dispatched task not yet reported) | Gate layer 3 on dispatch age; ship it as a follow-up after layers 1+2 are proven. |
| HA cleaner now writes task failures (previously only re-enqueued/GC'd) | Goes through the same idempotent helper and `removeStaleState`; covered by the new HA test. |

## Dependency

- Layer 2 reuses the runner `started_at` field defined in
  `runner-version-platform-uptime.md`. If this plan lands first, introduce a
  minimal `session_id` marker here and migrate to `started_at` later — the
  reconciler logic is identical.

## Follow-ups (not part of this plan)

- **Retry-on-runner-loss policy:** optionally re-enqueue (rather than fail) a
  task whose runner died, controlled per-template.
- **Operator visibility:** surface "runner lost" as a distinct task failure
  reason in the UI and in the runners table (pairs with the uptime plan's
  health column).
- **Push-based liveness:** a lightweight runner→server keepalive decoupled
  from job polling, for faster death detection than the poll interval allows.
