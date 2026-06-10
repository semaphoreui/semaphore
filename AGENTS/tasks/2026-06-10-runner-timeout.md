# Runner Timeout — Fail/Reassign Tasks of Offline Runners

**Date:** 2026-06-10
**Plan:** [`AGENTS/plans/2_19/runner-timeout.md`](../plans/2_19/runner-timeout.md)
**Status:** ✅ Done (layers 1 + 2; layer 3 deliberately deferred per plan)

## Goal

A task dispatched to a remote runner could hang forever in `starting`/`running`
when the runner fell off or restarted. Now:

- **Offline** (`touched` stale > **2 min**, `runners.offline_timeout_sec`): the
  runner gets no new tasks; its `starting` tasks are **reassigned** to another
  runner.
- **Lost** (`touched` stale > **7 min**, `runners.task_fail_timeout_sec`): its
  `running` tasks are **failed**. A runner that reconnects within the window
  kept its in-memory job pool and simply continues — nothing is failed.
- **Restarted** (reported `started_at` newer than the task's start): `running`
  tasks are failed **immediately** (job pool provably gone); `starting` tasks
  self-heal — the restarted runner re-pulls them from `NewJobs`.

## What was implemented

### Runner liveness data

- **`db/Runner.go`** — new `StartedAt *time.Time` (`db:"started_at"`) +
  `Runner.IsOnline(now, offlineTimeout)` (webhook runners are always "online";
  poll-based runners need a fresh `Touched`).
- **`db/sql/migrations/v2.18.8.sql`** (+ `.err.sql`) — nullable
  `runner.started_at` column; registered in **`db/Migration.go`**.
- **`db/sql/global_runner.go`** — `TouchRunner` now persists `started_at`
  together with `touched` (same UPDATE).
- **`services/runners/job_pool.go`** — `JobPool.startedAt` captured once at
  process start (`tz.Now()` in `NewJobPool`); sent on every poll/progress
  request via the `X-Runner-Started-At` header (RFC3339, new
  `setCommonHeaders` helper alongside `X-Runner-Token`).
- **`api/runners/runners.go`** — `GetRunner` parses the header and stores it on
  the runner before `TouchRunner` (invalid values are logged and ignored).

### Runner selection: single pass, offline excluded

- **`services/tasks/RemoteJob.go`** — the two-pass
  prefer-active/fall-back-to-stale loop is replaced by single-pass
  `selectRunner()`: candidate = online (`IsOnline`) + has capacity. The
  stale-runner fallback was exactly how tasks landed on dead runners. The
  `runnerActiveThreshold = 30 * time.Minute` constant is deleted —
  `runners.offline_timeout_sec` (2 min) replaces it. No online runner with
  capacity ⇒ `ErrAllRunnersBusy` ⇒ task stays queued, as before.

### Reconciler

- **`services/tasks/runner_reconciler.go`** (new):
  - `DecideRunnerTaskAction(status, taskStart, runner, now, offlineTimeout,
    taskFailTimeout)` — exported pure decision function (also used by the HA
    cleaner). Encodes the table from the plan, incl.: runner deleted
    (requeue/fail by status), restart detection with a 30s clock-skew margin,
    webhook-runner exception for `starting` tasks (a one-off runner may still
    be booting in response to the dispatch webhook), `Touched == nil`
    handling, and the deliberate no-op for `running` tasks between 2 and
    7 minutes of silence (recovery window).
  - `TaskPool.reconcileRunnerTasks(now)` — scans `GetRunningTasks()`, skips
    undispatched/finished tasks, loads each runner, applies the decision.
  - `TaskPool.failTaskRunnerLost(tsk, runner, reason)` — re-checks the
    persisted status (HA: re-reads the DB row), sets `TaskFailStatus` +
    message, runs `FinalizeRemoteTask` (dedups cluster-wide via the state
    store's `TryFinalize` lock).
  - `TaskPool.requeueTaskRunnerOffline(tsk, runnerID, reason)` — no-op unless
    the task is still `starting`/`waiting` and still owned by that runner;
    clears `RunnerID`, resets to `waiting`, persists, then re-enqueues via the
    same `EventTypeRequeued` flow as the `ErrAllRunnersBusy` path. Clearing
    `RunnerID` removes the task from the old runner's `NewJobs`; late progress
    reports are rejected by `UpdateRunner`'s ownership check.
  - Loop wired in **`services/tasks/TaskPool.go`** `Run()`
    (`go p.runnerTasksReconcileLoop()`, ticking every
    `runners.reconcile_interval_sec`, default 30s). No globals — everything
    lives on the pool instance.

### HA orphan cleaner

- **`pro_impl/services/ha/orphan_cleaner.go`** — the dead-owner "dispatched to
  a runner → leave it" hole is closed: new `reconcileDispatchedTask` loads the
  runner and applies the same `tasks.DecideRunnerTaskAction`. Requeue clears
  `RunnerID` + sets `waiting` in the DB **before** clearing Redis state and
  re-enqueueing (keeps concurrent passes idempotent); fail re-reads the row,
  bails on `IsFinished()`, then fails it in the DB and clears Redis state —
  same pattern as the existing `MaxTaskDurationSec` backstop in that file.
  Live-owner tasks are reconciled by the owning node's pool loop.

### Configuration

- **`util/config.go`** — new `RunnersConfig` section (json key `runners`),
  **separate from `RunnerConfig`**: server-side fleet settings use the
  `SEMAPHORE_RUNNERS_*` env prefix; `SEMAPHORE_RUNNER_*` stays for the runner
  process itself.
  - `offline_timeout_sec` — default **120** — `SEMAPHORE_RUNNERS_OFFLINE_TIMEOUT_SEC`
  - `task_fail_timeout_sec` — default **420** — `SEMAPHORE_RUNNERS_TASK_FAIL_TIMEOUT_SEC`
  - `reconcile_interval_sec` — default **30** — `SEMAPHORE_RUNNERS_RECONCILE_INTERVAL_SEC`
  - Accessors `RunnersOfflineTimeout()` / `RunnersTaskFailTimeout()` /
    `RunnersReconcileInterval()` apply defaults when the section is absent and
    clamp `task_fail` to ≥ `offline` (the plan's "validate at config load").
- **`config.schema.yaml`** — `runners` property + `$defs/RunnersConfig`
  (meta-validated against JSON Schema draft 2020-12).

## Tests

All green (`GOWORK=off go test ./db/... ./services/tasks/ ./services/runners/
./api/runners/ ./util/ -count=1`):

- **`db/Runner_test.go`** — `IsOnline` table (fresh/at-threshold/stale/never
  polled/webhook variants).
- **`services/tasks/runner_reconciler_test.go`**:
  - `TestDecideRunnerTaskAction` — 18-case table: alive (no-op), offline
    `starting` (requeue), 2–7 min silence on `running` (**no-op**, recovery
    window), >7 min (fail), restart → fail running immediately / keep starting,
    restart within skew margin (no-op), runner deleted (requeue/fail),
    finished/stopping (no-op), webhook-runner exception, never-polled.
  - `TestSelectRunner` — offline excluded; all-offline ⇒ nil
    (`ErrAllRunnersBusy`, task stays queued); webhook selectable regardless of
    heartbeat; at-capacity skipped; revived runner selectable again.
  - `TestRequeueTaskRunnerOffline` (+ noop-when-running) — clears and
    **persists** `RunnerID=nil`, resets to `waiting`, enqueues, emits
    `EventTypeRequeued`; refuses to touch a task that started running.
  - `TestFailTaskRunnerLost` — fails + finalizes (`EventTypeFinished`),
    persists status/End; second call is a no-op (idempotent).

## HA (cluster) verification

`go p.runnerTasksReconcileLoop()` starts on **every** node. Initially the
loop scanned `RunningRange()` — the **cluster-wide** running set — so N nodes
reconciled every task each ~30s (N× duplicated work and a requeue race, see
below). The work is now **partitioned by the existing claims**: the loop uses
the new `TaskStateStore.OwnedRunningRange()`, which in HA filters the shared
running set by the per-task claim value (`tasks:claim:<id>` = owning node ID,
set in `ClaimAndDequeue`, kept alive by `refreshClaims`; claims read in one
Redis pipeline, foreign tasks are not even hydrated). The memory store owns
everything, so single-node behavior is unchanged.

Resulting partition: each task is reconciled by exactly **one** live node
(its claim owner); tasks whose claim expired (owner died) are picked up by
the HA orphan cleaner's `reconcileDispatchedTask`. With 10 nodes that is ~1
reconcile pass per task per interval instead of 10.

Audit of the two action paths (locks kept as defense for the remaining
cross-actor overlap — cleaner vs. owner around claim expiry):

- **Fail path — safe as designed.** Concurrent `failTaskRunnerLost` on several
  nodes is serialized by `FinalizeRemoteTask`'s `TryFinalize` (Redis `SETNX`)
  plus the DB `End` re-check; the double `UpdateTask(fail)` before it is
  idempotent.
- **Requeue path — race found and fixed.** `RedisTaskStateStore.Enqueue` is a
  plain `RPUSH` with no dedup, and `ClaimAndDequeue` removes exactly one copy
  (`LREM 1`). Two nodes passing the requeue guards in the same tick would
  push the task ID twice; the surviving duplicate gets claimed after the task
  finishes (`SetStatus` does not block a success→starting transition) — i.e.
  a **second execution**. Fix: `requeueTaskRunnerOffline` now takes the same
  cluster-wide finalize lock (`state.TryFinalize`/`DeleteFinalize`) before the
  under-lock DB re-check; the loser either fails the `SETNX` or observes the
  cleared `RunnerID`. Exactly-once requeue.
- **Pool reconciler vs HA orphan cleaner.** Both can act on the same
  dead-node task. The cleaner's `reconcileDispatchedTask` now acquires the
  same Redis lock key (`tasks:finalize:<id>`, helper `tryFinalizeLock`) and
  re-loads the task row under the lock before requeueing/failing, so the two
  mechanisms are mutually exclusive. (The cleaner stays valuable as a
  fallback: it works from DB rows and does not depend on a successful
  `getOrHydrate`.)
- Residual, accepted: while the lock is briefly held by a reconciler, a
  racing legitimate `FinalizeRemoteTask` (runner reporting terminal status at
  that exact moment) bails on `TryFinalize` and does not retry; the finished
  task's Redis state is then GC'd by the cleaner's existing finished-in-DB
  branch. Window is milliseconds and requires the offline runner to report
  in that instant.
- Duplicate "Runner lost ..." task-log lines are possible when two nodes pass
  the pre-lock check together — cosmetic only.

## Deviations from the plan / notes

- **Layer 3** (authoritative reported-jobs reconciliation in `UpdateRunner`)
  is not implemented — the plan itself ships it as a follow-up after 1+2.
- **Grace window**: realized as (a) the webhook-runner exemption for
  `starting` tasks and (b) the 30s restart skew margin — for poll-based
  runners no extra grace is needed, since dispatch requires a fresh heartbeat
  and the offline threshold itself is the grace.
- **HA cleaner** shares the *decision* (`DecideRunnerTaskAction`) but not the
  *failure helper*: it has no `TaskPool`, so it fails tasks directly in the DB
  like the existing `MaxTaskDurationSec` backstop there (no finish webhook /
  autorun on that path — pre-existing behavior of the cleaner).
- `Task.Message` is set in memory/log but `UpdateTask` does not persist a
  `message` column (pre-existing); the reason is visible in the task log
  output and server logs.
- **`pro_impl` does not compile in this checkout** for a pre-existing reason:
  it is on branch `workflows`, which imports
  `services/tasks/artifacts` that does not exist on `develop`
  (`pro_impl/services/server/workflow_svc.go:12`). The orphan-cleaner change
  is review-verified but could not be compile-verified; main-module build,
  vet and tests pass (`GOWORK=off`).
- A task in `stopping` on a lost runner still hangs until stopped manually —
  out of scope per plan; candidate follow-up alongside layer 3.
