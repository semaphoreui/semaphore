# Performance bottlenecks — consolidated findings

Source: two static studies dated 2026-06-04 (`develop`), formerly `AGENTS/research/performance-*.md`, removed 2026-09-27 and kept only in git history. Every
finding re-checked on `develop` on 2026-09-27 (line numbers from that day; workflows redesign
merged 2026-09-28 as PR #4287); verdicts
**present** / **fixed** / **partly fixed** / **unverified**. Nothing was measured — see `scenarios.md`.
Amplifiers: `max_parallel_tasks` defaults to `9999` (TP-7) and the MySQL/Postgres pool is unbounded (DB-8).

## Task pool (`services/tasks/`)

- **TP-1 — Scheduler re-scans the whole queue on every event through one goroutine.**
  `TaskPool.handleQueue` (`services/tasks/TaskPool.go:321`) is the sole consumer of the unbuffered
  `queueEvents`; N completions cost O(N²) and producers block. Remedy: ready-task index per
  project/template, buffered/coalesced events. **present** — by-ID snapshot walk since `427c34c4`, still O(n).
- **TP-2 — `blocks()` issues a `GetProject` DB read per queued candidate per pass.**
  `TaskPool.blocks` (`TaskPool.go:653-687`). Remedy: cache project/template parallel limits, check
  in-memory counts first. **present**.
- **TP-3 — `RemoteJob.Run` busy-polled `GetTask` every second per remote task (+ a DB read in HA).**
  **fixed** — commit `427c34c4` "Ha remove owner (#3925)": `RemoteJob.Run` (`RemoteJob.go:203-283`)
  returns after dispatch; the runner `PUT` drives `TaskPool.FinalizeRemoteTask`; only a
  `time.AfterFunc` timeout remains. `GetTask` itself (`TaskPool.go:229-253`) is still an O(queue+running)
  scan over full copies — **present** but off the per-second path. Remedy: O(1) `GetByID`.
- **TP-4 — Runner load = scan of the running set per candidate runner.**
  `GetNumberOfRunningTasksOfRunner` (`TaskPool.go:166`) via `selectRunner` (`RemoteJob.go:246-250`).
  Remedy: `map[runnerID]int` kept in `onTaskRun`/`onTaskStop`. **present**, once per task start.
- **TP-5 — Queue is a slice; every removal is a linear search plus O(n) splice.**
  `DequeueAt/ClaimAndDequeue/DequeueByID` (`services/tasks/task_state_store.go:207-244`).
  Remedy: O(1)-removal structure. **present**.
- **TP-6 — One `RWMutex` for all state; hot readers copy whole collections under it.**
  `task_state_store.go:164-170`; `QueueRange` (:246), `RunningRange` (:283), `Snapshot` (:372).
  Remedy: remove full-copy callers, finer locks, allocation-free lookups. **present** for the memory
  store; the HA store (`pro/services/tasks/task_state_store_factory.go`) is **unverified**.
- **TP-7 — Unlimited parallelism by default, plus a goroutine parked 1 s per admitted task.**
  `util/config.go:167` (runner) and `:654` (server) default `9999`; `runTask` (`TaskPool.go:482-485`).
  Remedy: bounded worker pool with a sane default; drop the sleep. **present**.
- **TP-8 — Alerts sent synchronously on the status transition with no HTTP timeout.**
  `TaskRunner.SetStatus` (`services/tasks/TaskRunner_logging.go:142-152`) calls seven senders inline via
  bare `http.Post` (`services/tasks/alert.go:185-501`); mail re-reads each user (`:95`); `alertInfos`
  panics on DB error (`:545`). Remedy: background worker, client timeouts. **present** — `suppress_error_alerts` (`8a62e60d`) only reduces frequency.

## Task output capture and streaming

- **OUT-1 — Each stdout line is JSON-marshaled once per user and pushed synchronously before persistence.**
  `LogWithTime`/`sendToWs` (`TaskRunner_logging.go:28-55`); `t.users` = project users + all admins
  (`services/tasks/TaskRunner.go:504-528`). Hub back-pressure reaches `bufio.Scanner` and stalls the
  subprocess. Remedy: marshal once, async/lossy delivery, skip when nobody subscribes. **present** — in
  HA it is one `broadcaster.Publish` per user per line (`api/sockets/handler.go:184-193`).
- **OUT-2 — Unbuffered hub `broadcast`; one goroutine scans every connection per message.**
  `api/sockets/pool.go:49,65-82`. Remedy: buffered/lossy channel, connections indexed by user. **present**.
  ⚠️ 2026-09-27: the source called the per-connection `default:` a harmless drop; today a full
  `conn.send` closes the websocket (`pool.go:72-80`), so a slow viewer of a chatty task is disconnected.
- **OUT-3 — `db.StoreSession` opened per record inside the batched log writer.** **fixed** — commit
  `b92ec94b` "remove permanent connection flag" (2026-06-07) deleted `StoreSession`; `writeLogs`
  (`TaskPool.go:411-460`) is a plain loop plus one `InsertTaskOutputBatch`. `MoveToNextStage` still runs
  per record; its Pro-side DB cost is **unverified**.
- **OUT-4 — `unique(task_id, time)` fails a 500-line batch when two lines share a timestamp.**
  ⚠️ 2026-09-27: wrong — `db/sql/migrations/v2.2.1.sql` recreated `task__output` without the
  constraint (`v2.15.1.sqlite.sql:403` has none) and gorp's `SetUniqueTogether` (`db/sql/SqlDb.go:93`)
  is inert since gorp never creates tables here. Real residue: any batch-insert error drops the whole
  batch (`TaskPool.go:454-458`); reads order by `time, id` (`db/sql/task.go:409-411`). **fixed** / **present**.
- **OUT-5 — JSON output endpoint returns the entire log of a task.** `GetTaskOutput`
  (`api/projects/tasks.go:313-327`) → `GetTaskOutputs` with no LIMIT when `Count==0`
  (`db/sql/task.go:402-423`); the UI uses it (`web/src/components/TaskLogView.vue:467`); only
  `raw_output` chunks by 10 000 (`tasks.go:339-379`). Remedy: keyset paging on `id`. **present**.
- **OUT-6 — Remote runner buffers all pending output uncapped and ships it whole each second.**
  `runningJob.LogWithTime` (`services/runners/running_job.go:60-76`); `sendProgress`
  (`services/runners/job_pool.go:437-475`) trims only after a 2xx (`:526-540`); the server replays each
  record through `LogWithTime` (`api/runners/runners.go:450`), re-triggering OUT-1. Remedy: cap, chunked PUTs. **present**.
- **OUT-7 — 10 MB scanner buffer + 100 000-slot channel per pipe per task; a >10 MB line kills the task.**
  `TaskRunner.logPipe` (`TaskRunner_logging.go:176-217`). Remedy: grow lazily, truncate, shrink the
  channel once OUT-1 is async. **present**.
- **OUT-8 — Unconditional `fmt.Println` of every inbound websocket frame.** `api/sockets/handler.go:92`. **present**.

## Database queries and indexes (`db/sql/`)

- **DB-0 — Hot columns without indexes.** All `create index` statements in `db/sql/migrations/` through
  `v2.20.7.sql` were listed: none on `runner(token)`, `task(status)`, `task(created)`,
  `task(template_id, created)`, `task(project_id, created)`, `task(template_id, id)`,
  `task__output(task_id, stage_id)`, `task__output(task_id, time, id)`; `task__output(task_id)` and
  `task(project_id|template_id)` do exist. Remedy: one migration with the eight indexes. **present**.
- **DB-1 — Template list: correlated `last_task_id` subquery per row plus env and vault queries per template.**
  `getTemplates` (`db/sql/template.go:290`; `:412-422`). Remedy: batch with `IN`, add
  `task(template_id, id)`. **present** — last tasks are already fetched in one `IN` query (`:372-380`).
- **DB-2 — Task lists run `Fill()` per row: a `GetTask` for every `build_task_id`.** `getTasks`
  (`db/sql/task.go:335-340`) → `TaskWithTpl.Fill` (`db/Task.go:205-214`). Remedy: one `IN`. **present**.
- **DB-3 — Filters/sorts on unindexed `task.status` and `task.created`.** Status filter (`task.go:299`),
  retention `ORDER BY created` (`task.go:153-160`), stats (`SqlDb.go:964-983`). See DB-0. **present**.
- **DB-4 — Retention runs inline in `CreateTask`.** `clearTasks` (`task.go:121-168`, called at `:186`):
  random `count(*)`, `ORDER BY created ... OFFSET`, range `DELETE`. Remedy: background sweep, delete by
  `id` cutoff. **present**.
- **DB-6 — `task__output.stage_id` filtered without an index.** `GetTaskStageOutputs` (`task.go:426-435`). **present**.
- **DB-7 — Activity feed does a `GetTask` per task event.** `FillEvents`/`getEventObjectName`
  (`db/Event.go:75-134`); usernames memoised, tasks not. Remedy: batch, memoise. **present**.
- **DB-8 — MySQL/Postgres pool never bounded.** Only SQLite gets `SetMaxOpenConns(1)`
  (`db/sql/SqlDb.go:82-84`); no pool option in `util/config.go`. Remedy: configurable open/idle/lifetime. **present**.
- **DB-9 — BoltDB scanned all runners to match a token.** **fixed** — BoltDB removed (`46503d14`, #3876).
- **DB-10 — `GetTaskStats` aggregates the whole project history, no date floor, no cache.**
  `SqlDb.GetTaskStats` (`SqlDb.go:964-983`) from `api/projects/tasks.go:479-517`. Remedy: TTL cache,
  bounded default window, covering index. **present**.

## Inventory handling

- **INV-0 — Positive: the inventory is an opaque blob; no per-host loop in Go.** Keys installed once per
  inventory (`services/tasks/local_executor_inventory.go:15-30`); one `os.WriteFile` (`:98-105`).
  Still true. ⚠️ 2026-09-27: the source's `LocalJob_inventory.go`/`LocalJob.go` no longer exist —
  `LocalJob` became `LocalExecutor` (`local_executor*.go`).
- **INV-1 — Full inventory text inlined in the runner poll response on every poll while a job is starting.**
  `prepareRemoteJob` (`api/runners/runners.go:155-246`); `JobData.Inventory` (`services/runners/types.go:17`)
  and `Inventory.Inventory` (`db/Inventory.go:26`) are both `json:"inventory"`. Remedy: ship id + hash,
  fetch once; build the job once. **partly fixed** — see RUN-3 for what changed.
- **INV-2 — `db.Inventory` passed by value through every hop.** `types.go:17`, `TaskRunner.Inventory`.
  Cheap until marshaled or mutated. Remedy: pointer; keep the body out of logged DTOs. **present**.
- **INV-3 — Static inventory rewritten every run even when identical.**
  `LocalExecutor.installStaticInventory` (`local_executor_inventory.go:98-105`). Remedy: skip on matching
  content hash. **present** (low).

## API and frontend

- **API-5 — Admin "all tasks" copies the queued+running sets, no paging, polled every 10 s.**
  `api/tasks/tasks.go:43-71`; `web/src/views/Tasks.vue:93-95`. Remedy: paginate or push. **present**.
- **API-6 — Project history fetched 1 000 joined rows with the DB-2 N+1.** **partly fixed** — commit
  `11ea51d6` "tasks pagination (#3946)" (2026-06-12) gave `GetLastTasks` keyset paging (`before=<id>`,
  max 200; `api/projects/tasks.go:20,159-196`) and paged `History.vue`; design in
  `../ui-features/history-keyset-pagination.md`. `GetAllTasks` still hard-codes `Count = 1000`
  (`tasks.go:147-151`) and every row still runs `Fill`.
- **API-7 — `TouchSession` UPDATE on every cookie-authenticated request.** `api/auth.go:265` →
  `db/sql/session.go:74-78`. Remedy: debounce ~5 min in memory. **present**; API-token auth skips it.
- **API-9 — No read-path cache; `ClearTmpDir` runs synchronously in the handler.** `api/cache.go`.
  Remedy: TTL cache for token→runner, user, session; background the sweep. **present**.
- **API-10 — `/events` and `/project/{id}/events` return the whole table.** `getAllEvents`
  (`api/events.go:47-49`, `Count: 0`), routes `api/router.go:206,311`. **present** — no caller in `web/src`.

## Runners (server side of the runner protocol)

- **RUN-1 — Runner auth is `SELECT * FROM runner WHERE token=?` on an unindexed column, ~2×/s per runner.**
  `RunnerMiddleware` (`api/runners/runners.go:22-63`) → `GetRunnerByToken` (`db/sql/global_runner.go:12-32`).
  Progress `PUT` every 1 s (`services/runners/job_pool.go:267,288-291`), job `GET` every
  `runner.check_interval_seconds` (default 1 s). Remedy: unique index, short-TTL cache. **present**.
- **RUN-2 — `TouchRunner` UPDATE on every `GET` poll.** `runners.go:106` → `global_runner.go:138-157`
  (now also `started_at`); consumed at the 120 s offline timeout (`util/config.go:275`). Remedy:
  debounce to ~60 s. **present**.
- **RUN-3 — Job payload rebuilt, all secrets decrypted and RSA-encrypted per poll, for every polling runner.**
  **partly fixed** — RSA removed (`20c166a7` "remove enc key (#3984)", 2026-06-23); the `GET` serves only
  tasks whose `RunnerID` matches the caller (`runners.go:130-136`, `b23c91bb`). Still per poll while
  waiting/starting: `prepareRemoteJob` re-reads survey secrets, decrypts every key, inlines the inventory. Remedy: build once, mark dispatched.
- **RUN-4 — Runner list re-fetched from the DB on every task start.** `RemoteJob.Run`
  (`RemoteJob.go:226-232`). Remedy: brief cache. **present**.

Verified positives, do not regress: log batching at 500 records / 500 ms into one insert
(`TaskPool.go:44-45,382-403`); no lock held across I/O; HA claim right before dispatch; runner tags
load in one `IN`; API-token auth writes nothing.
