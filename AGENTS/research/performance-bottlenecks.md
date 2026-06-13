# Performance Research: Many Concurrent Tasks & Large Inventory (5000 hosts)

**Date:** 2026-06-04
**Scope:** Go backend of Semaphore (`semaphoreui/semaphore`).
**Question:** Where are the performance/scalability bottlenecks when (A) many tasks are running concurrently, and (B) an
inventory is very large (~5000 hosts)?
**Method:** Static read-through of the task pool, output/streaming path, SQL layer + migrations, API handlers, runner
polling protocol, and inventory lifecycle. Every "Critical/High" finding below was re-read against source and the exact
`file:line` confirmed.

---

## TL;DR — the shape of the problem

The system is architected around **in-memory task state scanned with O(n) loops**, **per-second polling loops** (both
runner→server and the server's own `RemoteJob.Run`), and a **synchronous, per-line, per-user websocket fan-out** on the
subprocess reader goroutine. None of these matter at 5 tasks / 50 hosts. All of them bite hard at hundreds of tasks /
5000 hosts.

Three structural decisions amplify everything else:

1. **`MaxParallelTasks` now defaults to `9999`** (effectively unlimited — `util/config.go:150,390`, commit `42e6c00d`).
   The only cheap global admission gate is gone.
2. **The MySQL/Postgres connection pool is never bounded** (`db/sql/SqlDb.go:82-84` sets a limit *only* for SQLite).
   Unlimited parallel tasks × unbounded pool → DB connection exhaustion.
3. **Hot lookups scan or full-copy collections** that used to be small: the in-memory running/queued sets, and the
   unindexed `runner.token` / `task.status` / `task.created` columns.

### Top findings, ranked

| #  | Finding                                                                                                             | Severity        | Scenario                  | Location                                                       |
|----|---------------------------------------------------------------------------------------------------------------------|-----------------|---------------------------|----------------------------------------------------------------|
| 1  | Synchronous per-line, **per-user** websocket marshal on the subprocess reader goroutine                             | **Critical**    | Many tasks / big output   | `services/tasks/TaskRunner_logging.go:27-54`                   |
| 2  | Runner poll auth = **full table scan** of unindexed `runner.token`, every poll, every runner                        | **Critical**    | Many runners              | `api/runners/runners.go:37`, `db/sql/global_runner.go:12-32`   |
| 3  | Per-record `db.StoreSession` **inside** the "batched" log writer (connect/close per line if not permanent conn)     | **High**        | Big output                | `services/tasks/TaskPool.go:300-352`                           |
| 4  | `unique(task_id, time)` constraint → a batch with 2 same-timestamp lines **fails and is silently dropped**          | **High**        | Big output                | `db/sql/SqlDb.go:93`                                           |
| 5  | Every queue event re-scans the whole queue **O(n²)** through one goroutine, calling `GetProject` (DB) per candidate | **High**        | Many tasks                | `services/tasks/TaskPool.go:206-268,463-496`                   |
| 6  | `RemoteJob.Run` **busy-polls every 1s per running task**; `GetTask` is an O(n) scan (+ a DB read per task in HA)    | **High**        | Many remote tasks         | `services/tasks/RemoteJob.go:209-251`, `TaskPool.go:142-166`   |
| 7  | Full **inventory body JSON-marshaled + RSA-encrypted on every runner poll** while a job is "starting"               | **High**        | Large inventory + runners | `api/runners/runners.go:138-262`                               |
| 8  | MySQL/Postgres **connection pool unbounded**; unlimited parallel tasks can exhaust DB connections                   | **High**        | Many tasks                | `db/sql/SqlDb.go:82-84`                                        |
| 9  | `GetTaskOutput` JSON endpoint loads the **entire** output table for a task, unpaginated                             | **High**        | Big output                | `api/projects/tasks.go:237`, `db/sql/task.go:375-397`          |
| 10 | Unlimited default parallelism + **`time.Sleep(1s)` per task goroutine** + ~3 goroutines/task                        | **High**        | Many tasks                | `util/config.go:150,390`, `TaskPool.go:367-370`                |
| 11 | `task.status`, `task.created`, `task__output.stage_id` **unindexed**; used in filters/sort/retention                | **High**        | Large history             | `db/sql/task.go:152-160,292`, `SqlDb.go:934-955`               |
| 12 | Per-poll **`TouchRunner` UPDATE** + per-request **`TouchSession` UPDATE** (write on every poll/request)             | **Medium-High** | Many runners / UI clients | `api/runners/runners.go:111-123`, `api/auth.go:264-268`        |
| 13 | `getTemplates` correlated subquery + **N+1** env/vault loads per template                                           | **High**        | Many templates            | `db/sql/template.go:283,406-412`                               |
| 14 | Single `RWMutex` + full-collection copies (`QueueRange`/`RunningRange`) under high-frequency contention             | **Medium**      | Many tasks                | `services/tasks/task_state_store.go:121-210`                   |
| 15 | Remote runner buffers the **entire task output** in an uncapped in-memory slice                                     | **High**        | Big output (runner)       | `services/runners/running_job.go:46-57`, `job_pool.go:288-339` |
| 16 | Synchronous, **timeout-less** alert HTTP/SMTP calls in the status-change hot path                                   | **Medium**      | Many tasks finishing      | `services/tasks/TaskRunner_logging.go:114-131`, `alert.go`     |
| 17 | 10 MB `bufio.Scanner` buffer + 100k-deep channel allocated **per pipe per task**; over-long line kills the task     | **Medium**      | Many tasks / big lines    | `services/tasks/TaskRunner_logging.go:153-201`                 |
| 18 | `GetTaskStats` GROUP-BY aggregation over **whole** task history, no date floor, no cache                            | **Medium**      | Large history             | `db/sql/SqlDb.go:934-955`                                      |
| 19 | Stray `fmt.Println` on every inbound websocket frame (stdout lock contention)                                       | **Low**         | Many clients              | `api/sockets/handler.go:91-92`                                 |

---

## Scenario A — Many tasks running concurrently

The damage is concentrated in three loops and one fan-out, all of which were sized for "a handful of tasks":

1. **The scheduler is a single goroutine doing O(n²) work.** `handleQueue` (`TaskPool.go:206-268`) is the only consumer
   of the **unbuffered** `queueEvents` channel. On every event — new task, finished task, requeue, and a 5-second tick —
   it re-scans the entire waiting queue from index 0. For each candidate it calls `blocks()` (`TaskPool.go:463-496`),
   which issues a `GetProject` **DB query** whenever the project has any active task. A burst of N completions = ~N
   events × O(N) scan × a DB round-trip per candidate. Producers block on the unbuffered channel until this one
   goroutine drains, so the scheduler becomes a global serialization point. **Finding 5.**

2. **Each running remote task spins its own 1-second poll.** `RemoteJob.Run` (`RemoteJob.go:209-251`) does
   `time.Sleep(1s)` → `GetTask(id)` forever. `GetTask` (`TaskPool.go:142-166`) **full-copies** the queue and the running
   map (`QueueRange()`/`RunningRange()`, `task_state_store.go:165-210`) and linear-scans them. In HA it also does a
   `GetTask` **DB** read per task per second. 300 remote tasks ⇒ ~300 DB reads/sec + ~300×(Q+R) pointer copies/sec of
   pure bookkeeping. **Finding 6.**

3. **Runner polls scan the global running set.** Each runner's GET poll (`runners.go:138`) calls `GetRunningTasks()` (
   full map copy under lock) and filters client-side; runner *selection* (`RemoteJob.go:164-179`) calls
   `GetNumberOfRunningTasksOfRunner` (another full scan) per candidate runner. O(runners × running) per task start. *
   *Findings 3-ish, 14.**

