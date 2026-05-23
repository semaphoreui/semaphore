# Plan — Improving the HA Cluster Implementation

- **Scope:** `pro_impl/services/ha/*`, `pro_impl/services/tasks/redis_task_state_store.go`,
  and their integration points (`cli/cmd/root.go`, `services/tasks/TaskPool.go`,
  `services/schedules/SchedulePool.go`, `api/sockets/pool.go`).
- **Branch:** `develop`
- **Status:** research + proposal (no code changes yet)

## 1. Current Architecture

Active-active HA is enabled by `SEMAPHORE_HA_ENABLED`. When on, every Semaphore
node shares one Redis instance and coordinates through five components, all
wired in `runService()` (`cli/cmd/root.go:80-185`):

| Component | File | Responsibility |
|-----------|------|----------------|
| `RedisTaskStateStore` | `pro_impl/services/tasks/redis_task_state_store.go` | Shared task queue / running / active / alias state + Pub/Sub sync |
| `RedisNodeRegistry` | `pro_impl/services/ha/node_registry.go` | Heartbeat-based cluster membership |
| `RedisOrphanCleaner` | `pro_impl/services/ha/orphan_cleaner.go` | Fails tasks whose owning node died |
| `RedisScheduleDeduplicator` | `pro_impl/services/ha/schedule_dedup.go` | Ensures one node fires each cron occurrence |
| `RedisWSBroadcaster` | `pro_impl/services/ha/ws_broadcaster.go` | Relays WebSocket events across nodes |

### 1.1 Complete Redis Key Inventory

The user asked to find *every* key the implementation touches. Below is the
full set, grouped by owning component.

**`RedisTaskStateStore` (prefix `tasks:`)**

| Key | Type | TTL | Written by | Cleaned by |
|-----|------|-----|------------|------------|
| `tasks:queue` | LIST | **none** | `Enqueue` / `DequeueAt` | `DequeueAt` only |
| `tasks:running` | SET | **none** | `SetRunning` | `DeleteRunning`, orphan cleaner |
| `tasks:active:<projectID>` | SET | **none** | `AddActive` | `RemoveActive`, orphan cleaner |
| `tasks:owners` | HASH `taskID→nodeID` | `HEXPIRE` 7d | `SetRunning` | `DeleteRunning`, orphan cleaner |
| `tasks:task_project` | HASH `taskID→projectID` | `HEXPIRE` 7d | `Enqueue` / `AddActive` | `DeleteRunning`, orphan cleaner |
| `tasks:aliases` | HASH `alias→taskID` | `HEXPIRE` 7d | `SetAlias` | `DeleteAlias`, orphan cleaner |
| `tasks:runtime:<taskID>` | HASH | `EXPIRE` 7d | `UpdateRuntimeFields` | `DeleteRunning`, orphan cleaner |
| `tasks:claim:<taskID>` | STRING `nodeID` | `EXPIRE` 5m | `TryClaim` | `DeleteClaim`, orphan cleaner |
| `tasks:events` | Pub/Sub channel | — | every mutating op | — |

**`RedisNodeRegistry` (prefix `ha:`)**

| Key | Type | TTL | Notes |
|-----|------|-----|-------|
| `ha:node:<nodeID>` | STRING `"1"` | `EXPIRE` 30s | Per-node heartbeat |
| `ha:nodes` | SET of node IDs | **none** | Membership; pruned by scan |

**`RedisScheduleDeduplicator`**

| Key | Type | TTL | Notes |
|-----|------|-----|-------|
| `ha:sched:lock:<scheduleID>:<minuteEpoch>` | STRING | `EXPIRE` 55s | One per schedule per minute |

**`RedisWSBroadcaster`**

| Key | Type | TTL | Notes |
|-----|------|-----|-------|
| `ha:ws:broadcast` | Pub/Sub channel | — | All cross-node WS traffic |

### 1.2 "Extra" / leaking keys found

These are keys that can persist beyond their useful life — i.e. real leaks:

