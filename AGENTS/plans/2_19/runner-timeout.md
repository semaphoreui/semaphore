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

In both cases the work did **not** complete. The required behavior depends on
how far the task got:

- A `starting` task never began executing on the runner — it is safe to
  **reassign** it to another (healthy) runner as soon as the original runner
  is considered offline.
- A `running` task may still be executing on a temporarily-unreachable runner
  — offline does **not** mean the job stopped. It gets a **recovery window**:
  if the runner comes back within the window it keeps reporting and the task
  continues; only after the window expires is the task failed with a clear
  message, its `End` set, and the usual finalization (finish webhook, autorun
  children, state/Redis cleanup) run — exactly as if the task had errored.

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
  runner re-pulls and re-runs it. This case self-heals on a *restart*; when
  the runner goes *offline* instead, the task hangs — fixed below by
  **reassigning** it to another runner.
- `running` — the handler returns it only in `CurrentJobs`
  (`runners.go:229-234`), which the runner uses for monitoring, not execution.
  A restarted runner never re-runs it. **This is the core hang to fix**, and
  failing (not resuming) is correct: a partially-run job must not silently
  restart.

## Scope

In scope:

- An **offline runner state**, derived from heartbeat staleness (default
  threshold **2 min**). Offline runners receive **no new tasks**, and their
  `starting` tasks are **reassigned** to another runner. Offline does *not*
  imply the runner's running jobs stopped — it may come back and continue.
- A server-side **runner-task reconciler** that (a) reassigns `starting`
  tasks off offline runners, and (b) fails `running` tasks whose runner has
  been silent past the recovery window (default **7 min**) or has restarted
  since the task started.
- A **runner liveness signal** strong enough to distinguish "alive",
  "offline/silent", and "restarted" — heartbeat staleness plus a per-process
  generation marker.
- A **simplified runner selection**: replace `RemoteJob.Run`'s two-pass
  prefer-active/fall-back-to-stale logic with a single pass over online
  runners. Offline runners are excluded from selection outright — no
  fallback.
- Wiring the reconciler into both the single-node task pool and the HA orphan
  cleaner, funnelling every failure through one idempotent helper.
- Configuration knobs for both thresholds, with safe defaults.

Out of scope:

- Resuming or retrying a **`running`** task on another runner — a
  partially-run job must not silently restart; we fail it. (Reassigning
  `starting` tasks *is* in scope — they never began executing.)
- Killing the actual remote process (a dead/restarted runner has none to
  kill).
- The cosmetic runner Version/Platform/Uptime UI work — tracked in
  `runner-version-platform-uptime.md`. This plan *reuses* the `started_at`
  field proposed there (see "Dependency" below).

## Design Summary

Two thresholds with distinct semantics, both derived from `runner.touched`:

- **Offline** — `now - touched > RunnerOfflineTimeoutSec` (default **2 min**).
  The runner stops receiving new tasks (excluded from dispatch selection) and
  its `starting` tasks are reassigned to another runner. Offline does **not**
  mean its running jobs stopped — the runner may still be executing them and
  may come back.
- **Lost** — `now - touched > RunnerTaskFailTimeoutSec` (default **7 min**).
  The runner is presumed dead; its `running` tasks are failed. If the runner
  returns *within* the window, it still has its in-memory job pool (it didn't
  restart, just lost connectivity), resumes reporting, and the tasks simply
  continue — nothing is failed.

Three layers, each independently valuable, sharing one failure path:

1. **Heartbeat-staleness detection (handles "fell off").** Each runner
   already updates `runner.touched` on every `GetRunner` poll
   (`api/runners/runners.go:117`). A reconciler periodically scans every
   non-finished task with a `RunnerID` and loads its runner, then applies the
   two thresholds above: past **offline** ⇒ reassign the task if `starting`;
   past **lost** ⇒ fail the task if `running`.

2. **Generation-based restart detection (handles "restarted").** The runner
   reports a marker that changes on every process start — its `started_at`
   timestamp (reused from `runner-version-platform-uptime.md`) or, if that
   field is not present, a per-process random `session_id`. The server stores
   the current marker on the runner row. The reconciler fails any `running`
   task whose owning runner's current generation is **newer** than the task's
   start: `runner.started_at > task.Start` ⇒ the runner booted after the task
   began ⇒ it cannot still be running it ⇒ fail **immediately** (no 7-minute
   wait — the job pool is provably gone, there is nothing to wait for). This
   catches the restart case even though `touched` is fresh. `starting` tasks
   are exempt: a restarted runner re-pulls them from `NewJobs` and they
   self-heal.

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