4. **Output fan-out throttles Ansible itself.** Every stdout line calls `sendToWs` (`TaskRunner_logging.go:41-54`) *
   *before** persistence, marshaling JSON **once per user** (project users + all admins) and pushing onto the *
   *unbuffered** hub `broadcast` channel. With many concurrent chatty tasks, the single hub goroutine + per-user marshal
   is O(lines × users × connections), and backpressure propagates up the channel chain to `bufio.Scanner`, stalling the
   subprocess. **Findings 1, 14.**

Amplifiers: unlimited default parallelism removes the global cap (**Finding 10**), the unbounded DB pool lets that
translate into connection exhaustion (**Finding 8**), and the per-record `StoreSession` in the log writer (**Finding 3
**) plus per-poll/per-request writes (**Finding 12**) pile writes onto the same DB — devastating on SQLite's single
writer lock.

## Scenario B — Large inventory (5000 hosts)

**Good news first** (a dedicated agent traced the whole inventory lifecycle): Semaphore treats the inventory as an *
*opaque blob**. It is never parsed, split per-host, or iterated in Go. There are **no per-host loops, no quadratic
string building, and no per-host key/file/syscall operations.** SSH/become keys are installed **once per inventory**
through a single in-process SSH agent regardless of host count (`services/tasks/LocalJob_inventory.go:15-28`,
`pkg/ssh/agent.go:170-212`). A static inventory is materialized with a **single** `os.WriteFile` (
`LocalJob_inventory.go:92-98`). Per-host connection fan-out is Ansible's concern (`forks`), not Semaphore's. So the "
5000 syscalls / O(hosts²)" class of bug **does not exist** on the Semaphore side.