1. **`tasks:queue` has no TTL and is never reconciled.** A queued task that is
   later deleted from the DB, or whose ID becomes invalid, stays in the list
   forever. The orphan cleaner only scans `tasks:running`, never the queue.
2. **`tasks:active:<projectID>` and `tasks:running` have no key-level TTL.**
   They are emptied member-by-member; if a `delete_running` / `active_remove`
   path is missed (Pub/Sub loss, crash between DB write and Redis op) the IDs
   stay. The orphan cleaner only removes IDs whose *owner node* is dead — a
   stale ID with no `owners` entry is skipped (`orphan_cleaner.go:80-83`,
   `continue` on missing owner).
3. **HASH field expiry depends on `HEXPIRE` (Redis ≥ 7.4).** On Redis < 7.4
   `HExpire` fails and the error is ignored, so `tasks:owners`,
   `tasks:task_project`, and `tasks:aliases` **grow unbounded forever**. There
   is no version check and no documented minimum Redis version.
4. **`ha:nodes` set has no TTL** and is pruned only by an O(N) scan inside
   `heartbeat` / `NodeCount`. If every node crashes hard, stale node IDs
   linger until some surviving node happens to prune them.
5. **`tasks:claim:<taskID>` orphaned by a node that dies between `TryClaim`
   and `SetRunning`.** Such a task is not in `tasks:running`, so the orphan
   cleaner never sees it; it is unrunnable until the 5-minute claim TTL lapses.
6. **In-process caches `byID` / `byAlias` leak.** They are only pruned by
   `dequeue` / `delete_running` / `active_remove` Pub/Sub events. Pub/Sub is
   fire-and-forget — a missed event leaves a `TaskRunner` cached forever. No
   periodic reconciliation against Redis exists.

## 2. Problems With the Current ("Old Model") Implementation

### 2.1 Correctness / reliability

- **P1 — Nil-pointer panic when HA is unconfigured.**
  `pro_impl/services/tasks/task_state_store_factory.go:9` reads
  `util.Config.HA.Enabled` directly. `util.Config.HA` is a `*HAConfig` and is
  `nil` unless configured. The rest of the codebase uses the nil-safe
  `util.HAEnabled()`. This factory will panic on any install that has no `ha`
  block in its config.
- **P2 — Pub/Sub message loss causes silent divergence.** Redis Pub/Sub drops
  messages on a brief disconnect or slow consumer. Missed `delete_*` events
  leak cache entries; missed `enqueue` events mean a node never learns about a
  queued task until a restore. There is no sequence number, no replay, no
  reconciliation tick.
- **P3 — Orphan detection latency.** `orphanCheckInterval` is 60s and node TTL
  is 30s, so a dead node's running task can stay "running" for up to ~90s
  before being failed. The cleaner also ignores orphaned *queued* and
  *claimed-not-running* tasks entirely.
- **P4 — Schedule dedup is clock-skew sensitive.** The lock key embeds
  `time.Now().UTC().Truncate(time.Minute)`. Two nodes whose clocks differ by a
  few seconds across a minute boundary truncate to different minutes and both
  fire the same schedule.
- **P5 — No timeouts on Redis calls.** Every call uses `context.Background()`.
  A hung or failing-over Redis blocks task-pool and WS goroutines
  indefinitely; there is no `MaxRetries`, no dial/read/write timeout, no
  circuit breaking.
- **P6 — `IsNodeAlive` returns `true` on Redis error**, so during a Redis
  outage the orphan cleaner believes every node is alive (fail-open). Combined
  with P5 this means a Redis blip can wedge the cluster.
- **P7 — Non-atomic claim + dequeue.** `handleQueue` does `TryClaim` →
  `DequeueAt` → `runTask` as separate Redis ops (`TaskPool.go:253-259`).
  There is no Lua/`MULTI` wrapping; a crash mid-sequence leaves a claimed task
  still in the queue.

### 2.2 Performance / scalability

