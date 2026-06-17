# Tasks API pagination

The project History page and the `GET /api/project/{project_id}/tasks/last` endpoint use **keyset (cursor) pagination** to page through task history without expensive `COUNT(*)` queries or deep `OFFSET` scans.

## Why keyset

Projects can accumulate millions of tasks. Offset pagination (`LIMIT n OFFSET m`) scans and discards every skipped row, so deep pages get slower. Counting all rows on every request is equally expensive.

Keyset pagination walks backward through the primary-key index using a cursor (`before=<task_id>`), so each page costs the same regardless of depth. There is intentionally **no total page count**.

## Request

```
GET /api/project/{project_id}/tasks/last?count=20
GET /api/project/{project_id}/tasks/last?count=20&before=842
```

| Parameter | Meaning |
|-----------|---------|
| `count` | Page size (default and max: 200) |
| `limit` | Legacy alias for `count` (still accepted) |
| `before` | Cursor: return tasks with `id` strictly less than this value |

Tasks are ordered by `id DESC` (newest first). The first page omits `before`; each subsequent page passes the smallest `id` from the previous page as `before`.

### Template-scoped variant

When the route is mounted under a template context, the same parameters apply but results are scoped to that template's tasks.

## Response

- **Body**: JSON array of `Task` objects (plain array, unchanged from legacy clients).
- **Header**: `X-Has-Next: true` or `X-Has-Next: false` — whether older tasks exist beyond this page.

The server fetches `count + 1` rows internally. If the extra row exists, it is trimmed and `X-Has-Next` is `true`.

Example:

```http
GET /api/project/1/tasks/last?count=20 HTTP/1.1

HTTP/1.1 200 OK
X-Has-Next: true
Content-Type: application/json

[{ "id": 900, ... }, { "id": 899, ... }, ...]
```

Next page:

```http
GET /api/project/1/tasks/last?count=20&before=881
```

## Backward compatibility

| Endpoint / caller | Behavior |
|-------------------|----------|
| `GET .../tasks/last?limit=200` | Returns up to 200 tasks; `X-Has-Next` ignored by legacy callers |
| `GET .../tasks` (`GetAllTasks`) | Unchanged: up to 1000 tasks, no cursor, no `X-Has-Next` |
| `TaskList.vue` (per-template list) | Still uses `limit=200`; header ignored |

## Frontend (History page)

`web/src/views/project/History.vue` maintains a cursor stack:

```
cursors: [null]     // cursors[i] = before id for page i (null = first page)
pageIndex: 0
hasNext: false      // from X-Has-Next
```

- **Next**: push the last row's `id` as the next cursor, increment `pageIndex`, reload.
- **Prev**: decrement `pageIndex`, reload from the stored cursor.
- Footer shows `‹  N  ›` with no total page count.

WebSocket live updates reload the current page by its cursor, so concurrent task creation only affects page 1.

## Cursor stability

Keyset cursors are stable under inserts: a new task gets a higher `id` and only appears on page 1. Navigating older pages is unaffected by concurrent activity.

## Related code

- `api/projects/tasks.go` — `parseTasksPageParams`, `writeTasksList`, `GetLastTasks`
- `db/sql/task.go` — `WHERE task.id < ?` keyset filter
- `db/Store.go` — `RetrieveQueryParams.BeforeID`
- [`AGENTS/plans/2_19/tasks-history-keyset-pagination.md`](../AGENTS/plans/2_19/tasks-history-keyset-pagination.md) — design notes
