# Plan — Real Pagination for the Project History Page (Keyset)

## Goal

The project History page (`/project/:id/history`) loaded the **200 most recent
tasks** in a single request and then paged through them 20-at-a-time entirely on
the client. The backend query is heavy (each task is `Fill`-ed with template,
user and runner data), so fetching 200 rows up front is wasteful, and the user
never sees anything past the newest 200 tasks.

Replace the pseudo-pagination with **real, server-side pagination**: the backend
returns one page at a time, and the page the user navigates to is the page the
database actually fetches.

A project can accumulate **millions of tasks**. That rules out the two classic
"page N of M" mechanisms:

- `COUNT(*)` to compute the total number of pages — a full scan on every page
  load.
- `LIMIT n OFFSET (page-1)*n` — the database still scans and discards every
  skipped row, so deep pages get linearly slower.

So the final design uses **keyset (cursor) pagination** with **no total count**.

## Scope

In scope:

- Refactor `api/projects/tasks.go` so every task handler and middleware is a
  method on `TaskController` with an injected `db.Store`. Remove all
  `helpers.Store(r)` usage from the file.
- Keyset pagination for the tasks list endpoint (`GET .../tasks/last`), driven
  by `db.RetrieveQueryParams`.
- Frontend cursor-based Prev/Next navigation on the History page.

Out of scope / deliberately unchanged:

- `GET .../tasks` (`GetAllTasks`) keeps its existing behaviour: up to 1000 tasks,
  no cursor, no `X-Has-Next`. Consumers like the build-version pickers in
  `TaskForm.vue` / `TaskParamsForm.vue` rely on it.
- `TaskList.vue` (the per-template task list) keeps calling
  `.../tasks/last?limit=200`. The legacy `limit` parameter is still honoured and
  the new `X-Has-Next` header is simply ignored by that component.
- The Bolt backend — there is none; only `db/sql` implements the task store.
- Server-side status/sort filtering on the History page (the params plumbing is
  there via `RetrieveQueryParams`, but no UI yet).

## Backward Compatibility

1. `.../tasks/last?limit=N` still returns up to N latest tasks (legacy callers).
2. `.../tasks` still returns up to 1000 tasks.
3. The response body stays a **plain JSON array** of tasks. Pagination metadata
   is carried in the `X-Has-Next` response header, so non-paginating consumers
   are unaffected.
4. `db.RetrieveQueryParams` gains one additive field (`BeforeID int`); the zero
   value preserves existing behaviour for every other caller (e.g.
   `services/export/Task.go`).

## Backend Design

### Controller refactor (`api/projects/tasks.go`)

`TaskController` now carries the store:

```go
type TaskController struct {
    store           db.Store
    ansibleTaskRepo db.AnsibleTaskRepository
}

func NewTaskController(store db.Store, ansibleTaskRepo db.AnsibleTaskRepository) *TaskController
```

All former package-level functions became methods on `*TaskController`:
`AddTask`, `GetAllTasks`, `GetLastTasks`, `GetTask`, `GetTaskStages`,
`GetTaskOutput`, `GetTaskRawOutput`, `ConfirmTask`, `RejectTask`, `StopTask`,
`RemoveTask`, `GetTaskStats`, plus the middlewares `NewTaskMiddleware`,
`GetTaskMiddleware`, `GetTaskPermissionsMiddleware`. Every `helpers.Store(r)`
call was replaced with `c.store`. `taskPool(r)` and `outputToBytes(...)` stay as
free helpers (no store dependency). `api/router.go` was updated to reference
`taskController.X` everywhere (handlers and `.Use(...)` middleware).

### Keyset pagination

`db.RetrieveQueryParams` gains:

```go
// BeforeID enables keyset (cursor) pagination for id-ordered lists.
// When > 0, only rows with primary key id strictly less than BeforeID are
// returned. Alternative to Offset that does not get more expensive deeper in.
BeforeID int
```

