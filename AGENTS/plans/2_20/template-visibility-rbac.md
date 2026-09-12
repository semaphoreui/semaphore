# Implementation Plan — Template Visibility for Custom Roles (Extended RBAC)

## Goal

Let several IT teams work inside **one project** while each team can **see and run
only the task templates assigned to it**. Example from the user request:

- "Workstation and application installation" — visible and runnable only by the
  Infrastructure team.
- "Requirements collection" — visible and runnable only by the Support team,
  which must not even see the Infrastructure templates or their task history.

Running only assigned templates already works in Enterprise (custom roles +
per-template role permissions, see `docs/docs/user-guide/team.md`, section
"Extended RBAC"). What is missing is **visibility**: today every project member
can read every template, its task history, task logs and schedules.

This plan adds a new permission bit `CanViewProjectTemplates` and enforces it on
every read path. The feature is Enterprise-only, like the rest of Extended RBAC.

## Current State (reference paths)

- `db/ProjectUser.go` — `ProjectUserPermission` bits (1, 2, 4, 8) and
  `rolePermissions` for the four built-in roles.
- `db/Role.go` — `Role` (custom role, table `role`, `project_id` NULL = global) and
  `TemplateRolePerm` (table `project__template_role`).
- `api/projects/project.go:25-80` — `ProjectMiddleware`: resolves role
  permissions, ORs template-level permissions when `template_id` is in the URL,
  puts `"permissions"` into context.
- `api/projects/project.go:85` — `GetMustCanMiddleware`: **skips the check for
  GET/HEAD**. This is why read access is unrestricted for every member.
- `db/sql/template.go:290-300` — `getTemplates` LEFT JOINs `project__user` and
  `project__template_role` only to return the `permissions` column; no filtering.
- `db/sql/template.go:523` — `GetTemplatePermission(projectID, templateID, userID)`
  returns built-in role bits OR template-role bits. For a **custom** role it does
  not include the role's project-wide bits (those are added in `ProjectMiddleware`).
- `db/sql/task.go:277` — `getTasks` (used by `GetProjectTasks`, `GetTemplateTasks`,
  last tasks, workflow run tasks). Joins `project__template as tpl`; no user
  filtering.
- `api/projects/tasks.go:203` — `GetTaskPermissionsMiddleware`: ORs template
  perms for task-scoped routes. Bug: on error it writes 400 **and still calls
  `next`**.
- `api/router.go`:
  - `:307` `projectUserAPI` — all project read routes go through
    `ProjectMiddleware` + `GetMustCanMiddleware(CanManageProjectResources)`.
  - `:331-345` `/tasks`, `/tasks/last`, `/stats`, `/templates`, `/schedules`,
    `/views`, `/workflows`.
  - `:461` `projectTmplManagement` (`/templates/{template_id}/...` incl. `tasks`,
    `tasks/last`, `schedules`, `stats`, `refs`, `perms`).
  - `:505` `projectTaskManagement` (`/tasks/{task_id}/output`, `raw_output`,
    `stages`, `ansible/hosts`, `ansible/errors`).
  - `:311` `/events`, `/events/last` — `db.Event` has `ObjectType`/`ObjectID`
    (template / task ids leak through descriptions).
  - `:529` `/views/{view_id}/templates` → `GetViewTemplates`
    (`api/projects/views.go:37`).
- `services/tasks/TaskRunner.go:490-513` — task websocket updates are sent to
  **every** project user (`GetProjectUsers`) plus admins.
- `services/tasks/TaskRunner_logging.go:53` — log lines are sent to the same
  `t.users` list.
- `web/src/lib/constants.js:21` — `USER_PERMISSIONS` (1, 2, 4, 8).
- `web/src/components/EditRoleForm.vue`, `EditTemplatePermissionForm.vue`
  (hard-coded `[1, 2, 4, 8]`), `PermissionsCheck.js`, `TemplatePermissionsChips.vue`.
- `web/src/views/project/Templates.vue:469` — `canRun(item)` already uses
  `item.permissions`.
- `db/Migration.go:140` — last registered migration `2.20.4`.
- Docs: `docs/docs/user-guide/team.md` (Extended RBAC section).

## Design

### New permission bit

