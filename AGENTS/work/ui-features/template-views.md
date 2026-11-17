# Template views and the "all" view filter

A project's Task Templates page has tabs, one per **view** (`project__view`). A view is either
**custom** (`type = ''`), listing the templates assigned to it through `project__template.view_id`,
or **all** (`type = 'all'`), listing the project's templates that match the view's **filter**.
Migration `2.17.0` created an "All" view in every project.

Status (2026-10-28): implemented on branch `prod-11-filter-on-backend-the-templates-in-view-by-title`
(Linear PROD-11, "Filter (on backend) the templates in View by title"): **a name filter saved on an
"all" view**, applied in SQL. Before PROD-11 the filter column existed but nothing read or wrote it
(`getTemplates` had `// TODO: implement filter`), so every "all" view listed every template.

⚠️ 2026-10-30: the filter is set in Edit Views (the cog next to the tabs), field "Name contains",
shown only for a view of type "all" — not the view's title box, where testers first typed it. A
search box on the page itself was built and removed again: the ticket asks for the backend filter
only, and the author has not asked for a search.

## Decisions and why

- **The filter is `{"name": "<text>"}`, and only that key.** `View.Validate` refuses any other key:
  the backend would store it and silently apply nothing. `View.NameFilter()` trims the value.
- **Contains, ignoring case, in SQL** — `whereNameContains` in `db/sql/template.go`. `LOWER` on
  both sides because `LIKE` is case-sensitive in
  PostgreSQL and not in SQLite or MySQL's default collation. Verified with the same results on
  SQLite, PostgreSQL and MySQL 8.4.
- **`!` is the LIKE escape character** (`escapeLike` in `db/sql/user.go`). SQLite has no default
  escape, so the old backslash escape made `_` and `%` match nothing there — the users search was
  broken on SQLite for any term holding them. A backslash is also written differently in a MySQL
  and a PostgreSQL string literal; `!` is not. Every query using `escapeLike` must say
  `ESCAPE '!'`.
- **The filter applies to "all" views only.** A custom view's templates are chosen by hand. The UI
  still sends the filter for a custom view, so switching a view to custom and back keeps it.
- **The filter is bounded by its encoded JSON**, not the name: `project__view.filter` is
  `varchar(1000)`, and `json.Marshal` writes `<`, `>`, `&` as six bytes each.
- **`UpdateView` keeps the stored filter when the body has none**, and `"filter": {}` clears it.
  Every other field is replaced, but clients written before filters existed — the UI's Edit Views
  form among them — send no `filter` and would otherwise wipe it on any save.
- **`UpdateView` matches on the project and does not write it.** It used to set `project_id` from
  the request body, so a member could move their view into another project. The handler now also
  takes the project from the loaded view.

## Open, broken or deferred

- ⚠️ 2026-10-28: SQLite's `LOWER` folds ASCII only, so on SQLite a filter does not match across the
  case of non-ASCII letters (`über` does not find `Über`). PostgreSQL and MySQL fold them.
- A project restore goes through `CreateView`, which did not write the filter either; it does now,
  so restored "all" views keep their filter.

## Key files and references

- `db/View.go` (`Validate`, `NameFilter`, `ViewFilterName`), `db/sql/view.go`,
  `db/sql/template.go` (`getTemplates`), `db/sql/user.go` (`escapeLike`)
- `api/projects/views.go` (`UpdateView`), `web/src/components/EditViewsForm.vue`
- Tests: `db/View_test.go`, `db/sql/view_test.go`, `api/projects/views_test.go`
- Views were introduced in `5457cf33 feat(views): implement sql`, which left the "all" filter as a TODO