- **P8 — Read amplification on every cache hit.** `getOrHydrate` calls
  `LoadRuntimeFields` (an `HGETALL`) on *every* hit
  (`redis_task_state_store.go:209-212`). `QueueRange`, `RunningRange`,
  `GetActive` loop over all tasks → **N Redis round-trips per list call**, and
  these lists are read on every API request and every 5s pool tick.
- **P9 — DB-load amplification from Pub/Sub.** Every `enqueue` / `set_running`
  / `active_add` / `alias_set` event triggers `updateTaskRunner`, which calls
  the hydrator → a DB query, on *every* node. An N-node cluster multiplies task
  lifecycle DB reads by N.
- **P10 — `handleQueue` is O(N) Redis calls per event.** `QueueLen` +
  `QueueGet(i)` per index are individual round-trips; the queue is re-walked
  on every Pub/Sub event and every 5s tick.
- **P11 — Five independent Redis clients / connection pools.** Each component
  builds its own client (`NewRedisClient` plus a hand-rolled one inside
  `redis_task_state_store.go`). Client-construction code is duplicated and
  diverges (the store does not call `NewRedisClient`).
- **P12 — Single Pub/Sub channel for all events.** `tasks:events` and
  `ha:ws:broadcast` are global; every node decodes and filters every event for
  every project/user. No sharding.
- **P13 — `NodeCount()` / `pruneStaleNodes` are O(N) `EXISTS` calls** per
  invocation instead of one ranged query.

### 2.3 Operability / security

- **P14 — No observability.** No metrics for Redis latency, Pub/Sub lag, node
  count, orphan-cleanup counts, cache size. Failures are only `log.Error`.
- **P15 — No multi-tenant isolation.** Key prefixes `tasks:` / `ha:` are
  hard-coded; two Semaphore clusters cannot share one Redis except by DB
  number. No configurable namespace.
- **P16 — `TLSSkipVerify` is exposed** and there is no minimum-Redis-version
  validation at startup, so misconfigurations surface only as silent leaks.
- **P17 — No Redis Sentinel / Cluster support.** `redis.NewClient` only — a
  single Redis is itself a SPOF, undermining the "HA" promise.

## 3. Proposed Improvements

Organised into three phases by risk and value.

### Phase 1 — Correctness & safety (low risk, high value)

1. **Fix P1:** change the factory to `util.HAEnabled()`.
2. **Shared Redis client (P11):** add `pro_impl/services/ha/redis.go` exposing
   a single process-wide `*redis.Client` (lazy singleton) with proper
   `Options`: `PoolSize`, `MinIdleConns`, `MaxRetries`, `DialTimeout`,
   `Read/WriteTimeout`. Delete the duplicated constructor in
   `redis_task_state_store.go`. Inject this client into all five components.
3. **Per-call timeouts (P5):** wrap every Redis call in
   `context.WithTimeout`. Add a small `redisCall` helper to keep call sites
   tidy.
4. **Startup validation (P3/P16):** on `Start()`, `PING` Redis and read
   `INFO server`; if `redis_version < 7.4`, log a loud warning (or refuse, by
   config) because hash-field TTLs silently no-op below 7.4.