```go
// db/ProjectUser.go
const (
    CanRunProjectTasks ProjectUserPermission = 1 << iota // 1
    CanUpdateProject                                     // 2
    CanManageProjectResources                            // 4
    CanManageProjectUsers                                // 8
    CanViewProjectTemplates                              // 16 (new)
)
```

Semantics mirror `CanRunProjectTasks`:

| Level | Meaning of bit 16 |
| --- | --- |
| Role (project-wide) | User sees **all** templates of the project and everything derived from them (tasks, logs, schedules, stats, events, websocket updates). |
| `project__template_role` | User sees **this** template and everything derived from it. |

**Visibility rule** (single function, reused everywhere):

```
visible(template) =
    user.Admin
 || rolePerms & CanViewProjectTemplates != 0
 || templateRolePerms(template, role) != 0        // any bit on the template implies view
```

"Any bit implies view": a role that may run or update a template obviously has to
see it. Bit 16 on the template level exists for **read-only** visibility (see
history and logs, but not run).

All four built-in roles get `CanViewProjectTemplates`, so **community edition and
existing projects behave exactly as today**.

### Backward compatibility / migration

Custom roles created before this feature have no bit 16 and would suddenly lose
visibility of everything. Migration sets the bit on all existing custom roles:

```sql
update `role` set permissions = permissions | 16;
```

`permissions` is an integer column (`bigint`/`integer` per dialect), bitwise OR is
supported by MySQL, PostgreSQL and SQLite. Verify with a migration test like
`db/sql/migration_2_20_0_test.go`.

Template-level rows (`project__template_role`) are **not** touched: their existing
non-zero permissions already imply visibility under the rule above.

In the role form the new checkbox is **checked by default** for new roles; an
administrator unchecks it to create a "team" role.

### Effective permissions helper

Add one authoritative function and use it from middleware, websocket and
workflow code instead of hand-rolled OR-ing:

```go
// db/Store.go (Store / TemplateRepository)
// Returns rolePerms | templatePerms for the user. Admin is NOT handled here.
GetEffectiveTemplatePermissions(projectID, templateID, userID int) (ProjectUserPermission, error)
```

Implementation reuses `GetProjectOrGlobalRoleBySlug` and the existing
`project__template_role` lookup (`db/sql/template.go:523`). Keep
`GetTemplatePermission` as is for now or make it a thin wrapper.

### SQL-level filtering

Both `getTemplates` and `getTasks` receive an optional viewer:

```go
// db/Template.go
type TemplateViewer struct {
    UserID  int
    ViewAll bool // admin or role has CanViewProjectTemplates
}
```

`nil` viewer ⇒ no filtering (internal callers: scheduler, backup, runners,
workflows engine). Non-nil with `ViewAll == false` ⇒ add:

```sql
LEFT JOIN project__user pu  ON pu.project_id = pt.project_id AND pu.user_id = ?
LEFT JOIN project__template_role ptr ON ptr.template_id = pt.id AND ptr.role_slug = pu.role
WHERE ptr.permissions IS NOT NULL AND ptr.permissions <> 0
```

`ViewAll` is computed in Go from the context `"permissions"` value; this avoids
joining the `role` table in SQL and keeps built-in roles code-defined (same
reasoning as the comment in `api/projects/project.go:46`).

Filtering must be in SQL, never in Go after fetching: task history uses keyset
pagination (2.19) and per-page `LIMIT`, post-filtering would break page sizes.

---

## Implementation Steps

### Phase 1 — Core permission bit and helpers

**1.1 `db/ProjectUser.go`** — add `CanViewProjectTemplates`; add it to all four
entries of `rolePermissions`. Update `db/ProjectUser_test.go`.

**1.2 Migration `db/sql/migrations/v2.20.5.sql`** (check the next free version at
implementation time; register in `db/Migration.go` after `2.20.4`):

```sql
update `role` set permissions = permissions | 16;
```

Add `db/sql/migration_2_20_5_test.go` (SQLite) asserting a pre-existing custom
role with `permissions = 1` becomes `17`.

**1.3 `db/Store.go` + `db/sql/template.go`** — `GetEffectiveTemplatePermissions`
(see Design). Table-driven tests in `db/sql/` for: built-in role, custom role
with project-wide bits only, custom role with template bits only, user not in
project (returns 0, no error), unknown role slug.

