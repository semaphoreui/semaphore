# Performance Research: Many Concurrent Tasks & Large Inventory (5000 hosts)

> Deep research into the Semaphore Go backend to locate performance and scalability
> bottlenecks under two stress scenarios:
> 1. **Many tasks** — hundreds of queued/running tasks at once.
> 2. **Large inventory** — a single static inventory with ~5000 hosts.
>
> Date: 2026-06-04 · Branch: `develop`
> Method: full-file reads of the task pool, output pipeline, SQL store, API/runner
> protocol and inventory path, with every cited `file:line` and SQL string verified
> against source. Severity reflects impact at the two scenarios above, not general code quality.

---

## 1. Executive summary

The backend is architecturally sound in outline (batched output inserts, a buffered
persist channel, single-key SSH agent, opaque inventory blob). The scaling problems
come from a small number of recurring patterns rather than isolated bugs:

| Theme                                                                | What it is                                                                                                                                                                         | Where it hurts                                                          |
|----------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| **A. 1-second polling everywhere**                                   | Runner→server poll, `RemoteJob` status poll, and UI all poll every 1–10 s instead of pushing events.                                                                               | Cost grows as `O(runners × running_tasks)` and `O(clients)` per second. |
| **B. O(n) scans/copies through one mutex + one scheduler goroutine** | The in-memory task state store answers every lookup by copying the whole queue/running collection under a single `RWMutex`; the scheduler re-scans the whole queue on every event. | Degrades as `O(tasks²)` with many tasks.                                |
| **C. Missing DB indexes on hot columns**                             | `runner.token`, `task.status`, `task.created`, `task__output.stage_id` are unindexed but filtered/sorted on hot paths.                                                             | Full table scans grow with history & runner count.                      |
| **D. Per-poll / per-request writes**                                 | `TouchRunner` and `TouchSession` issue an `UPDATE` on every poll/request.                                                                                                          | Serializes writers (catastrophic on SQLite's single write lock).        |
| **E. Unbounded reads & payloads**                                    | `GetTaskOutput`, `GetTaskStats`, `/events`, and the runner job payload (full inventory) are returned without limit/cache.                                                          | Memory + CPU spikes that grow with output volume / inventory size.      |
| **F. Synchronous work on hot paths**                                 | Per-line × per-user websocket marshal, per-line `StoreSession`, timeout-less alerts.                                                                                               | Throttles the subprocess reader and serializes the whole instance.      |
| **G. No DB pool limits + "unlimited parallel tasks" default**        | MySQL/Postgres pool is never bounded; default `MaxParallelTasks` is now `9999`.                                                                                                    | Unbounded goroutines & DB connections under load.                       |

**The single highest-leverage theme is G acting as an amplifier.** Commit `42e6c00d`
("unlimited number of parallel tasks by default") removed the only cheap global cap on
concurrency, while the MySQL/Postgres connection pool remains untuned. Every other
finding below gets multiplied by the number of simultaneously-running tasks, which is
now unbounded by default.

### Top 10 fixes by leverage

| #  | Fix                                                                                                                       | Severity | Effort  |
|----|---------------------------------------------------------------------------------------------------------------------------|----------|---------|
| 1  | Add unique index on `runner(token)` (+ cache token→runner)                                                                | Critical | Trivial |
| 2  | Marshal websocket log frame **once per line**, deliver async/lossy, skip when no subscriber                               | Critical | Medium  |
| 3  | Bound MySQL/Postgres connection pool (`SetMaxOpenConns/Idle/Lifetime`)                                                    | High     | Trivial |
| 4  | Add indexes: `task(status)`, `task(template_id, created)`, `task(project_id, created)`, `task__output(task_id, stage_id)` | High     | Trivial |
| 5  | Hoist `StoreSession` out of the per-record `writeLogs` loop                                                               | High     | Trivial |
| 6  | Debounce `TouchRunner` & `TouchSession` (write only if stale > 60 s)                                                      | High     | Easy    |
| 7  | Drop `unique(task_id, time)`; order output by `id`; per-row fallback on batch failure                                     | High     | Easy    |
| 8  | Paginate `GetTaskOutput` (the JSON endpoint loads the whole log)                                                          | High     | Easy    |
| 9  | Replace 1 s `RemoteJob`/runner polls with push/long-poll; stop shipping full inventory each poll                          | High     | Larger  |
| 10 | O(1) `GetByID` + per-runner counters in the task state store; cache parallel-limits out of `blocks()`                     | High     | Medium  |

---

## 2. Scenario mapping — which findings bite when

**Scenario 1 — many concurrent tasks (hundreds):**
TP-1, TP-2, TP-3, TP-4, TP-5, TP-6, TP-7 (task pool); OUT-1, OUT-2, OUT-3 (output);
DB-3, DB-8 (indexes/pool); API-1, API-2, API-3, API-4, API-5, API-7 (polling/writes).

**Scenario 2 — 5000-host inventory:**
INV-1 (runner re-serializes full inventory each poll — the dominant inventory cost);
OUT-1, OUT-3, OUT-4, OUT-5, OUT-6, OUT-7 (a 5000-host run emits 100k+ output lines);
DB-5 (unbounded output read); INV-3 (per-run materialization).
Notably, the inventory itself is **not** parsed per-host in Go (see §7), so the
classic `O(hosts²)` traps are absent — the cost is volume of *output*, not host iteration.

---

## 3. Task Pool & Concurrency

Files: `services/tasks/TaskPool.go`, `TaskRunner.go`, `task_state_store.go`,
`RemoteJob.go`, `TaskRunner_logging.go`, `alert.go`, `util/config.go`.

### TP-1 — Every queue event triggers a full O(n) re-scan, serialized through one goroutine — **High**

`TaskPool.handleQueue` is the single consumer of an **unbounded-but-unbuffered**
`queueEvents` channel. On every event (new task, finished task, requeue, and a 5 s tick)
it walks the entire waiting queue from index 0:

```go
// services/tasks/TaskPool.go:206-268
func (p *TaskPool) handleQueue() {
    for t := range p.queueEvents {
        var i = 0
        for i < p.state.QueueLen() {
            curr := p.state.QueueGet(i)
            if p.blocks(curr) { i++; continue }   // TP-2: DB call inside this loop
            if !p.state.TryClaim(curr.Task.ID) { ... }
            _ = p.state.DequeueAt(i)              // TP-5: O(n) slice splice
            runTask(curr, p)
        }
    }
}
```

**Mechanism:** With N queued tasks, a burst of N completions produces ~N events ×
O(N) scans = **O(N²)** passes, all funneled through one goroutine. Producers
(`AddTask` at `TaskPool.go:854`, finish/requeue sends at `TaskRunner.go:175,189`)
block on the unbuffered channel until the scheduler is free, so scheduling becomes a
global serialization point under load.

**Fix:** Maintain a per-project/template index of *ready* tasks so the loop pops in
O(1); buffer/coalesce `queueEvents`; break the inner loop once `MaxParallelTasks` is hit.

### TP-2 — `blocks()` does a synchronous `GetProject` DB read inside the scheduler hot loop — **High**

```go
// services/tasks/TaskPool.go:463-496
func (p *TaskPool) blocks(t *TaskRunner) bool {
    if util.Config.MaxParallelTasks > 0 && p.state.RunningCount() >= util.Config.MaxParallelTasks {
        return true                                   // with default 9999 this never fires
    }
    if p.state.ActiveCount(t.Task.ProjectID) == 0 { return false }
    for _, r := range p.state.GetActive(t.Task.ProjectID) { ... }
    proj, err := p.store.GetProject(t.Task.ProjectID) // <-- DB round-trip per candidate, per pass
    ...
}
```

**Mechanism:** Called for each queued candidate inside TP-1's inner loop. Once a
project has any active task, every scheduling pass does one `GetProject` DB round-trip
per queued candidate of that project. Combined with TP-1's O(N²), a few hundred queued
tasks in one project produce hundreds–thousands of `GetProject` queries/sec on the
single scheduler goroutine — DB latency now directly throttles task throughput. The
"unlimited" default (TP-7) makes the cheap early-out at line 465 dead code, so the
expensive path is always taken.

**Fix:** Cache `Project.MaxParallelTasks` / `Template.AllowParallelTasks` in memory,
refresh on update. Reorder so the in-memory `ActiveCount >= cachedMax` check happens
before any DB access.

### TP-3 — `RemoteJob.Run` busy-polls `GetTask` every second;

`GetTask` is an O(n) scan that copies the whole collection — **High** (dominant cost for remote/HA at scale)

```go
// services/tasks/RemoteJob.go:209-251 — one goroutine PER running remote task
for {
    time.Sleep(1_000_000_000)                  // 1s
    tsk, err = t.taskPool.GetTask(t.Task.ID)   // O(queue + running), see below
    if util.HAEnabled() {
        db.StoreSession(..., func() {
            row, rowErr = t.taskPool.store.GetTask(tsk.Task.ProjectID, t.Task.ID) // DB read every 1s/task
        })
    }
}
// services/tasks/TaskPool.go:142-166 — GetTask scans both collections
for _, t := range p.state.QueueRange()  { if t.Task.ID == id { ... } }  // QueueRange = full copy
for _, t := range p.state.RunningRange(){ if t.Task.ID == id { ... } }  // RunningRange = full copy
```

`QueueRange()`/`RunningRange()` each allocate a **full copy** of the collection under
the lock (`task_state_store.go:165-171, 202-210`) before the linear search.

**Mechanism:** With R running remote tasks there are R goroutines each calling
`GetTask` every second — **O(R × (Q+R)) per second** plus 2 slice-copy allocations per
call. In HA, add one `GetTask` *DB* read per running task per second. At 300 concurrent
remote tasks: ~300 DB reads/sec and ~300×600 pointer copies/sec just to poll status.

**Fix:** Add O(1) `GetByID` backed by the existing `running map[int]` (`task_state_store.go:124`).
Replace the 1 s poll with event-driven updates — the runner already PUTs progress to the
API (`api/runners/runners.go:326`); signal the waiting goroutine via channel/cond instead.

### TP-4 — `GetNumberOfRunningTasksOfRunner` is an O(running) scan called per candidate runner during selection — *

*Medium**

```go
// services/tasks/TaskPool.go:129-136
func (p *TaskPool) GetNumberOfRunningTasksOfRunner(runnerID int) (res int) {
    for _, task := range p.state.RunningRange() {   // full copy + scan, per candidate
        if task.Task.RunnerID != nil && *task.Task.RunnerID == runnerID { res++ }
    }
}
```

Called inside a double loop over runners in `RemoteJob.Run` (`RemoteJob.go:164-175`).
Starting one task is `O(passes × runners × running)`. A scheduler burst with many
runners and many running tasks makes this multiplicative cost noticeable.

**Fix:** Keep a `map[runnerID]int` counter updated in `onTaskRun`/`onTaskStop`; runner load becomes O(1).

### TP-5 — Queue is a slice with O(n) splice deletes → O(n²) drain — **Medium**

```go
// services/tasks/task_state_store.go:154-163
func (s *MemoryTaskStateStore) DequeueAt(index int) error {
    s.queue = append(s.queue[:index], s.queue[index+1:]...)  // shifts the whole tail
}
```

Each `DequeueAt` is O(n); draining N ready tasks is O(N²). `StopTasksByTemplate`
(`TaskPool.go:583-612`) removes matches one-by-one → also O(N²).

**Fix:** Use an O(1)-removal structure (linked list, or map+ordered index, or
tombstone-and-compact once per pass).

### TP-6 — A single `RWMutex` guards all in-memory task state; hot ops copy whole collections under it — **Medium**

```go
// services/tasks/task_state_store.go:121-127
type MemoryTaskStateStore struct {
    mu         sync.RWMutex
    queue      []*TaskRunner
    running    map[int]*TaskRunner
    activeProj map[int]map[int]*TaskRunner
    aliases    map[string]*TaskRunner
}
```

No I/O is held under the lock (good), but it is *one* lock shared by the scheduler
(TP-1), every per-task 1 s poll (TP-3), runner selection (TP-4), the cluster dashboard
`Snapshot()` (`task_state_store.go:286-318`, copies the entire state under `RLock`), and
status saves. The `RunningRange()`/`QueueRange()` full-copy calls hold the lock for an
O(n) copy, colliding with high-frequency pollers → lock convoy even though no single
critical section is "slow."

**Fix:** Eliminate the high-frequency full-copy callers (TP-3, TP-4); split into
finer-grained locks or `sync.Map` for the running set; expose count/lookup methods that don't allocate.

### TP-7 — Default `MaxParallelTasks` is `9999` (effectively unlimited): unbounded goroutines + per-task

`Sleep`, no admission control — **High** (amplifier for everything)

```go
// util/config.go:150  (runner) and :390 (server)
MaxParallelTasks int `json:"..." default:"9999" ...`

// services/tasks/TaskPool.go:354-371 — every admitted task parks a goroutine for 1s first
go func() { time.Sleep(1 * time.Second); task.run() }()
```

**Mechanism:** With the old default of 10, `blocks()` capped concurrency. Now the global
gate never fires. Consequences:

- **Goroutines:** ~1 per running task, plus per-task log-pipe goroutines
  (`TaskRunner_logging.go:64-65,158`) each with a 100k-buffered channel (OUT-7) → memory
  grows fast at hundreds of tasks.
- **DB connections:** the MySQL/Postgres pool is **never bounded** (DB-8) — only SQLite
  sets `SetMaxOpenConns(1)` (`SqlDb.go:83`). Unlimited parallel tasks each doing frequent
  `StoreSession`/`UpdateTask`/`GetTask` can exhaust the database's `max_connections`.
- **Thundering herd:** the unconditional `time.Sleep(1s)` parks a burst of admitted
  goroutines for a second, which then all hit `GetTask`/runner-selection at once.

**Fix:** Pair the "unlimited" default with real admission control (a bounded worker pool /
semaphore at a sane default), bound the DB pool (DB-8), and drop the unconditional `Sleep`.

### TP-8 — Alerts run synchronously in the status hot path with no HTTP timeout — **Medium**

`SetStatus` (`TaskRunner_logging.go:114-131`) calls `sendMailAlert` + 6 chat-webhook
senders inline on every notifiable transition. Each `send*Alert` does a blocking
`http.Post` with the default (timeout-less) client (`alert.go:171-487`); `sendMailAlert`
loops users doing blocking SMTP and re-fetches each user from the DB (`alert.go:81`).
These run on the task's own goroutine, so a slow alert endpoint pins that goroutine
indefinitely, delaying the `EventTypeFinished` send (`TaskRunner.go:189`) — keeping the
task in the running set longer and inflating every O(running) scan above.

**Fix:** Fan alerts out to a bounded background worker pool; set explicit `http.Client`
timeouts; reuse the already-loaded user list instead of per-alert DB reads.

### Positive findings (task pool)

- **Log batching is well designed:** `handleLogs` flushes by size (500) or every 500 ms
  with a 10000-buffered `logger` channel (`TaskPool.go:271-298, 82`). DB writes are off the per-line path.
- **In-memory locks are never held across I/O** — the critical sections are pure map/slice work.
- **HA distributed claim** (`TryClaim`/`DeleteClaim`) is correctly placed right before dequeue.
- **Schedule pool** (`robfig/cron`) runs each fired job in its own goroutine; its `locker`
  is held only during `Refresh()`, not during task execution.

---

## 4. Task Output Capture & Streaming

Files: `services/tasks/TaskRunner_logging.go`, `TaskPool.go`, `api/sockets/pool.go`,
`api/sockets/handler.go`, `db/sql/task.go`, `db/sql/SqlDb.go`, `api/projects/tasks.go`,
`services/runners/running_job.go`, `job_pool.go`, `api/runners/runners.go`.

> A 5000-host Ansible run can emit 100k+ stdout lines. Every per-line cost below is
> multiplied by that volume, then again by the number of concurrent tasks.

### OUT-1 — Synchronous websocket fan-out on the subprocess reader path, marshaled per line × per user — **Critical**

```go
// services/tasks/TaskRunner_logging.go:27-54
func (t *TaskRunner) LogWithTime(now time.Time, msg string) {
    t.sendToWs(now, msg)                       // synchronous, BEFORE persistence
    t.pool.logger <- logRecord{ task: t, output: msg, time: now }
    ...
}
func (t *TaskRunner) sendToWs(now time.Time, msg string) {
    for _, user := range t.users {             // every project user PLUS every admin
        b, err := json.Marshal(&map[string]any{ "type":"log", "output":msg, ... }) // per user!
        util.LogPanic(err)
        sockets.Message(user, b)               // pushes to unbuffered hub channel
    }
}
```

`t.users` includes every project user **plus every admin** (`TaskRunner.go:377-395`).
This runs on the same goroutine that drains the subprocess pipe
(`TaskRunner_logging.go:158-164`). So per line: one `json.Marshal` **per user** and one
push onto the **unbuffered** `h.broadcast` channel (OUT-2) per user — all before the
line is even queued for the DB. Cost ≈ `O(lines × users × connections)` through one hub
goroutine. Because `h.broadcast` is unbuffered, hub slowness backpressures the log
consumer → `linesCh` → `bufio.Scanner` → stops draining the OS pipe → **throttles
Ansible itself**. There is no "is anyone watching this task?" check.

**Fix:** Marshal the frame **once per line**; let the hub filter by user. Make delivery
async + lossy so it never blocks the reader. Skip fan-out entirely when no subscriber is
connected for that task/project.

### OUT-2 — `h.broadcast` is unbuffered; the whole instance serializes through one hub goroutine — **High**

```go
// api/sockets/pool.go:48-87
var h = hub{ broadcast: make(chan *sendRequest), ... }   // UNBUFFERED
func (h *hub) run() {
    for { select { case m := <-h.broadcast:
        for conn := range h.connections {                // O(all connections) per message
            select { case conn.send <- m.msg: default: /* drop slow client — good */ }
        }
    }}
}
```

The per-connection `default` drop correctly stops one slow client from blocking the hub.
But the unbuffered `broadcast` channel means every `sockets.Message` blocks its caller
until the single hub goroutine picks it up, and the hub re-scans **all** connections per
message even when it targets one user.

**Fix:** Buffer `broadcast` (lossy/coalescing when full); index connections by `userID`
(`map[int][]*connection`) so a targeted message is O(connections-for-that-user).

### OUT-3 — Per-record `StoreSession` inside the "batch" writer defeats batching — **High**

```go
// services/tasks/TaskPool.go:300-352
func (p *TaskPool) writeLogs(logs []logRecord) {
    taskOutput := make([]db.TaskOutput, 0)
    for _, record := range logs {                        // up to 500 records
        db.StoreSession(p.store, "logger", func() {      // <-- PER RECORD
            newStage, newState, err := stage_parsers.MoveToNextStage(...)
            ...
        })
        taskOutput = append(taskOutput, newOutput)
    }
    db.StoreSession(p.store, "logger", func() {
        err := p.store.InsertTaskOutputBatch(taskOutput) // the one good batched insert
    })
}
```

`StoreSession` (`db/Store.go:790-800`) wraps `Connect()`/`Close()` around the callback
**unless** `PermanentConnection()` is true. So if the store is not a permanent
connection, this opens/closes a DB connection **per output line** — 100k connect/close
cycles per task on the single `handleLogs` goroutine that serves *all* tasks. Even with a
permanent connection you pay 500 closure allocations + 500 calls per flush. (In this OSS
build `MoveToNextStage` is a no-op stub at `pro/pkg/stage_parsers/next_step.go:8-20`, so
the body does nothing useful yet still pays the wrapper cost; the Pro build does real
per-line DB work.) `handleLogs` is a single goroutine shared by all tasks, so any
per-record DB work here is a global serialization point.

**Fix:** Open one session, process all records, batch-insert, close once. Only invoke
stage parsing when the app uses stages; batch its DB effects. Consider sharding
`writeLogs` per task so one chatty task can't starve others.

### OUT-4 — `unique(task_id, time)` silently drops whole 500-line batches when two lines share a timestamp — **High

** (correctness + perf)

```go
// db/sql/SqlDb.go:93
d.sql.AddTableWithName(db.TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
```

`time` is captured per line at sub-second granularity. At 5000 hosts many lines share
the same `time`. `InsertTaskOutputBatch` (`task.go:244-264`) is one multi-row `INSERT`,
so a batch containing two equal timestamps **fails entirely**; the error is only logged
(`TaskPool.go:347-350`) and the whole 500-line batch is **silently dropped** — data loss

+ perf cliff (lost work + log spam). Even when timestamps differ, the unique index adds
  write amplification on a table taking 100k+ inserts/task.

**Fix:** Drop the `(task_id, time)` unique constraint (rows are already unique by
autoincrement `id`); order output by `id`; on batch failure fall back to per-row insert.

### OUT-5 — `GetTaskOutput` JSON endpoint loads the entire output unpaginated — **High**

```go
// api/projects/tasks.go:237
output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{})
helpers.WriteJSON(w, http.StatusOK, output)
```

`RetrieveQueryParams{}` ⇒ `Count == 0` ⇒ **no LIMIT** (`task.go:386-388`): every row for
the task (100k+ at 5000 hosts) is loaded and marshaled into one in-memory JSON array,
repeated per viewer. The raw-text endpoint (`tasks.go:262-268`) *does* paginate
(`chunkSize=10000`) — only the primary JSON endpoint the UI uses does not.

**Fix:** Paginate `GetTaskOutput` (server-side max `Count` + cursor), or point the UI at
the streaming raw endpoint. Add a composite `task__output(task_id, time, id)` index and
prefer keyset pagination (`WHERE task_id=? AND id > ?`) over `OFFSET` (which is O(offset)).

### OUT-6 — Remote runner buffers the entire task output in an uncapped in-memory slice — **High**

On the runner, every line is appended to `p.logRecords` (`running_job.go:46-57`) with no
cap. `sendProgress` runs on a 1 s ticker (`job_pool.go:179`), PUTs the whole slice
(`job_pool.go:288-298`), and only trims after a successful round-trip (`:333-338`). A
fast 5000-host run accumulates tens of thousands of lines per tick into one request body;
if the server is briefly slow the slice grows unbounded → runner memory blowup. On the
server, `runnerSetTaskProgress` (`runners.go:316-318`) replays every received record
through `LogWithTime`, re-triggering OUT-1's synchronous fan-out in bursts of thousands.

**Fix:** Cap `logRecords` (ring buffer / max bytes), chunk `sendProgress` to a bounded
size, send more often when the backlog is large; on the receiver, feed records into the
async coalesced broadcast path instead of per-line synchronous fan-out.

### OUT-7 — 10 MB scanner buffer per pipe per task + per-line string copy + 100k-deep channel — **Medium**

```go
// services/tasks/TaskRunner_logging.go:153-201
linesCh := make(chan string, 100000)           // ~1.6 MB of slots per pipe, per task
scanner := bufio.NewScanner(reader)
const maxCapacity = 10 * 1024 * 1024           // 10 MB allocated up-front per pipe
buf := make([]byte, maxCapacity); scanner.Buffer(buf, maxCapacity)
for scanner.Scan() { linesCh <- scanner.Text() } // Text() allocates a new string per line
```

Each task allocates **20 MB** of scanner buffers (stdout+stderr) + two 100k-deep
channels up front, regardless of output. With unlimited parallel tasks (TP-7) that is
tens of MB × N tasks of resident memory. `Text()` also allocates a fresh string per line
(100k allocations/task of GC churn). A single line > 10 MB triggers
`"bufio.Scanner: token too long"` which **kills the task** (`:188-192`). The huge
`linesCh` buffer mostly masks the structurally-slow downstream (OUT-1/OUT-3).

**Fix:** Start the scanner buffer small and let it grow
(`scanner.Buffer(make([]byte, 64*1024), maxCapacity)`); truncate over-long lines instead
of killing the task; after fixing OUT-1/OUT-3, shrink `linesCh` to a few thousand.

### OUT-8 — Stray `fmt.Println` on the websocket read path — **Low**

```go
// api/sockets/handler.go:91-92
_, message, err := c.ws.ReadMessage()
fmt.Println(string(message))   // unconditional stdout write per inbound ws frame
```

Leftover debug code: writes to `os.Stdout` (serialized by a global lock) for every
inbound frame from every client — exactly the high-fan-out scenario.

**Fix:** Delete it or gate behind debug-level structured logging.

---

## 5. Database Query Patterns & Indexes

Files: `db/sql/SqlDb.go`, `task.go`, `template.go`, `event.go`, `schedule.go`,
`db/Task.go`, `db/Event.go`, and `db/sql/migrations/`.

### Index inventory — EXISTS vs MISSING on hot tables

| Table          | Indexed (EXISTS)                                                                                                                | **MISSING** on hot columns                                                                                             |
|----------------|---------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `task`         | `template_id`, `project_id`, `integration_id`, `inventory_id`, `schedule_id` (`v2.15.1.sqlite.sql:370-383`, `v2.17.15.sql:7-8`) | **`status`, `created`, `user_id`, composite `(template_id, id)` / `(template_id, created)` / `(project_id, created)`** |
| `task__output` | `task_id` (`v2.15.1.sqlite.sql:413`), `time` (`v2.14.12.sql:1`)                                                                 | **`stage_id`**, composite `(task_id, time, id)`                                                                        |
| `event`        | `project_id`, `user_id` (`v2.15.1.sqlite.sql:116-120`)                                                                          | composite `(project_id, id desc)`                                                                                      |
| `runner`       | `project_id`, `registration_token` (`v2.18.7.sql:3`)                                                                            | **`token`** (see API-1 — used on every poll)                                                                           |

Good news: the feared "`task__output` with no `task_id` index" is **not** present — that index exists.

### DB-1 — `getTemplates`: correlated subquery per row + N+1 for environments/vaults — **High**

```go
// db/sql/template.go:283 — correlated subquery, runs once per template row
"(SELECT `id` FROM `task` WHERE template_id = pt.id ORDER BY `id` DESC LIMIT 1) last_task_id",
// db/sql/template.go:412 — GetTemplateEnvironments per template
// db/sql/template.go:406 — GetTemplateVaults per template (GetTemplates passes loadVaults=true at :433)
```

The templates page issues `1 + N (last_task) + N (envs) + N (vaults)` queries. 200
templates ⇒ 400–600 queries/page. **Fix:** batch envs/vaults with `WHERE template_id IN (...)`;
add `task(template_id, id)` so the `last_task_id` subquery is an index-only backward scan.

### DB-2 — `getTasks`: N+1 `Fill()` loop calling `GetTask` per build task — **Medium** (High for deploy projects)

```go
// db/sql/task.go:314-319
for i := range *tasks { err = (*tasks)[i].Fill(d) }
// db/Task.go:184-196 — Fill calls d.GetTask(...) for every task with BuildTaskID != nil
```

`GetProjectTasks`/`GetLastTasks` fetch up to 200–1000 rows; for deploy projects where
most tasks have a `build_task_id` that is up to ~1000 extra join queries per page.
**Fix:** collect non-nil `BuildTaskID`s and resolve in one `WHERE id IN (...)`.

### DB-3 — `task.status` and `task.created` unindexed but filtered/sorted/retained on — **High**

- `task.go:292-293` — list filter `WHERE status IN (...)` (no `status` index).
- `SqlDb.go:934-955` (`GetTaskStats`, dashboard) — `GROUP BY DATE(created), status` filtered on
  `created/status/start/end/user_id`, only `project_id` indexed ⇒ full per-project scan + filesort every load.
- `task.go:152-160` (`clearTasks`, runs on **every** task creation) —
  `SELECT created ... ORDER BY created DESC LIMIT 1 OFFSET ?` then
  `DELETE ... WHERE template_id=? AND created<?`, no `(template_id, created)` index, on the write path.

**Fix:** add `task(status)` (or `task(project_id, status)`), `task(template_id, created)`,
`task(project_id, created)`. The `(template_id, created)` index fixes both `clearTasks` queries.

### DB-4 — Retention (`clearTasks`) runs synchronously on the insert hot path — **Medium**

`CreateTask` (`task.go:186`) runs `clearTasks` inline; with a 10 % deadzone (`:147`) and a
random recount `SELECT count(*) FROM task WHERE template_id=?` (`:129`). Under bursty
concurrent creation each insert may trigger a count + an `ORDER BY created` scan + a range
delete, all unindexed, serializing writers. **Fix:** add the index (DB-3), move retention
to a background sweep, delete by monotonic `id` cutoff instead of `created`.

### DB-5 — `GetTaskOutput` loads the entire log into memory (same as OUT-5) — **High**

See OUT-5. The JSON endpoint (`api/projects/tasks.go:237`) calls `GetTaskOutputs` with
empty params ⇒ no LIMIT.

### DB-6 — `task__output.stage_id` unindexed but filtered directly — **Medium**

`stage_id` added by `v2.16.8.sql:1` with no index; `GetTaskStageOutputs` (`task.go:405-408`)
filters `WHERE task_id=? AND stage_id=?`. The `task_id` index narrows it, but on a
large-output task the residual `stage_id` filter scans all of that task's rows.
**Fix:** `CREATE INDEX task__output_task_stage_idx ON task__output(task_id, stage_id)`.

### DB-7 — `FillEvents` N+1 (per-event task lookups) — **Medium**

`FillEvents` (`db/Event.go:73-114`) loops events calling `GetTask` per task-type event
(usernames are memoized, tasks are not). The activity feed often returns hundreds of rows.
**Fix:** batch-resolve task object names with one `IN (...)`; memoize like usernames.

### DB-8 — MySQL/Postgres connection pool is never bounded — **High under concurrency**

```go
// db/sql/SqlDb.go:82-84 — ONLY SQLite is bounded
if d.GetDialect() == util.DbDriverSQLite { sqlDb.SetMaxOpenConns(1) }
```

No `SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime/SetConnMaxIdleTime` for
MySQL/Postgres anywhere. Go's default is unlimited open + 2 idle. With unlimited parallel
tasks (TP-7) each writing status/output/stage/event rows, the pool opens unbounded
connections then churns them back to 2 idle — can exhaust the DB server's
`max_connections` and add connection-setup latency. **Fix:** set sane, configurable pool
bounds where the SQLite branch is.

### DB-9 — Output write path: batching good; per-record session + per-record stage call (same as OUT-3) — **Low–Medium**

The batched `InsertTaskOutputBatch` is correct; the per-record `StoreSession` +
`MoveToNextStage` loop above it is the problem (see OUT-3).

### Positive findings (DB)

- `task__output(task_id)` **is** indexed — the catastrophic missing-index case is absent.
- `loadRunnerTags` uses a single `WHERE runner_id IN (...)` batch — no N+1 listing runners.
- For SQL backends `PermanentConnection()` is `true`, so `StoreMiddleware`/`StoreSession`
  is a no-op per request (no per-request connect overhead) — except inside `writeLogs` (OUT-3).

---

## 6. API Layer & Runner Polling

Files: `api/runners/runners.go`, `api/tasks/tasks.go`, `api/projects/tasks.go`,
`api/events.go`, `api/auth.go`, `api/cache.go`, `db/sql/global_runner.go`,
`db/sql/session.go`, `services/runners/job_pool.go`.

### API-1 — Runner poll auth does a full table scan on the unindexed `token` column, every poll, every runner — *

*Critical**

```go
// api/runners/runners.go:37 — on EVERY runner request
RunnerMiddleware → store.GetRunnerByToken(token)
// db/sql/global_runner.go:12-32 → SqlDb.go:491-501 produces:
SELECT * FROM runner pe WHERE token=?     // GlobalRunnerProps.IsGlobal ⇒ no project_id predicate
```

The `runner` table has indexes on `project_id` and `registration_token` — **not `token`**
(verified across all migrations). Bolt is worse: it iterates all runner objects comparing
`r.Token == token` (`db/bolt/global_runner.go:12-30`). Each runner polls ~every 1 s
(`job_pool.go:179`) issuing both a `PUT` (sendProgress) and a `GET` (checkNewJobs) — both
pass through `RunnerMiddleware`. N runners ⇒ ~2N full scans/sec of `runner` over an
unindexed column.

**Fix:** Add a unique index on `runner(token)` (and a token→id secondary index for Bolt);
cache token→runner in memory with a short TTL so repeated polls skip the DB.

### API-2 — Every runner poll writes to the DB (`TouchRunner` UPDATE) on the hot GET path — **High**

`GetRunner` (`runners.go:111-123`) unconditionally calls `TouchRunner` →
`UPDATE runner SET touched=? WHERE id=?` (`global_runner.go:138-154`). N runners ⇒ N
UPDATE writes/sec (after the API-1 scan). `touched` is only consumed at 30-min
granularity (`RemoteJob.go:23`). On SQLite the single write lock serializes these against
all task-status/output writes. **Fix:** debounce — write only if last touch > ~60 s ago
(in-memory per-runner last-seen).

### API-3 — Runner job-poll response serializes + RSA-encrypts the full job payload every poll — **High**

`GetRunner` ranges **all** running tasks (`GetRunningTasks()` copies the whole running map
under lock), filters client-side, and for each `TaskStartingStatus` job builds full
`JobData` — decrypting every secret (SSH key, become key, all vaults, repo keys,
`runners.go:159-226`) — then `json.Marshal`s and **chunked-RSA-encrypts** it
(`runners.go:236-262`, 245-byte blocks for a 2048-bit key). A starting job stays in the
pool until the runner picks it up, so this heavy decrypt+JSON+RSA work repeats every
second per starting job per polling runner ⇒ `O(runners × running_tasks)` crypto/sec.
**Fix:** index running set by runner id; mark a job "dispatched" after first hand-off and
stop rebuilding it; move to long-poll/push so the payload is built once per job.

### API-4 —

`RemoteJob.Run` polls the DB once/sec/task + re-fetches the full runner list per task start (same as TP-3/TP-4) — **High
**

See TP-3 (1 s `GetTask` DB poll per running task, doubled in HA) and TP-4 (re-fetch
`GetRunners` + `GetAllRunners` + per-candidate running-count scan at task start,
`RemoteJob.go:135-179`). 200 running remote tasks ⇒ 200–400 task queries/sec independent
of runner polling. **Fix:** drive transitions off the runner's progress PUT; cache the
runner list briefly; per-runner running-count map.

### API-5 — Admin "all tasks" endpoint: no pagination/cache, copies the whole pool, polled every 10 s — **Medium**

`GetTasks` (`api/tasks/tasks.go:43-71`) ranges all queued + all running tasks and
serializes them with no pagination/cache; the admin Tasks view polls it every 10 s
(`web/src/views/Tasks.vue:92-96`). `QueueRange()`/`RunningRange()` copy the whole
collection under the state lock per call, contending with the pool's hot path (TP-6).
**Fix:** paginate; serve from a rate-limited cached snapshot; push pool changes over the websocket.

### API-6 — Per-project task list fetches up to 1000 joined rows, no pagination, + conditional N+1 — **Medium**

`GetAllTasks` (`api/projects/tasks.go:100-102`) calls `GetTasksList(..., 1000)` — a
4-table join (`task.go:269-322`) with the per-row `Fill` N+1 (DB-2). No keyset/offset
pagination. **Fix:** real keyset pagination on `id desc`; batch the `BuildTaskID` lookups.

### API-7 — Authenticated requests write to the DB on every request (`TouchSession` UPDATE) — **Medium**

`authenticationHandler` (`api/auth.go:264-268`) runs `TouchSession` →
`UPDATE session SET last_active=? ...` on every cookie-auth request (preceded by
`GetSession` + `GetUser` reads). Every UI poll (task list 10 s, dashboards, stats,
history) is such a request, so this is a steady write stream that on SQLite contends with
task writes. `last_active` is only used to expire > 7-day sessions. **Fix:** debounce —
write only if older than ~5 min (in-memory per-session last-touch).

### API-8 — `GetTaskStats` aggregates the whole `task` table with no date floor and no cache — **Medium**

`GetTaskStats` (`api/projects/tasks.go:398-444` → `SqlDb.go`) does
`SELECT DATE(created), status, COUNT(*) ... GROUP BY date,status` with no lower date bound
when `filter.Start` is absent, no cache, and no `(project_id, created, status)` index ⇒
full scan + sort over the project's entire history per dashboard load. **Fix:** cache per
(project, template, range) with short TTL; default a bounded window; add a covering index.

### API-9 — No read-path cache anywhere; `ClearTmpDir` runs synchronously in the handler — **Low**

`api/cache.go` contains only `clearCache` → synchronous `ClearTmpDir` (`:12-32`). There is
**no** response/query cache layer — the root enabler of API-1/4/5/7/8. **Fix:** small
in-process TTL cache for token→runner, userID→user, session validity; run `ClearTmpDir`
in a goroutine.

### API-10 — `/events` "all events" route loads the full table unbounded — **Low**

`getAllEvents` (`api/events.go:30-49`) passes `Count: 0` ⇒ no LIMIT, returning every event
for the project/user; non-admins also incur an extra `GetProjectUser` read per call.
**Fix:** default limit + pagination; index `(project_id, id desc)`.

### Positive findings (API)

- For SQL backends `PermanentConnection()` is true ⇒ no per-request DB connect.
- API-token auth (`auth.go:225`) avoids the `TouchSession` write — the right model.
- Note: `api/sockets/pool.go:19,48` use module-level `var h hub` / `var broadcaster`,
  which technically conflicts with the project's "no global variables" rule
  (`.claude/CLAUDE.md`) — not a perf issue, but worth flagging.

---

## 7. Large Inventory Handling (5000 hosts)

Files: `db/Inventory.go`, `db/sql/inventory.go`, `services/server/inventory_svc.go`,
`services/tasks/LocalJob_inventory.go`, `LocalJob.go`, `db_lib/AnsiblePlaybook.go`,
`pkg/ssh/agent.go`, `api/runners/runners.go`.

> **Headline:** Semaphore treats the inventory as an **opaque blob**. It is never parsed,
> split per-host, or iterated in Go — there are no per-host loops, no quadratic string
> building, and no per-host key/file/syscall operations. The classic `O(hosts²)` traps you
> might expect **do not exist here.** The real costs are (a) carrying the full inventory
> text through copies, and (b) re-serializing it on every runner poll.

### INV-1 — Full inventory text re-marshaled into JSON on every runner poll — **High** (remote-runner path only)

```go
// api/runners/runners.go:138-157 — on each GetRunner poll
tasks := c.taskPool.GetRunningTasks()
for _, tsk := range tasks {
    if tsk.Task.Status == task_logger.TaskStartingStatus {
        data.NewJobs = append(data.NewJobs, runners.JobData{
            ...
            Inventory: tsk.Inventory,   // full multi-MB Inventory string, marshaled inline
        })
    }
}
```

`JobData.Inventory db.Inventory` (`services/runners/types.go:17`) carries the
`Inventory string` field (`db/Inventory.go:24`) with no `json:"-"`. A 5000-host static
inventory is ~0.5–2 MB of text. A task can sit in `TaskStartingStatus` across multiple
polls, so the same multi-MB inventory is JSON-escaped and written to the socket on **every
poll, by every polling runner, for every starting task** — `O(inventory_bytes × polls ×
runners)` of allocation + byte-wise JSON escaping (CPU-heavy) + network, none cached or
diffed. This is the dominant inventory-related cost and the one place to fix first.

**Fix:** Don't ship the body in the poll response — send `inventory_id` + content hash and
let the runner GET the body once via a cacheable endpoint (ETag/304); at minimum, include
the full `Inventory` only for jobs actually assigned this cycle (dedupe so a starting job
is serialized once, not every poll); gzip the runner response.

### INV-2 — Inventory string copied by value through 4–5 hops — **Medium**

`db.Inventory` is passed by value at every stage (`TaskRunner.go:404`, `TaskPool.go:842/433`,
`job_pool.go:689`, `runners.go:153`). Go strings share backing arrays, so each copy is a
16-byte header — *until* a mutation or marshal forces a full allocation (INV-1's JSON, INV-3's
disk write). Medium because the by-value design means any future `inventory.Inventory =
transform(...)` or log/marshal silently allocates the whole blob. **Fix:** pass
`*db.Inventory`; keep the body out of any DTO that gets logged/marshaled.

### INV-3 — Static inventory written to a fresh temp file on every run — **Low**

```go
// services/tasks/LocalJob_inventory.go:92-98
func (t *LocalJob) installStaticInventory() error {
    return os.WriteFile(t.tmpInventoryFullPath(), []byte(t.Inventory.Inventory), 0664)
}
```

A **single** `os.WriteFile` of the whole blob per task (one `[]byte(...)` copy), removed
once at the end — no per-host syscalls, no re-read, no re-parse. Only Low: the `[]byte`
conversion transiently duplicates the ~2 MB blob, and it is rewritten even when identical
to the previous run. **Fix (optional):** skip the rewrite when a stored content hash
matches (mirror the `requirements.md5` pattern in `db_lib/AnsibleApp.go:42-49`).

### INV-4 / INV-5 — Positive findings (no per-host multiplier)

- **SSH/become key install is per-inventory, not per-host** (`LocalJob_inventory.go:15-28`):
  at most one SSH key + one become key, installed into a **single** in-process SSH agent
  (`pkg/ssh/agent.go:170-212`) — one Unix socket, one goroutine, one key — that all 5000
  hosts connect through. No per-host file/exec/syscall. Per-host connection fan-out is
  Ansible's concern (`forks`), outside this codebase.
- **Secrets/vars injected globally, not per-host** (`LocalJob.go:451-464`,
  `getEnvironmentENV:157-179`): env + a single `--extra-vars` JSON argument built once from
  Environment/Secret/Survey data, independent of host count.

### Inventory cost ranking at 5000 hosts

| Path                                              | Dominant cost                                                   | Scales with                                     |
|---------------------------------------------------|-----------------------------------------------------------------|-------------------------------------------------|
| **Remote runner** (`UseRemoteRunner`/`RunnerTag`) | JSON-encode full inventory on every poll (INV-1)                | `inventory_bytes × polls × runners` — **worst** |
| **Local static**                                  | one `os.WriteFile` + one `[]byte` copy per run (INV-3)          | `inventory_bytes × tasks`                       |
| **Local file** (`InventoryFile`)                  | none — Ansible reads from the repo; Semaphore only joins a path | nothing in Go                                   |
| **Dynamic / Terraform**                           | n/a — state is a separate blob, not host-expanded in Go         | nothing per-host                                |

---

## 8. Prioritized remediation roadmap

### Phase 1 — Quick wins (indexes + small, isolated changes)

1. **Index `runner(token)`** unique (API-1) — eliminates a full scan on the busiest endpoint.
2. **Bound the MySQL/Postgres pool** at `SqlDb.go:82` (DB-8/TP-7) — `SetMaxOpenConns/Idle/Lifetime`, configurable.
3. **Add task/output indexes** (DB-3/4/6): `task(status)`, `task(template_id, created)`,
   `task(project_id, created)`, `task(template_id, id)`, `task__output(task_id, stage_id)`,
   `task__output(task_id, time, id)`.
4. **Hoist `StoreSession` out of the `writeLogs` loop** (OUT-3) — one session per flush, not per line.
5. **Marshal the websocket frame once per line** + skip when no subscriber (OUT-1).
6. **Debounce `TouchRunner` & `TouchSession`** (API-2/API-7) — write only if stale.
7. **Paginate `GetTaskOutput`** JSON endpoint (OUT-5/DB-5).
8. **Drop `unique(task_id, time)`**, order output by `id`, per-row fallback on batch failure (OUT-4).
9. **Remove `fmt.Println`** on the ws read path (OUT-8); **drop the per-task `time.Sleep(1s)`** (TP-7).

### Phase 2 — Medium refactors

10. **O(1) `GetByID` + per-runner running-count map** in the task state store (TP-3/TP-4); buffer `queueEvents`.
11. **Cache project/template parallel limits** out of `blocks()`'s hot loop (TP-2).
12. **Buffer the ws hub + index connections by userID** (OUT-2); make broadcast async/lossy.
13. **Batch the N+1s**: `getTemplates` envs/vaults (DB-1), `getTasks.Fill` (DB-2), `FillEvents` (DB-7).
14. **Cache `GetTaskStats`** + default a bounded date window (API-8); add the read-path TTL cache (API-9).
15. **Don't ship the inventory body on every runner poll** — hash + fetch-once (INV-1).
16. **Move retention (`clearTasks`) off the insert path** to a background sweep, delete by `id` (DB-4).
17. **Fan alerts out to a bounded worker pool** with HTTP timeouts (TP-8).

### Phase 3 — Larger architectural changes

18. **Replace 1-second polling with push / long-poll** across the runner protocol, `RemoteJob`
    status, and the UI (API-3/API-4/TP-3) — the single biggest structural win for "many tasks."
19. **Real admission control** for the unlimited-parallelism default — a bounded worker pool /
    semaphore sized sanely (TP-7), pairing the convenient default with backpressure.
20. **Restructure the task state store** for finer-grained locking and a ready-task index so the
    scheduler is O(1) per dispatch instead of O(n²) (TP-1/TP-5/TP-6).
21. **Stream task output** end-to-end (cap runner buffers, keyset pagination, drop slow ws clients) (OUT-6/OUT-7).

### What's already good (don't regress)

Log batching (`handleLogs`), the single batched `InsertTaskOutputBatch`, locks never held
across I/O, HA distributed claim placement, the single-key SSH agent, the opaque-blob
inventory design (no per-host Go iteration), and `task__output(task_id)` being indexed.

---

## 9. Finding index

| ID           | Title                                                             | Severity     | Primary location                                          |
|--------------|-------------------------------------------------------------------|--------------|-----------------------------------------------------------|
| API-1        | Runner token full table scan every poll                           | **Critical** | `api/runners/runners.go:37`, `db/sql/global_runner.go:12` |
| OUT-1        | Per-line × per-user synchronous ws marshal                        | **Critical** | `services/tasks/TaskRunner_logging.go:27-54`              |
| TP-1         | O(n²) queue re-scan through one goroutine                         | High         | `services/tasks/TaskPool.go:206-268`                      |
| TP-2         | `GetProject` DB call in scheduler hot loop                        | High         | `services/tasks/TaskPool.go:463-496`                      |
| TP-3         | 1 s busy-poll + O(n) `GetTask` per running task                   | High         | `RemoteJob.go:209-251`, `TaskPool.go:142-166`             |
| TP-7         | Unlimited parallel default + per-task Sleep, no admission control | High         | `util/config.go:150,390`, `TaskPool.go:368`               |
| OUT-2        | Unbuffered ws hub, single goroutine                               | High         | `api/sockets/pool.go:48-87`                               |
| OUT-3        | Per-record `StoreSession` defeats batching                        | High         | `services/tasks/TaskPool.go:300-352`                      |
| OUT-4        | `unique(task_id, time)` drops whole batches                       | High         | `db/sql/SqlDb.go:93`                                      |
| OUT-5 / DB-5 | `GetTaskOutput` loads entire log                                  | High         | `api/projects/tasks.go:237`                               |
| OUT-6        | Runner buffers entire output uncapped                             | High         | `services/runners/job_pool.go:288-338`                    |
| DB-1         | `getTemplates` subquery + env/vault N+1                           | High         | `db/sql/template.go:283,406,412`                          |
| DB-3         | `task.status`/`task.created` unindexed                            | High         | `db/sql/task.go:152,292`, `SqlDb.go:934`                  |
| DB-8         | MySQL/Postgres pool unbounded                                     | High         | `db/sql/SqlDb.go:82-84`                                   |
| API-2        | `TouchRunner` write every poll                                    | High         | `api/runners/runners.go:111`, `global_runner.go:138`      |
| API-3        | Full job payload RSA-encrypted every poll                         | High         | `api/runners/runners.go:138-262`                          |
| API-4        | `RemoteJob` DB poll/sec + runner re-fetch                         | High         | `services/tasks/RemoteJob.go:135-251`                     |
| INV-1        | Full inventory re-marshaled every runner poll                     | High         | `api/runners/runners.go:138-157`                          |
| TP-4         | O(running) scan per candidate runner                              | Medium       | `services/tasks/TaskPool.go:129-136`                      |
| TP-5         | O(n) splice dequeue → O(n²) drain                                 | Medium       | `services/tasks/task_state_store.go:154-163`              |
| TP-6         | Single RWMutex + full-collection copies                           | Medium       | `services/tasks/task_state_store.go:121-127`              |
| TP-8         | Synchronous timeout-less alerts in hot path                       | Medium       | `TaskRunner_logging.go:114-131`, `alert.go`               |
| OUT-7        | 10 MB scanner buffer + 100k channel per pipe                      | Medium       | `TaskRunner_logging.go:153-201`                           |
| DB-2         | `getTasks.Fill` N+1                                               | Medium       | `db/sql/task.go:314`, `db/Task.go:184`                    |
| DB-4         | Retention on the write path                                       | Medium       | `db/sql/task.go:121-168`                                  |
| DB-6         | `task__output.stage_id` unindexed                                 | Medium       | `db/sql/task.go:405-408`                                  |
| DB-7         | `FillEvents` per-event task N+1                                   | Medium       | `db/Event.go:73-114`                                      |
| API-5        | Admin all-tasks: no page/cache, 10 s poll                         | Medium       | `api/tasks/tasks.go:43-71`                                |
| API-6        | Per-project list: 1000 rows, no pagination                        | Medium       | `api/projects/tasks.go:100`, `task.go:269`                |
| API-7        | `TouchSession` write every request                                | Medium       | `api/auth.go:264`, `session.go:74`                        |
| API-8        | `GetTaskStats` full-history aggregation, no cache                 | Medium       | `api/projects/tasks.go:398`, `SqlDb.go:934`               |
| INV-2        | Inventory string copied by value × 5 hops                         | Medium       | `TaskRunner.go:404`, `runners.go:153`                     |
| OUT-8        | Stray `fmt.Println` on ws read path                               | Low          | `api/sockets/handler.go:91`                               |
| API-9        | No read-path cache; sync `ClearTmpDir`                            | Low          | `api/cache.go:12-32`                                      |
| API-10       | `/events` all-events unbounded                                    | Low          | `api/events.go:30-49`                                     |
| INV-3        | Static inventory rewritten every run                              | Low          | `services/tasks/LocalJob_inventory.go:92-98`              |