5. **Reconciliation tick (P2/P6, leak #6):** add a periodic
   (e.g. 30–60s) full reconcile that re-reads `tasks:running` /
   `tasks:queue` / `tasks:active:*` from Redis and drops `byID` / `byAlias`
   entries no longer present. This makes Pub/Sub a fast-path optimisation
   rather than a correctness dependency.
6. **Schedule dedup clock-safety (P4):** include a small grace window — either
   acquire locks for both the current and previous minute bucket, or key by
   the schedule's *intended* fire time (already known to the cron scheduler)
   instead of `time.Now()`.

### Phase 2 — Leak elimination & orphan robustness

7. **Sorted-set node registry (leak #4, P13):** replace `ha:node:<id>` + the
   `ha:nodes` SET with a single `ha:nodes` **ZSET** scored by last-heartbeat
   epoch. Heartbeat = one `ZADD`; liveness = `score > now - TTL`;
   `NodeCount` = `ZCOUNT`; periodic `ZREMRANGEBYSCORE` prunes the dead. No
   per-node keys, no N+1 `EXISTS`.
8. **Orphan cleaner covers all states (P3, leaks #1/#2/#5):** scan
   `tasks:queue` and `tasks:claim:*` in addition to `tasks:running`; for an
   entry whose owner is dead *or* whose owner record is missing *and* the
   task no longer exists in the DB, remove it everywhere. Drive cleanup off
   `taskID` consistently so empty `active:<projectID>` sets disappear.
9. **Guaranteed key expiry (leak #3):** on Redis < 7.4, fall back to storing
   per-task metadata as individual `EXPIRE`-able keys instead of HASH fields,
   so nothing can grow unbounded regardless of Redis version.
10. **Atomic claim+dequeue (P7):** move `TryClaim` + `LREM` into a single Lua
    script so a node either fully owns a task or leaves the queue untouched.
11. **Reduce orphan latency (P3):** shorten `orphanCheckInterval` to ~20–30s
    and make node TTL / heartbeat interval configurable.

### Phase 3 — Performance & scale

12. **Cache runtime fields, drop per-hit `HGETALL` (P8):** stop calling
    `LoadRuntimeFields` on every `getOrHydrate` hit; rely on `runtime_update`
    Pub/Sub events plus the Phase-1 reconcile tick. Add a short TTL'd local
    freshness marker if a refresh is still needed.
13. **Pipeline list reads (P8/P10):** `QueueRange` / `RunningRange` /
    `GetActive` should fetch all IDs, then hydrate misses in one pipelined
    batch; `handleQueue` should snapshot the queue once per pass instead of
    `QueueGet(i)` per index.
14. **Cut DB amplification (P9):** carry enough state in the Pub/Sub envelope
    (`taskRef` already has status-relevant fields) so other nodes update
    caches *without* a hydrator/DB call; hydrate lazily only when a node
    actually needs to serve that task.
15. **Optional: Redis Streams instead of Pub/Sub (P2/P12):** a consumer-group
    stream gives at-least-once delivery, replay after reconnect, and natural
    sharding — removing the root cause of P2 rather than papering over it.
16. **Sentinel / Cluster support (P17):** allow `redis.NewFailoverClient` /
    `NewClusterClient` via config so Redis itself is not a SPOF.

### Cross-cutting

17. **Observability (P14):** Prometheus metrics — `ha_nodes_total`,
    `ha_redis_call_duration`, `ha_redis_errors_total`,
    `ha_orphan_tasks_cleaned_total`, `ha_pubsub_events_total`,
    `ha_task_cache_size`. Surface node list / health in the admin API.
18. **Configurable key namespace (P15):** add `SEMAPHORE_HA_KEY_PREFIX` so
    multiple clusters can share one Redis safely.

## 4. Suggested Sequencing

1. Phase 1, items 1–4 — small, isolated, ship immediately (fixes a crash and
   the silent-leak-on-old-Redis footgun).
2. Phase 1, items 5–6 + Phase 2, items 7–9 — the correctness backbone.
3. Phase 2, items 10–11 and Phase 3 — performance and scale, behind the now
   solid correctness base.

## 5. Open Questions

- Minimum supported Redis version — commit to 7.4+, or keep a < 7.4 fallback?
- Is Redis Streams acceptable as a dependency, or must Pub/Sub stay for
  compatibility with managed Redis offerings that restrict Streams?
- Expected cluster size (number of nodes, tasks/min) — sets the bar for how
  far Phase 3 sharding needs to go.

## 6. Test Strategy

- Unit tests with `miniredis` for each store/registry method (key shapes,
  TTLs, expiry fallback).
- A fault-injection test that drops Pub/Sub messages and asserts the
  reconcile tick converges caches.
- A simulated dead-node test asserting orphan cleanup of running, queued, and
  claimed-not-running tasks within the configured window.
- Clock-skew test for schedule dedup across a minute boundary.
</content>
</invoke>