**The real inventory cost is carrying the blob and re-serializing it:**

- **Remote runner path (Finding 7) is the only place that genuinely hurts.** A 5000-host static inventory is ~0.5–2 MB
  of text. On *every* `GetRunner` poll, the server walks all running tasks and, for any task still in
  `TaskStartingStatus`, appends the **entire** `db.Inventory` body into `RunnerState.NewJobs`, then JSON-encodes (
  byte-escapes) **and chunked-RSA-encrypts** the whole payload (`runners.go:138-262`). A starting task survives several
  polls until the runner grabs it, so the same multi-MB blob is re-encoded and re-encrypted every second, by every
  polling runner. Cost ≈ `O(inventory_bytes × polls × runners)` of alloc + escape + public-key crypto. **This is the
  headline inventory bottleneck.**
- **Local static path:** one `os.WriteFile` + one `[]byte(string)` copy per run (`LocalJob_inventory.go:92`). O(bytes)
  once per task, not cached between identical runs. Low. (**Finding, ranked Low.**)
- **Local file inventory:** cheapest — Semaphore only joins a path; Ansible reads the file (`LocalJob.go:352-357`).
  Touches zero host bytes.
- **Big output is correlated with big inventory:** a 5000-host run emits 10⁵+ stdout lines, which is exactly what makes
  Findings 1, 3, 4, 9, 15, 17 fire. *At 5000 hosts, the output path — not the inventory blob — is the dominant cost on
  the local execution path.*

---

## Detailed findings by subsystem

### 1. Task pool & concurrency (`services/tasks/`)

- **O(n²) scheduler through one goroutine + DB in the hot loop (Finding 5).** `handleQueue` (`TaskPool.go:206-268`)
  re-scans the full queue per event; `blocks()` (`TaskPool.go:463-496`) calls `store.GetProject` per candidate once a
  project has any active task. Unbuffered `queueEvents` (one reader) makes producers block on the scheduler.
  *Fix:* index "ready" tasks per project/template for O(1) pop; cache `project.MaxParallelTasks`/
  `Template.AllowParallelTasks` in memory; buffer/coalesce `queueEvents`; break the inner loop once the parallel cap is
  hit.
- **1s busy-poll + O(n) `GetTask` per remote task (Finding 6).** `RemoteJob.go:209-251` + `TaskPool.go:142-166`.
  `QueueRange()`/`RunningRange()` (`task_state_store.go:165-210`) allocate a full copy each call.
  *Fix:* add O(1) `GetByID` backed by the existing `running map[int]` (`task_state_store.go:124`); replace the poll with
  an event/condition signaled by the runner's progress `PUT` (which already calls `SetStatus`).