`db/sql/task.go` `getTasks(...)` adds, before the existing `LIMIT` and the
existing `ORDER BY id DESC`:

```go
if params.BeforeID > 0 {
    q = q.Where("task.id < ?", params.BeforeID)
}
```

This walks backwards through the primary-key index — cheap regardless of how
deep the user pages. (The earlier `OFFSET` experiment was removed.)

### Detecting the next page without COUNT

The controller asks for **one extra row** (`Count = pageSize + 1`). If the store
returns more than `pageSize`, a next page exists; the extra row is trimmed and
the fact is reported via the `X-Has-Next` header. No `COUNT(*)` is ever issued.

`parseTasksPageParams(query, base) (params, pageSize)`:

- `count` (or legacy `limit`) → `pageSize`, capped at `maxTasksPageSize` (200),
  default 200.
- `before` (a task id) → `params.BeforeID`.
- sets `params.Count = pageSize + 1`.

`writeTasksList(w, r, params, pageSize)` fetches project- or template-scoped
tasks (depending on the `template` context value), and when `pageSize > 0`
trims the sentinel row and sets `X-Has-Next: true|false`.

### Request / response contract

```
GET /api/project/{id}/tasks/last?count=20            -> newest 20 tasks
GET /api/project/{id}/tasks/last?count=20&before=842 -> 20 tasks with id < 842
Response: 200, body = JSON array (desc by id), header X-Has-Next: true|false
```

## Frontend Design (`web/src/views/project/History.vue`)

The `v-data-table` no longer paginates client-side: `hide-default-footer`,
`:items-per-page="-1"`, and it renders exactly the page the backend returned.

Cursor state:

```
cursors: [null]   // cursors[i] = `before` id used to load page i (null = first)
pageIndex: 0
hasNext: false    // from X-Has-Next
```

- `loadItems()` fetches `count=20` plus `&before=<cursor>` for the current page;
  sets `items` and `hasNext`.
- `goNext()` — guard on `hasNext`; the last (smallest-id) row of the current
  page becomes the next cursor; push it, advance `pageIndex`, reload.
- `goPrev()` — decrement `pageIndex`, reload from the stored cursor.
- Custom footer: `‹  {{ pageIndex + 1 }}  ›`, buttons disabled on
  `loading` / boundaries. There is intentionally **no "page N of M"** — the
  total is unknown by design.

Keyset cursors are **stable under inserts**: a freshly created task gets a
higher id and only affects page 1, so navigating older pages is unaffected by
concurrent activity (a correctness win over offset pagination). The existing
WebSocket live-update path (`reloadItems` / `onWebsocketDataReceived`) reloads
the current page by its cursor.

## Files Touched

- `db/Store.go` — add `RetrieveQueryParams.BeforeID`.
- `db/sql/task.go` — keyset `WHERE task.id < ?` in `getTasks`.
- `api/projects/tasks.go` — controller refactor; `parseTasksPageParams`,
  `writeTasksList`, `GetLastTasks`, `GetAllTasks`.
- `api/projects/tasks_test.go` — table-driven tests for `parseTasksPageParams`.
- `api/router.go` — `NewTaskController(store, ...)`; route/middleware wiring.
- `web/src/views/project/History.vue` — cursor pagination UI.

## Testing

- Unit: `TestParseTasksPageParams` (count/limit precedence, cap, `before`
  parsing, sentinel `Count = pageSize + 1`, base-param preservation).
- `go build ./...`, `go vet`, and ESLint on `History.vue` are clean.
- Manual: page through History with `count=20`, verify Next stops when
  `X-Has-Next` is false, and that no `COUNT(*)` runs.

## Notes / Follow-ups

- `GetRunnerCount` (`db/sql/runner.go`) has a latent placeholder bug
  (`d.Sql().SelectInt` without `PrepareQuery`) — out of scope here, worth a
  separate fix.
- Possible future work: server-side status/user filters on History, and a
  "jump to newest" affordance.
