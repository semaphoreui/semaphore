# Runner liveness and lost tasks

How the server notices that a remote runner fell off or restarted, and what it does with the tasks
it had dispatched there. Before this a task dispatched to a runner that died stayed `running`
forever: dispatch hands completion to the runner and returns, the only backstop
(`MaxTaskDurationSec`) was opt-in and node-local, and a restarted runner polls with an empty job
list that the server ignored.

Status (2026-09-27): implemented on develop and this branch for single-node and for the
node-owned half of HA — PR #3937 (`runner_timeout_2`, merged 2026-06-11, first in v2.19.0-beta6),
migration `v2.19.2` (`runner.started_at`), `services/tasks/runner_reconciler.go`,
`util.RunnersConfig`. The HA orphan-cleaner half is **not** in the local `pro_impl` — see
"Open" below.

## Decisions and why

- **Two thresholds from one heartbeat (`runner.touched`).** Offline after
  `runners.offline_timeout_sec` (default 120): the runner gets no new tasks and its
  `starting`/`waiting` tasks are requeued. Lost after `runners.task_fail_timeout_sec` (default
  420, clamped to ≥ offline): its `running` tasks are failed. Between the two a running task is
  deliberately left alone — offline does not mean the job stopped, and a runner that reconnects
  still has its in-memory job pool. Nothing is resumed on another runner: a partially-run job must
  not silently restart, so `running` is failed, never reassigned.
- **Restart detection by process start time, not a session id.** The runner sends
  `X-Runner-Started-At` (RFC3339, captured once in `NewJobPool`) on every request; the server
  persists it next to `touched` in the same `TouchRunner` UPDATE. `runner.started_at > task.start`
  on a `running` task means the job pool is provably gone → fail immediately, no 7-minute wait.
  `starting` tasks are exempt because a restarted runner re-pulls them from `NewJobs`. Chosen over
  a per-task `session_id` column because it needs no task column and doubles as an uptime source.
- **Single-pass runner selection, offline excluded.** The old two-pass "prefer fresh, fall back
  to stale" loop was exactly how tasks landed on dead runners; `selectRunner` in `RemoteJob.go`
  now takes the first online runner with capacity, else `ErrAllRunnersBusy` and the task stays
  queued. The 30-minute `runnerActiveThreshold` constant is gone. Webhook-driven runners never
  poll, so `IsOnline` treats them as always online and a `starting` task on one is never requeued
  by staleness (the runner may still be booting in response to the webhook).
- **Exactly-once actions under the finalize lock.** Both `failTaskRunnerLost` and
  `requeueTaskRunnerOffline` take `TaskStateStore.TryFinalize` (Redis `SETNX` in HA, `sync.Map`
  single-node), re-read the DB row in HA, and bail if the task is finished or already reassigned.
  Needed because the HA queue is a plain `RPUSH` with no dedup — a double requeue would run the
  task twice.
- **Work partitioned by claim in HA.** The loop runs on every node but iterates
  `OwnedRunningRange()` (tasks whose `tasks:claim:<id>` names this node), so each task is
  reconciled by one node. Tasks with an expired claim belong to the orphan cleaner.
- **Config lives in a new `runners` section, not `RunnerConfig`.** `SEMAPHORE_RUNNERS_*` is the
  server's view of the fleet; `SEMAPHORE_RUNNER_*` configures a runner process. Reference rows are
  generated into `docs/docs/reference/configuration.md`.
- **Rejected for now — layer 3 (reported-jobs reconciliation).** Treating a polling runner's
  reported job set as authoritative needs a dispatch grace window; deferred until layers 1+2 are
  proven. `UpdateRunner` still returns 204 on `body.Jobs == nil`.
- **Follow-on fixes** that grew out of the same loop: `requeueUndispatchedTask` returns a
  `starting` stub with no runner whose dispatch goroutine died with a node restart
  (`b23c91bbe`); the reconciler falls back from project runner to global runner lookup
  (`9d3b10080`); runner `status` online/offline in the API and runners table (`903f82b53`,
  v2.19.5-beta2); debug diagnostics under the `runner` debug context (`e0043e4be`, unreleased).

## Open, broken or deferred

- ⚠️ 2026-09-27: the HA orphan cleaner was **never changed**. `AGENTS/tasks/2026-06-10-runner-timeout.md`
  reports `reconcileDispatchedTask` + `tryFinalizeLock` in `pro_impl/services/ha/orphan_cleaner.go`
  as done ("review-verified, could not compile"); no commit on any `pro_impl` branch contains
  `DecideRunnerTaskAction`, and `orphan_cleaner.go` on `pro_impl` `main` still logs "task of dead
  node keeps running on its runner" and leaves it. Consequence: in HA, a task whose owning node
  died *and* whose runner then died/restarted is never failed or requeued (only the
  `MaxTaskDurationSec` backstop applies). The decision function is exported for exactly this use.
- ⚠️ 2026-09-27: the task note's "30 s clock-skew margin" on the restart check does not exist —
  `DecideRunnerTaskAction` compares `StartedAt.After(taskStart)` directly, and the test case
  named "restart within skew margin" expects `RunnerTaskFail`. Skew between the runner clock and
  the server clock therefore fails a task started within the skew.
- A task in `stopping` on a lost runner is kept forever (decision returns `Keep` for anything but
  starting/running); needs manual stop.
- `Task.Message` ("Runner lost: …") is set in memory and logged but `UpdateTask` does not write the
  `message` column; the reason is only in the task log and server log.
- Layer 3 and a per-template "retry on runner loss" policy remain follow-ups; so does a
  push-based keepalive faster than the poll interval.

## Key files and references

- `services/tasks/runner_reconciler.go` — `DecideRunnerTaskAction`, loop, fail/requeue helpers.
- `services/tasks/RemoteJob.go` — `selectRunner`, `runnerExclusionReason`, dispatch debug log.
- `services/tasks/task_state_store.go` — `OwnedRunningRange`, `TryFinalize`;
  `pro_impl/services/tasks/redis_task_state_store.go` for the Redis version.
- `db/Runner.go` — `StartedAt`, `IsOnline`, `HasFreeCapacity`, `FillStatus`;
  `db/sql/migrations/v2.19.2.sql`; `db/sql/global_runner.go` `TouchRunner`.
- `services/runners/job_pool.go` — `startedAt`, `setCommonHeaders`;
  `api/runners/runners.go` — header parsing in `GetRunner`, `body.Jobs == nil` in `UpdateRunner`.
- `util/config.go` — `RunnersConfig`, `RunnersOfflineTimeout/TaskFailTimeout/ReconcileInterval`.
- `pro_impl/services/ha/orphan_cleaner.go` — the unported branch.
- Tests: `services/tasks/runner_reconciler_test.go`, `db/Runner_test.go`.
- Sources: `AGENTS/plans/2_19/runner-timeout.md`, `AGENTS/tasks/2026-06-10-runner-timeout.md`.
  PR https://github.com/semaphoreui/semaphore/pull/3937.