**1.4 `db/Template.go`** — `TemplateViewer`. Extend `TemplateFilter` with
`Viewer *TemplateViewer`.

**1.5 `db/sql/template.go` `getTemplates`** — apply the WHERE from Design when
`filter.Viewer != nil && !filter.Viewer.ViewAll`. Reuse the existing joins (they
are already added when `userID != nil`).

**1.6 `db/sql/task.go` `getTasks`** — add `viewer *db.TemplateViewer` parameter,
apply the same joins/WHERE on `tpl`. Thread it through `GetProjectTasks`,
`GetTemplateTasks`, `GetLastTasks`-style callers and the workflow-run task getter.
Internal callers pass `nil`.

**1.7 `db/sql/schedule.go` `GetProjectSchedules`** — same viewer parameter, join
`project__template` when filtering. `GetTemplateSchedules` needs no change (the
template route is gated in Phase 2).

**1.8 `db/sql/event.go` `GetEvents`** — events whose object is derived from a
template must be filtered by template visibility (`EventObjectType` values in
`db/Event.go:56-65`):

- `task` — join `task` → `project__template` by `object_id`.
- `template` — join `project__template` directly by `object_id`.
- `schedule` — join `project__schedule` → `project__template`.
- `workflow` — hide when the workflow references a hidden template (same rule as
  step 2.6); simplest is to hide all workflow events for restricted viewers.

Events of other object types (`environment`, `inventory`, `key`, `project`,
`repository`, `user`) are unaffected.

### Phase 2 — API enforcement

**2.1 Helper `projects.TemplateViewerFromContext(r)`** in `api/projects/` —
builds `*db.TemplateViewer` from `"user"` (admin) and `"permissions"` context.
Returns `nil` when the user may view all (avoids pointless joins).

**2.2 Read-gate middleware for template-scoped routes.** `GetMustCanMiddleware`
intentionally ignores GET, so add a dedicated middleware and register it on
`projectTmplManagement` (`api/router.go:462`) after `TemplatesMiddleware`:

```go
// 403 unless admin || (permissions & CanViewProjectTemplates) != 0
//   where "permissions" already includes template-level bits (ProjectMiddleware)
func MustCanViewTemplateMiddleware(next http.Handler) http.Handler
```

Because "any template bit implies view", the check is:
`admin || perms & CanViewProjectTemplates != 0 || templatePerms != 0`. Store the
raw template-level bits in context (`"templatePermissions"`) from
`ProjectMiddleware` to make this check explicit instead of inferring it.

Covers: `GET /templates/{id}`, `/refs`, `/tasks`, `/tasks/last`, `/schedules`,
`/stats`, `/perms`, `/perms/{perm_id}` and all mutating routes (they are already
gated by `CanManageProjectResources` / `CanRunProjectTasks`).

**2.3 Task-scoped routes** (`projectTaskManagement`, `api/router.go:505`):
extend `GetTaskPermissionsMiddleware` to use `GetEffectiveTemplatePermissions`
and return 403 for GET when the visibility rule fails. **Fix the existing bug**:
`return` after writing the error status. Register it on `projectTaskManagement`
(today it is only on start/stop routes).

**2.4 Project-wide lists** — pass the viewer:

- `projects.GetTemplates` (`api/projects/templates.go:70`)
- `projects.GetViewTemplates` (`api/projects/views.go:37`)
- `taskController.GetAllTasks`, `GetLastTasks`, `GetTaskStats`
  (`api/projects/tasks.go`); for stats the SQL in `GetTaskStats` store method
  must also take the viewer.
- `projects.GetProjectSchedules`
- `getAllEvents` / `getLastEvents` for the project variant (`api/router.go:311`)

**2.5 Task start with explicit template id** (`api/router.go:295`) — already
gated by `CanRunProjectTasks`; no change, but add a test proving a role with
bit 16 only (view) cannot start a task.

**2.6 Workflows (pro).** `GetWorkflows`, `GetWorkflow`, `GetWorkflowRuns`,
`GetWorkflowRun*` — hide a workflow if it references at least one template the
user cannot see. Implement in `pro_impl` behind the same `TemplateViewer`;
document the decision in the workflow plan if it changes.

**2.7 `api-docs.yml`** — document bit 16 in the permissions description of
`Role`, `TemplateRolePerm` and the `GET /project/{id}/role` response; add 403
responses to template/task GET routes.

