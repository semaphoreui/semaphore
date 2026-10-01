# Task history: keyset pagination

The project History page (`/project/:id/history`) and the per-template task list page through
tasks one server page at a time using a keyset cursor (`before=<task id>`) instead of loading 200
rows and paging on the client. There is deliberately no total count and no "page N of M".

Status (2026-09-27): implemented — `db.RetrieveQueryParams.BeforeID` (`db/Store.go`), the
`task.id < ?` clause in `db/sql/task.go` `getTasks`, `parseTasksPageParams`/`writeTasksList`/
`X-Has-Next` in `api/projects/tasks.go` (+ `tasks_test.go`), cursor state in
`web/src/views/project/History.vue` and `web/src/components/TaskList.vue`. Merged as
"tasks pagination (#3946)", 11ea51d6e, 2026-06-12.

## Decisions and why

- **Keyset, not offset, and no `COUNT(*)`.** A project can hold millions of tasks. Offset makes
  the DB scan and discard skipped rows (deep pages get linearly slower); a count is a full scan
  per page load. Walking `id DESC` with `WHERE task.id < ?` is a primary-key range scan whatever
  the depth. An `OFFSET` experiment was tried and removed.
- **Next-page detection by sentinel row.** The controller asks for `pageSize + 1` rows; if the
  extra row exists it is trimmed and `X-Has-Next: true` is set. No second query.
- **Metadata in a header, body stays a plain array.** Existing consumers of `tasks/last`
  (build-version pickers etc.) keep working; they just ignore `X-Has-Next`.
- **Backward compatibility:** `?limit=N` is still honoured as an alias of `count`; `GET .../tasks`
  (`GetAllTasks`) is unchanged at a fixed 1000 rows with no cursor; `BeforeID` zero value keeps
  every other caller (e.g. `services/export`) unaffected. Page size is capped at
  `maxTasksPageSize = 200`.
- **`TaskController` owns its store.** All task handlers and middlewares became methods on
  `TaskController` with an injected `db.Store` (`NewTaskController(store, ansibleTaskRepo)`);
  `helpers.Store(r)` was removed from the file. This follows the no-globals rule and makes the
  handlers unit-testable.
- **Cursors are stable under inserts.** A new task gets a higher id and only affects page 1, so
  older pages do not shift while someone is browsing — a correctness gain over offset paging.
  WebSocket live updates reload the current page by its cursor.
- HA: cursor state is client-side only; any node answers any page. No new authz surface.

## Open, broken or deferred

- ⚠️ 2026-09-27: `TaskList.vue` also moved to cursor paging (`count=${pageSize}` + `X-Has-Next`),
  although the plan said it would keep `limit=200` and ignore the header.
- ⚠️ 2026-09-27: `History.vue` has a per-page selector (10/20/50/100) that resets cursors on
  change; the plan described a fixed 20.
- ⚠️ 2026-09-27: `api-docs.yml` does not document `count`, `before` or the `X-Has-Next` header on
  `/project/{project_id}/tasks/last`.
- Server-side status/user filters on History and a "jump to newest" affordance are not built;
  `RetrieveQueryParams` plumbing exists, UI does not.
- The plan flagged `GetRunnerCount` (`db/sql/runner.go`) as calling `SelectInt` without
  `PrepareQuery`; the code still does so today. It works only while the dialect placeholders
  match squirrel's default `?` — worth a separate check on PostgreSQL.

## Key files and references

- `db/Store.go` (`RetrieveQueryParams.BeforeID`), `db/sql/task.go` (`getTasks`)
- `api/projects/tasks.go`, `api/projects/tasks_test.go` (`TestParseTasksPageParams*`)
- `api/router.go` (`NewTaskController`)
- `web/src/views/project/History.vue`, `web/src/components/TaskList.vue`
- PR https://github.com/semaphoreui/semaphore/pull/3946
- Source plan: `AGENTS/plans/2_19/tasks-history-keyset-pagination.md`
