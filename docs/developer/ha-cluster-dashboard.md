# HA cluster dashboard and task state API

## Purpose

Semaphore can run in **high-availability (HA)** mode: several app nodes share one Redis-backed task queue and related coordination (see `util.HAEnabled()` and `services/tasks`). The **Cluster** screen (`/cluster`) helps administrators inspect membership, Redis usage, and in-memory task-pool records, and perform targeted maintenance clears when state is stuck.

OSS builds use an in-process task state store; Enterprise/Pro inject Redis-backed stores and a cluster inspector. The UI still renders for all admins; capability differs by license (`features.high_availability`) and config (`ha_enabled` in API responses).

## UI entry points

- **Route:** `/cluster` (`web/src/router/index.js`).
- **Navigation:** Admin sidebar only (`web/src/App.vue` — `user.admin`).
- **Feature gating:** `PageMixin` exposes `features` from `GET /api/info` (`SystemInfo.Features`). The destructive **Clear tasks from Redis** action stays disabled unless `features.high_availability` is true **and** cluster status reports `ha_enabled: true` (`web/src/views/Cluster.vue`).
- **Enterprise banner:** When `high_availability` is false, the page shows an upgrade notice; nodes/Redis panels still appear if the server returns them (e.g. licensed HA).

## REST API (admin only)

Handlers are mounted on the **admin** sub-router (`api/router.go`): callers must be authenticated administrators.

### `GET /api/cluster`

Returns JSON including:

| Field | Meaning |
| --- | --- |
| `ha_enabled` | `true` when `SEMAPHORE_HA_ENABLED` (via config) is on and HA config is present (`util.HAEnabled()`). |
| `node_id` | This instance’s HA node id when `util.Config.HA` is set. |
| `nodes` | Cluster membership (from `pro_interfaces.ClusterInspector`), when an inspector is injected and listing succeeds. |
| `redis` | Redis connection and key-group stats from the same inspector. |

**Responses**

- HA **disabled:** `200` with at least `{ "ha_enabled": false }`. No cluster inspector is required.
- HA **enabled** but no inspector (typical OSS): `503` with message *cluster inspection is unavailable (HA mode disabled or overlay missing)*.
- HA **enabled** with inspector: `200` with `ha_enabled`, `node_id`, and best-effort `nodes` / `redis` (errors on those slices are logged; missing keys are omitted).

### `GET /api/cluster/tasks`

Returns a `TaskStateSnapshot` (`services/tasks/task_state_store.go`): `queue`, `running`, `active_by_project`, `aliases`, `claims`. Works without HA; if the active `TaskPool` has no `TaskStateInspector`, the handler returns empty structures (`200`).

### `DELETE /api/cluster/tasks`

Body: `{ "scope": { "queue": bool, "running": bool, "active": bool, "aliases": bool, "claims": bool, "runtime_fields": bool } }` — at least one flag must be true (`api/cluster.go`).

Removes the selected record groups from the backing store. Intended **only** for recovery from inconsistent Redis/task state. Responds `501` if the store does not support clearing.

## Task details: runner name

Task list/detail payloads include `used_runner_name` when the task is associated with a runner: the SQL layer joins `runner.name` (`db/sql/task.go`, `db.Task.UsedRunnerName`). The task detail template shows a row when that field is set (`web/src/components/TaskDetails.vue`).

## Related code

- API: `api/cluster.go`, `api/cluster_test.go`
- Router: `api/router.go` (`/cluster` admin routes)
- Types: `pro_interfaces/ha.go`, `services/tasks/task_state_store.go`, `services/tasks/TaskPool.go` (`StateStore`, HA integration)
- UI: `web/src/views/Cluster.vue`, `web/src/components/RedisMemoryChart.vue`

For Redis key design and known limitations of the HA layer, see [`docs/plans/ha-cluster-improvements.md`](../plans/ha-cluster-improvements.md).