Two outcomes, two helpers:

**Reassignment** (`starting` on an offline runner) goes through
`requeueTaskRunnerOffline`, which:

- re-loads the task and returns if it is no longer `starting` (the runner may
  have just picked it up and reported `running`),
- clears `task.RunnerID` and resets `Status` to the queued state, so the
  normal dispatch loop selects another (healthy) runner,
- relies on the existing per-job ownership check in `UpdateRunner`
  (`runners.go:312-315`) to reject any late progress from the old runner.

**Failure** (everything else) converges on a single idempotent helper,
`failTaskRunnerLost`, that:

- re-loads the task and returns immediately if `Status.IsFinished()` (guards
  the race where a real terminal status arrives concurrently),
- logs a clear line (`"Runner #X lost: marking task failed"`),
- sets `Status = TaskFailStatus`, `End = now`, a descriptive `Message`,
- runs the existing finalization (`TaskPool.FinalizeRemoteTask`,
  `services/tasks/TaskPool.go:392`), which already deduplicates **across the
  cluster** via the state store's `TryFinalize`/`DeleteFinalize` lock
  (`task_state_store.go:123-129` — Redis `SETNX` keyed by task ID in HA, an
  in-process `sync.Map` in single-node) and, in HA mode, re-reads the task row
  and bails if it is already ended (`TaskPool.go:402-407`). So
  webhooks/autorun/state cleanup happen exactly once per task even when several
  nodes detect the lost runner in the same pass.

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

In `services/tasks/` add a `reconcileRunnerTasks` routine plus the
`failTaskRunnerLost(tsk, runner, reason)` and
`requeueTaskRunnerOffline(tsk, runner)` helpers described in the Design
Summary. The reconcile pass:

```
for each tsk in pool.GetRunningTasks():               // services/tasks/TaskPool.go:144
    if tsk.Task.Status.IsFinished():          continue
    if tsk.Task.RunnerID == nil:              continue  // not dispatched yet
    runner := load(tsk.Task.RunnerID)
    if runner missing/deleted:
        if tsk.Task.Status == starting:       requeue("runner no longer exists")
        else:                                 fail("runner no longer exists")
    // Layer 2 — restarted runner: job pool provably gone, no point waiting
    if runner.StartedAt != nil && tsk.Task.Start != nil &&
       runner.StartedAt.After(*tsk.Task.Start):
        if tsk.Task.Status == running:        fail("runner restarted; task lost")
        continue                              // starting self-heals via NewJobs
    silence := now - runner.Touched
    // Layer 1a — offline runner (2 min): reassign starting tasks
    if silence > offlineTimeout && tsk.Task.Status == starting:
                                              requeue("runner offline; reassigning")
    // Layer 1b — lost runner (7 min): fail running tasks
    if silence > taskFailTimeout && tsk.Task.Status == running:
                                              fail("runner stopped responding")
```

Note the asymmetry: between 2 and 7 minutes of silence a `running` task is
deliberately left alone — the runner is offline (gets no new work) but its
jobs may still be executing; if it reconnects within the window it resumes
reporting and the task finishes normally.

Apply a **dispatch grace period** before layer 1/2 can fire on a brand-new
task (e.g. skip tasks whose `Start`/assignment is younger than the grace
window) so a task dispatched to a runner that is briefly between polls is not
killed prematurely. Grace ≈ a small multiple of the poll interval.

### 4. Rework runner selection: single pass, offline excluded

Replace the two-pass selection in `RemoteJob.Run`
(`services/tasks/RemoteJob.go:159-176`) — pass 0 "prefer recently-touched
runners", pass 1 "fall back to stale ones" — with a **single pass**. The
stale-runner fallback is exactly how a task lands on a dead runner today;
with an explicit offline state it makes no sense and is removed.

- A runner is a candidate iff it is **online** — `touched` within
  `RunnerOfflineTimeoutSec` — and has capacity
  (`GetNumberOfRunningTasksOfRunner(r.ID) < r.MaxParallelTasks` or unlimited).
  Webhook-driven runners (`r.Webhook != ""`) don't poll, so `touched`
  staleness doesn't apply — they remain candidates as today.