- **O(running) scan per runner candidate (Finding, Medium).** `GetNumberOfRunningTasksOfRunner` (`TaskPool.go:129-136`)
  inside the runner-selection double loop (`RemoteJob.go:164-179`).
  *Fix:* maintain a `map[runnerID]int` updated in `onTaskRun`/`onTaskStop`.
- **Slice-splice dequeue → O(n²) drain (Finding, Medium).** `DequeueAt` (`task_state_store.go:154-163`) shifts the tail
  per removal; `StopTasksByTemplate` (`TaskPool.go:583-612`) removes one-by-one.
  *Fix:* O(1) removal structure (linked list / tombstone-and-compact).
- **Single `RWMutex` + full-collection copies (Finding 14).** `MemoryTaskStateStore` (`task_state_store.go:121-127`)
  guards everything with one lock; `RunningRange`/`QueueRange`/`Snapshot` copy the whole collection while holding it.
  Not held across I/O (good), but taken at very high frequency by Findings 5/6 and the cluster dashboard `Snapshot()`.
  *Fix:* finer-grained locks or `sync.Map`; eliminate the high-frequency full-copy callers (Findings 5/6) so the lock is
  taken far less.
- **Unlimited default parallelism + per-task `Sleep` (Finding 10).** `MaxParallelTasks` default `9999` (
  `config.go:150,390`); `runTask` parks a goroutine for `time.Sleep(1s)` before `task.run()` (`TaskPool.go:367-370`);
  each local task also spawns 2 `logPipe` goroutines each spawning another (`TaskRunner_logging.go:64-65,158`) with
  100k-buffered channels.
  *Fix:* a bounded worker pool / semaphore with a sane default instead of `9999`; drop the unconditional `time.Sleep`.
- **Synchronous, timeout-less alerts in the status hot path (Finding 16).** `SetStatus` → `sendMailAlert`/
  `sendTelegramAlert`/… (`TaskRunner_logging.go:114-131`), each a blocking `http.Post`/SMTP with the default (
  no-timeout) client (`alert.go:171-487`); `sendMailAlert` re-fetches users from DB per alert (`alert.go:81`);
  `alertInfos()` can `panic` on a DB error (`alert.go:531`). A slow endpoint pins the task's goroutine, delaying
  `EventTypeFinished` and keeping the task in the running set (inflating every scan above).
  *Fix:* bounded background worker for alerts; explicit `http.Client` timeouts; reuse already-loaded user list.

*Verified-OK:* log batching design (`handleLogs` flushes by size 500 / every 500ms, `TaskPool.go:271-298`); locks are
never held across I/O; HA `TryClaim`/`DeleteClaim` placement; the schedule pool (`robfig/cron`, lock held only during
`Refresh`).

### 2. Task output capture & streaming

- **Synchronous per-line, per-user marshal on the reader goroutine (Finding 1, Critical).** `sendToWs` (
  `TaskRunner_logging.go:41-54`) marshals once *per user* and runs *before* the line is queued for DB. `t.users` = all
  project users **+ all admins** (`TaskRunner.go:377-395`). No "is anyone watching this task?" check.
  *Fix:* marshal **once per line**; make delivery async + lossy off the reader goroutine; skip fan-out when no
  subscriber for that task/project; let the hub filter by user.