### Phase 3 — Websocket and log streaming

**3.1 `services/tasks/TaskRunner.go:490`** — when collecting `t.users`, keep only
users for whom `GetEffectiveTemplatePermissions(...)&CanViewProjectTemplates != 0`
or template bits are non-zero. Admins stay. One DB round-trip per project user at
task start is acceptable; do not cache (HA: every node computes it itself from
the DB, no shared state needed).

**3.2 `TaskRunner_logging.go`** uses the same `t.users`, so it is covered.

**3.3 Remote task stop broadcast (`TaskPool.go:216`)** is node-to-node, not
user-facing; no change.

### Phase 4 — UI

**4.1 `web/src/lib/constants.js`** — `viewProjectTemplates: 16`.

**4.2 `EditRoleForm.vue`** — checkbox `canViewProjectTemplates` (bit 16),
default checked for new roles; hint text: "Uncheck to restrict this role to the
templates it is explicitly added to."

**4.3 `EditTemplatePermissionForm.vue`** — replace the hard-coded `[1, 2, 4, 8]`
with `Object.values(USER_PERMISSIONS)`; add checkbox "Can view" (bit 16).
`TemplatePermissionsChips.vue` — chip for the new bit.

**4.4 Translations** — `canViewProjectTemplates`, `canViewTemplate` in
`web/src/lang/*.js` (English first, other languages fall back).

**4.5 Navigation** — nothing to hide: lists come back already filtered. Verify
Dashboard (last tasks), History and Schedules pages render correctly with an
empty result for a restricted user. Templates list: keep `canRun(item)` logic.

**4.6 Team page** — no change; a user still has exactly one role.

### Phase 5 — Tests

- `db/ProjectUser_test.go` — built-in roles include bit 16.
- `db/sql` — `getTemplates`/`getTasks`/`GetProjectSchedules`/`GetEvents` with
  viewer: restricted user sees only templates with a role row; `ViewAll` sees all;
  pagination `LIMIT` respected after filtering.
- `api/projects/tasks_test.go` — `GetTaskPermissionsMiddleware` returns 403 for a
  hidden task and stops the chain on store error.
- `api/projects` — template-scoped GET returns 403 for a restricted role, 200 for
  a role with a template row, 200 for admin.
- Migration test for `2.20.5`.
- Manual check on the screenshot stand (`ui-screenshot-stand` memory): two custom
  roles, two templates, two users; each user sees only their template in
  Templates, History, Dashboard, Schedules and receives no websocket update for
  the other team's task.

### Phase 6 — Documentation (`docs/` submodule)

`docs/docs/user-guide/team.md`, Extended RBAC section:

- Add **Can view project templates** to the "Project-wide permissions" table and
  **Can view** to template permissions.
- Replace the tip "Template-only access" with a recipe **"Restrict a team to its
  own templates"**: create role without project-wide permissions (uncheck
  *Can view project templates*), add the role to each team template with
  *Can run tasks*, assign the role to team members.
- Update the Guest description and FAQ 5: Guest still sees everything; hiding
  requires a custom role.
- Add to "Not currently supported": a user has one role per project, so a member
  of two teams needs a combined role; non-template resources (inventory, keys,
  repositories) stay visible to all members.

---

## Security Notes

- Default-deny on every **new** code path: a missing role row or a store error
  must yield 403, never fall through to `next`.
- Filtering is done in SQL with the viewer computed server-side from the session;
  never trust `permissions` sent by the client.
- Admins bypass visibility (consistent with the rest of the API).
- API tokens go through the same middleware chain; no separate path.
- Backup/restore and export already require `CanManageProjectResources`, which is
  only granted to roles that can see everything anyway. Keep it that way: a role
  with `CanManageProjectResources` but without bit 16 is a misconfiguration; the
  role form should force bit 16 on when bit 4 is checked.

## High Availability

Stateless: all decisions are made per request from the database. The websocket
user list is computed on the node that starts the task, from the same tables. No
new caches or globals.

## Out of Scope (possible follow-ups)

- Multiple roles per user / groups mapped from LDAP or OIDC.
- Permissions on Views ("team folders") as a convenience layer on top of
  per-template rows (`project__template.view_id` already exists).
- Granular visibility of inventories, keys, repositories, environments.