- Delete the `runnerActiveThreshold = 30 * time.Minute` constant
  (`RemoteJob.go:23`); the config knob `RunnerOfflineTimeoutSec` replaces it.
- If no online runner has capacity, return `ErrAllRunnersBusy` as today — the
  task stays queued and is retried by the pool, picking up runners that come
  back online.
- The offline state is **derived** from `touched` at read time — no new DB
  column, no state machine to keep in sync. Optionally surface it in the
  runners API/UI (pairs with the uptime plan's health column).

### 5. Run the reconciler in both deployment modes

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

### 6. (Optional, layer 3) Authoritative reported-jobs reconciliation

- In `UpdateRunner` (`api/runners/runners.go:271-345`) replace the
  `body.Jobs == nil` early return with logic that still reconciles: for every
  task the server believes is `running`/`starting` on this runner but **not**
  present in `body.Jobs`, and whose dispatch is older than the grace window,
  call `failTaskRunnerLost`. Keep the per-job ownership check that already
  exists (`runners.go:312-315`).
- This requires the runner to keep sending a poll even when it has no jobs
  (it already does — `sendProgress` runs on the request timer regardless),
  and the server to treat "absent from a fresh runner's set" as lost.

### 7. Configuration

These are **server-side** settings describing how the server treats its
runner fleet. They must **not** live in `RunnerConfig`
(`util/config.go:120-151`) — that struct configures a runner *process* and
owns the `SEMAPHORE_RUNNER_*` env prefix. Server-side runner settings use the
**`SEMAPHORE_RUNNERS_*`** prefix (plural) to avoid collisions between the two
config surfaces.

- Add a new nested section in `ConfigType` (`util/config.go`), e.g.
  `RunnersConfig`, json key `runners`:
    - `OfflineTimeoutSec` — default **120** (2 min). Past this the runner is
      offline: no new dispatches, its `starting` tasks are reassigned. Env var
      `SEMAPHORE_RUNNERS_OFFLINE_TIMEOUT_SEC`.
    - `TaskFailTimeoutSec` — default **420** (7 min). Past this the runner's
      `running` tasks are failed. Env var
      `SEMAPHORE_RUNNERS_TASK_FAIL_TIMEOUT_SEC`.
    - `ReconcileIntervalSec` — default **30**. Env var
      `SEMAPHORE_RUNNERS_RECONCILE_INTERVAL_SEC`.
- Elsewhere in this plan these knobs are referred to as
  `RunnerOfflineTimeoutSec` / `RunnerTaskFailTimeoutSec` for brevity; the
  actual fields live on the `runners` section as above.
- Constraints: `OfflineTimeoutSec` must be comfortably larger than the runner
  poll interval (a few multiples) so a healthy-but-slow runner is never
  marked offline; `TaskFailTimeoutSec` must be ≥ `OfflineTimeoutSec`
  (validate at config load). Document both. Regenerate `config.schema.yaml`
  via the config-schema skill.

### 8. Tests

- Unit-test `failTaskRunnerLost`: idempotent (no double-finalize when called
  twice), no-op on an already-finished task, sets `Status`/`End`/`Message`.
- Unit-test `requeueTaskRunnerOffline`: clears `RunnerID`, resets status to
  queued; no-op if the task is no longer `starting` (runner picked it up
  concurrently).
- Table-driven tests for the reconcile decision: alive runner (no-op);
  `starting` + silence past offline threshold (reassign); `running` + silence
  between offline and fail thresholds (**no-op** — recovery window);
  `running` + silence past fail threshold (fail); `started_at` after
  `task.Start` + `running` (fail immediately); `started_at` after
  `task.Start` + `starting` (no-op — self-heals); `started_at` before
  `task.Start` (no-op); task within grace window (no-op); runner deleted
  (reassign if `starting`, fail if `running`); task already finished (no-op).
- Dispatch selection (single pass): offline runner excluded; webhook runner
  remains a candidate regardless of `touched`; online-but-at-capacity runner
  skipped; all runners offline ⇒ `ErrAllRunnersBusy` (task stays queued, not
  failed); previously-offline runner becomes selectable again after a fresh
  poll.
- Initialize `util.Config` / `util.Config.Runner` in a helper per the project
  test conventions; reset between tests.
- HA: a focused test of the modified `cleanupRunning` branch covering
  dead-runner-with-live-node (the previously-uncovered hole).

## Rollout

- Backend-only behavioral change plus one additive runner column
  (`started_at`/`session_id`). No backfill: a runner with a NULL marker simply
  skips layer 2 until it next polls; layer 1 still protects it.
- Defaults are conservative (**2 min** offline threshold, **7 min** recovery
  window before failing running tasks) so the change is safe to enable by
  default. Operators with very long poll intervals can raise both.
- Single-node and HA reconcilers ship together and share the helper.

## Risks & Notes

| Risk                                                                                                                                        | Mitigation                                                                                                                                                                                                                                                                                                                                                                                  |
|---------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Failing a healthy runner's `running` task during a transient network blip                                                                   | The 7-minute `RunnerTaskFailTimeoutSec` recovery window — many multiples of the poll interval; a runner that reconnects within it resumes reporting and nothing is failed.                                                                                                                                                                                                                  |
| Double execution: a `starting` task is reassigned, but the old runner had actually pulled it and starts it too                              | The window is narrow (`starting` flips to `running` on first progress report). `requeueTaskRunnerOffline` re-checks status before requeuing; clearing `RunnerID` removes the task from the old runner's `NewJobs`; the per-job ownership check in `UpdateRunner` rejects late reports from the old runner. Residual risk is inherent to reassignment and accepted for not-yet-running work. |
| `running` task on a truly dead runner shows as running for up to 7 min                                                                      | Accepted trade-off: the runner is offline within 2 min (no new work lands on it); the extra wait is the price of letting a live-but-disconnected runner finish its jobs. Layer 2 shortcuts this to immediate failure when a *restart* is detected.                                                                                                                                          |
| Race: runner reports completion at the same instant the reconciler fails the task — or two HA nodes detect the lost runner together         | `failTaskRunnerLost` re-loads and bails on `IsFinished()`; `FinalizeRemoteTask` then dedups cluster-wide via the state store's `TryFinalize`/`DeleteFinalize` lock (Redis `SETNX` in HA, `sync.Map` in single-node) plus the HA DB re-check, so finalization runs at most once.                                                                                                             |
| Clock skew between runner-reported `started_at` and `task.Start` (server clock)                                                             | Compare with a small margin; or use an opaque `session_id` (step 2 fallback) which is skew-immune.                                                                                                                                                                                                                                                                                          |
| Reassigned `starting` task loops forever across dying runners                                                                               | The reassignment goes through the normal queue, so each hop re-selects among *online* runners only; optionally cap reassignment attempts (small counter or rely on `MaxTaskDurationSec`).                                                                                                                                                                                                   |
| Behavior change: tasks that previously fell back to a stale runner now wait in the queue (`ErrAllRunnersBusy`) when every runner is offline | Intended: queued-and-waiting is recoverable, dispatched-to-a-dead-runner is the hang this plan fixes. The task is picked up as soon as any runner polls again.                                                                                                                                                                                                                              |
| Layer 3 grace-window subtlety (just-dispatched task not yet reported)                                                                       | Gate layer 3 on dispatch age; ship it as a follow-up after layers 1+2 are proven.                                                                                                                                                                                                                                                                                                           |
| HA cleaner now writes task failures (previously only re-enqueued/GC'd)                                                                      | Goes through the same idempotent helper and `removeStaleState`; covered by the new HA test.                                                                                                                                                                                                                                                                                                 |

## Dependency

- Layer 2 reuses the runner `started_at` field defined in
  `runner-version-platform-uptime.md`. If this plan lands first, introduce a
  minimal `session_id` marker here and migrate to `started_at` later — the
  reconciler logic is identical.

## Follow-ups (not part of this plan)

- **Retry-on-runner-loss policy for `running` tasks:** optionally re-enqueue
  (rather than fail) a *partially-run* task whose runner died, controlled
  per-template. (`starting` tasks are already reassigned by this plan.)
- **Operator visibility:** surface "runner lost" as a distinct task failure
  reason in the UI and in the runners table (pairs with the uptime plan's
  health column).
- **Push-based liveness:** a lightweight runner→server keepalive decoupled
  from job polling, for faster death detection than the poll interval allows.