- **Unbuffered hub `broadcast` channel, single goroutine, O(all-connections) per message (Finding, High).**
  `api/sockets/pool.go:48-87`. The per-connection `default:` drop is correct (one slow client can't block), but the
  unbuffered `broadcast` blocks every caller until the single hub goroutine picks up, and it re-scans all connections
  even for a user-targeted message.
  *Fix:* buffer + lossy `broadcast`; index connections by `userID`.
- **Per-record `StoreSession` inside the batch writer (Finding 3, High).** `writeLogs` (`TaskPool.go:300-352`) wraps
  `stage_parsers.MoveToNextStage` in `db.StoreSession` **per record** (lines 314-341). `StoreSession` (
  `db/Store.go:790-800`) does `Connect()/Close()` around the callback unless `PermanentConnection()` — i.e. a **DB
  connect/close per output line** on the single shared `handleLogs` goroutine. In this build `MoveToNextStage` is a
  no-op stub but still pays the wrapper cost per line; the Pro build does real per-line DB work.
  *Fix:* hoist the session out of the loop (one session per flush); only run stage parsing for stage-using apps; batch
  its DB effects.
- **`unique(task_id, time)` drops whole batches (Finding 4, High + correctness).** `SqlDb.go:93`
  `SetUniqueTogether("task_id","time")`; timestamps are per-line `tz.Now()`. At 5000 hosts many lines share a `time`; a
  single `InsertTaskOutputBatch` (`task.go:244-264`) is one multi-row INSERT, so a duplicate-timestamp pair fails the *
  *entire 500-line batch**, which is then merely logged and **silently dropped** (`TaskPool.go:347-350`). Data loss +
  perf cliff + log spam.
  *Fix:* drop the unique constraint (rows are keyed by autoincrement `id`); order output by `id`, not `time`; fall back
  to per-row insert on batch failure.
- **Unbounded `GetTaskOutput` JSON endpoint (Finding 9, High).** `api/projects/tasks.go:237` calls
  `GetTaskOutputs(..., RetrieveQueryParams{})`; with `Count==0` no LIMIT is applied (`task.go:386-388`), so the entire
  `task__output` for the task is loaded and marshaled into one JSON array. (The *raw* endpoint `tasks.go:258-298`
  paginates in 10000-row chunks — the JSON one does not.)
  *Fix:* paginate (keyset `WHERE task_id=? AND id > ?`), or point the UI at the streaming raw endpoint; add composite
  `(task_id, time, id)` index.
- **Remote runner buffers entire output in memory (Finding 15, High).** `running_job.go:46-57` appends every line to an
  **uncapped** `logRecords`; `sendProgress` (1s ticker, `job_pool.go:179,288-296`) PUTs the whole slice and only trims
  after success — a slow/unreachable server ⇒ unbounded growth; the server then replays every record through the
  synchronous per-line fan-out of Finding 1 in bursts (`runners.go:316-318`).
  *Fix:* cap/ring-buffer `logRecords`; chunk the payload; feed the receiver into the async coalesced broadcast path.
- **10 MB scanner buffer + 100k channel per pipe per task; over-long line kills the task (Finding 17, Medium).**
  `TaskRunner_logging.go:153-201`: `make([]byte, 10MB)` per pipe (stdout+stderr) per task; `scanner.Text()` copies a new
  string per line (GC churn ∝ output); a `>10MB` line triggers "token too long" and **aborts the run** (`:188-192`).
  *Fix:* start the buffer small and let it grow; truncate over-long lines instead of killing the task; shrink the 100k
  channel once the downstream (Findings 1, 3) is fixed.
- **Stray `fmt.Println` per inbound ws frame (Finding 19, Low).** `api/sockets/handler.go:91-92` — unconditional stdout
  write (serialized on the stdout lock) for every inbound frame from every client. Looks like leftover debug.
  *Fix:* delete it.

### 3. Database query patterns & indexes (`db/sql/`)

**Index inventory.** EXISTS: `task(template_id|project_id|integration_id|inventory_id|schedule_id)`;
`task__output(task_id)` and `(time)`; `event(project_id|user_id)`; `runner(project_id)` and `(registration_token)`. *
*MISSING on hot columns:** `runner(token)`, `task(status)`, `task(template_id, created)`, `task(project_id, created)`,
`task(template_id, id)`, `task__output(task_id, stage_id)`, `task__output(task_id, time, id)`.
*(Note: the `task__output(task_id)` index the brief worried about does
exist — `migrations/v2.15.1.sqlite.sql:413`, `v2.17.15.sql:6`. Good.)*

- **`runner.token` full scan on every poll (Finding 2, Critical).** `RunnerMiddleware` (`runners.go:37`) →
  `GetRunnerByToken` (`global_runner.go:12-32`) → `SELECT * FROM runner WHERE token=?` (`SqlDb.go:491-501`); no
  `project_id` predicate (global), **no index on `token`** (only `project_id` and `registration_token` exist). Bolt is
  O(N) in memory (`bolt/global_runner.go:12-30`). With N runners polling ~1/s and a `PUT`+`GET` per cycle, that is ~2N
  full scans/sec.
  *Fix:* unique index on `runner(token)`; in-memory token→runner cache with short TTL.
- **`task.status` / `task.created` unindexed (Finding 11, High).** Status filter (`task.go:292`), `GetTaskStats`
  group/filter on `created,status,start,end,user_id` (`SqlDb.go:934-955`), and `clearTasks` `ORDER BY created` **on the
  task-creation write path** (`task.go:152-160`).
  *Fix:* add `task(status)` (or `task(project_id,status)`), `task(template_id,created)`, `task(project_id,created)`.
- **`getTemplates` correlated subquery + N+1 (Finding 13, High).** Per-template
  `(SELECT id FROM task WHERE template_id=pt.id ORDER BY id DESC LIMIT 1)` (`template.go:283`) + a
  `GetTemplateEnvironments` and `GetTemplateVaults` query **per template** (`template.go:406-412`, `loadVaults=true`).
  200 templates ⇒ 400–600 queries/page.
  *Fix:* batch env/vault with `WHERE template_id IN (...)`; add `task(template_id, id)` for an index-only backward scan
  of last_task_id.
- **`getTasks` N+1 `Fill()` (Finding, Medium/High for deploy projects).** `task.go:314-319` loops `Fill()`;
  `TaskWithTpl.Fill` (`db/Task.go:184-196`) does a `GetTask` per row with a non-nil `BuildTaskID`. Up to ~1000 extra
  joins per history page.
  *Fix:* resolve all `BuildTaskID`s in one `WHERE id IN (...)`.
- **`clearTasks` retention on the write path (Finding, Medium).** Runs synchronously inside `CreateTask` (
  `task.go:121-168`): a `rand`-gated `count(*)`, an `ORDER BY created` scan, and a range `DELETE` — all without
  `task(created)`.
  *Fix:* the `(template_id, created)` index; move retention to a background sweep; delete by `id` cutoff (monotonic,
  index-backed).
- **`FillEvents` N+1 (Finding 7-DB, Medium).** `db/Event.go:73-114` calls `GetTask` per task-event (usernames are
  memoized, tasks are not). Hundreds of joins per activity feed.
  *Fix:* batch-resolve task object names with `IN`; memoize like usernames.
- **`GetTaskStats` unbounded aggregation, no cache (Finding 18, Medium).** `SqlDb.go:934-955` groups the whole project
  history with no date floor when `start` is absent.
  *Fix:* short-TTL cache per (project, template, range); default a bounded window; covering index
  `(project_id, template_id, created, status)`.
- **MySQL/Postgres pool unbounded (Finding 8, High).** `SqlDb.go:82-84` sets `SetMaxOpenConns(1)` for SQLite only;
  nothing for MySQL/Postgres. Go default = unlimited open / 2 idle. Unlimited parallel tasks each writing
  status/output/stage/event ⇒ connection-count blowup + constant idle churn.
  *Fix:* set `SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime` (configurable) right where the SQLite branch is.

### 4. API layer & runner polling

- **`runner.token` scan + `GetRunningTasks` per poll (Findings 2, 3-API, 7).** Covered above; the GET poll also ranges
  the global running set (`runners.go:138`, `GetRunningTasks` full-copies the map) and, for "starting" tasks, decrypts
  every secret + JSON + RSA-encrypts the whole payload including the inventory body, **every second per runner** even
  when there is no new job.
  *Fix:* index the running set by runner id; mark a job "dispatched" so it's built/encrypted once; long-poll or push
  instead of 1s poll; ship inventory by id+hash, not inline.
- **Per-poll / per-request writes (Finding 12, Medium-High).** `TouchRunner` UPDATE on every GET poll (
  `runners.go:111-123`, only needed at 30-min granularity per `RemoteJob.go:23`); `TouchSession` UPDATE on every
  authenticated request (`auth.go:264-268`, `last_active` only used to expire >7-day sessions). N runners + many UI
  tabs ⇒ steady write stream; brutal on SQLite's single writer.
  *Fix:* debounce both via an in-memory last-touch map (write only if older than ~60s / ~5min).
- **Admin "all tasks" + per-project task list, no pagination/cache (Findings 5-API, 6-API, Medium).**
  `api/tasks/tasks.go:43-71` copies the whole queued+running set per poll (admin Tasks view polls every 10s,
  `Tasks.vue:92`); `GetAllTasks` fetches up to 1000 joined rows with a conditional N+1 (`tasks.go:100`,
  `task.go:269-322`).
  *Fix:* pagination (keyset on `id desc`); serve admin list from a rate-limited cached snapshot or push over websocket.
- **No read-path cache anywhere (Finding 9-API, Low-but-enabling).** `api/cache.go` only clears the tmp dir (
  synchronously, in-handler). Every hot read hits the DB.
  *Fix:* small in-process TTL cache for token→runner, userID→user, session validity; run `ClearTmpDir` in a goroutine.
- **`getAllEvents` unbounded (Finding 10-API, Low).** `api/events.go:30-49` passes `Count:0` ⇒ no LIMIT ⇒ entire events
  table for the project/user.
  *Fix:* default a sane limit + pagination; index `event(project_id, id desc)`.

*Verified-OK:* SQL `PermanentConnection()` is true so `StoreMiddleware` is a no-op per request (`SqlDb.go:563`);
runner-tag loading is a single `IN` batch (`runner_tag.go:13-52`); output write batching is efficient.

### 5. Large inventory — confirmation of what is *not* a problem

Static inventory: one `os.WriteFile` per run (`LocalJob_inventory.go:92-98`). SSH/become keys: one per inventory via one
in-process agent (`LocalJob_inventory.go:15-28`, `pkg/ssh/agent.go:170-212`). Vars/secrets: built once as process env +
one `--extra-vars` JSON, independent of host count (`LocalJob.go:451-464`). No per-host loop, no quadratic string
building, no per-host syscall anywhere in Go. The only inventory-scaling cost is the blob being carried by value through
4–5 hops (cheap until marshaled) and the runner-poll re-serialization of Finding 7.

---

## Cross-cutting root causes

1. **Polling instead of events.** Both `RemoteJob.Run` (server-side, per task) and the runner job pool (client-side)
   poll every 1s; the UI polls every 10s. Each poll re-does full work (scans, full payload build, heartbeat write). The
   runner already PUTs progress — status transitions and job hand-off should be event/long-poll driven, not re-derived
   every second.
2. **O(n) scans/copies of in-memory state for the most frequent operations.** Status lookup, scheduling, runner load,
   and the admin list all full-scan or full-copy the queue/running collections under one lock. These need an `id`-keyed
   map and per-runner counters.
3. **The cheap global cap was removed without adding a real one.** `MaxParallelTasks=9999` + unbounded DB pool means "
   many tasks" now translates directly into goroutine count, DB connections, and contention with no backstop.
4. **Missing indexes on the columns the hot queries actually use** (`runner.token`, `task.status`, `task.created`,
   `task__output.stage_id`).
5. **The output path does per-line work that should be per-batch** (per-user marshal, per-record `StoreSession`) and
   persists with a constraint (`unique(task_id,time)`) that high-volume output structurally violates.

---

## Prioritized remediation roadmap

### Phase 0 — Cheap, high-leverage (migrations + a few lines)

1. **Add indexes:** `runner(token)` (unique), `task(status)` (or `task(project_id,status)`),
   `task(template_id,created)`, `task(project_id,created)`, `task(template_id,id)`, `task__output(task_id,stage_id)`,
   `task__output(task_id,time,id)`. *(Findings 2, 11, 13, 18, 9)*
2. **Bound the MySQL/Postgres pool** at `db/sql/SqlDb.go:82` (`SetMaxOpenConns/Idle/Lifetime`, configurable). *(Finding
   8)*
3. **Drop the `unique(task_id, time)` constraint**, order output by `id`, and fall back to per-row insert on batch
   failure. *(Finding 4)*
4. **Hoist `db.StoreSession` out of the `writeLogs` loop** (one session per flush). *(Finding 3)*
5. **Debounce `TouchRunner` and `TouchSession`** via in-memory last-touch maps. *(Finding 12)*
6. **Delete the `fmt.Println`** in the websocket handler. *(Finding 19)*

### Phase 1 — Output & streaming (the 5000-host local path)

7. **Marshal each log line once**, deliver async + lossy off the reader goroutine, skip fan-out when no subscriber;
   index hub connections by user. *(Findings 1, 14-hub)*
8. **Paginate `GetTaskOutput`** (keyset) or route the UI to the streaming raw endpoint. *(Finding 9)*
9. **Cap the runner's in-memory `logRecords`** and chunk `sendProgress`. *(Finding 15)*
10. **Grow the scanner buffer lazily; truncate over-long lines** instead of aborting. *(Finding 17)*

### Phase 2 — Scheduling & polling (the many-tasks path)

11. **O(1) `GetByID`** in the state store (back the running set with the existing id map); add a `map[runnerID]int` load
    counter. *(Findings 6, runner-select)*
12. **Replace the 1s `RemoteJob.Run` poll** with an event/condition signaled by the runner progress `PUT`. *(Finding 6)*
13. **Cache project/template parallel-task limits**; index "ready" tasks so `handleQueue` pops in O(1); buffer/coalesce
    `queueEvents`. *(Finding 5)*
14. **Reconsider unlimited default parallelism** — use a bounded worker pool with a sane default; drop the per-task
    `time.Sleep(1s)`. *(Finding 10)*
15. **Long-poll / push the runner job hand-off**; build+encrypt `JobData` once per job; ship inventory by id+hash, not
    inline per poll. *(Finding 7)*

### Phase 3 — DB hygiene & dashboards

16. **Batch the N+1s** (`getTemplates` env/vault, `getTasks.Fill`, `FillEvents`) with `IN (...)`. *(Findings 13,
    getTasks, FillEvents)*
17. **Move task retention off the write path** (`clearTasks`), delete by `id` cutoff. *(clearTasks)*
18. **Cache dashboard stats** (short TTL) and default a bounded date window. *(Finding 18)*
19. **Fan alerts out to a bounded background worker** with HTTP timeouts. *(Finding 16)*

---

## Suggested measurements (to confirm before/after)

- `pprof` CPU + goroutine profile of the server while running ~200 concurrent tasks emitting heavy output; expect
  `sendToWs`/`json.Marshal`, `handleQueue`/`GetProject`, and `RemoteJob.Run`/`GetTask` to dominate.
- DB: enable slow-query log; expect `SELECT * FROM runner WHERE token=?`, the `GetTaskStats` GROUP BY, and `clearTasks`
  `ORDER BY created` to appear. Watch active connection count vs. `max_connections`.
- A single 5000-host task on a **remote runner**: capture the `GetRunner` response size and per-poll CPU; expect
  multi-MB JSON + RSA cost repeated per poll until pickup.
- Memory: heap profile during many concurrent tasks; expect the 10 MB scanner buffers and 100k channels (Finding 17) and
  uncapped runner `logRecords` (Finding 15) to show.

---

## Appendix — verification status

Re-read and confirmed against source by the primary researcher (not just sub-agent report): Findings 1 (
`TaskRunner_logging.go:27-54`), 2 (`runner.token` indexes across all migrations + `runners.go:37`), 3 (
`TaskPool.go:300-352`), 4 (`SqlDb.go:93`), 7 (`runners.go:138-262`), 8 (`SqlDb.go:82-84`), 10 (`config.go:150,390` +
`TaskPool.go:367-370`). The remaining findings come from focused sub-agent read-throughs with `file:line` citations and
were cross-checked where they overlapped (the unbounded DB pool, the 1s poll loops, and the inventory-in-poll
serialization were each independently reported by multiple agents). Line numbers reflect the working tree on branch
`develop` as of 2026-06-04; a few may drift by a line or two after edits.
