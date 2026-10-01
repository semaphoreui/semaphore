# Template visibility for custom roles (Extended RBAC)

Open problem: let several teams share one project while each team sees and runs only the task
templates assigned to it. Enterprise custom roles plus per-template role permissions already
restrict *running*; every project member can still *read* every template, its task history,
logs, schedules, stats and events, and receives websocket updates for every task.

Status (2026-09-27): not implemented — no `CanViewProjectTemplates`, `TemplateViewer` or
`GetEffectiveTemplatePermissions` on any branch; `db/ProjectUser.go` still defines bits 1, 2, 4,
8 and `web/src/lib/constants.js` mirrors them; migrations run up to `v2.20.7` with no
`permissions | 16`; `docs/docs/user-guide/team.md` still carries the "Template-only access" tip.
Not implemented; to be created as a workbench task when picked up.

## Motivation

The request: "Workstation and application installation" visible and runnable only by the
Infrastructure team; "Requirements collection" only by Support, who must not even see the
Infrastructure templates or their history. Read access is unrestricted because
`GetMustCanMiddleware` (`api/projects/project.go:94`) skips the check for GET/HEAD, and
`getTemplates` / `getTasks` join `project__template_role` only to report a `permissions` column,
never to filter.

## Agreed approach

- New permission bit `CanViewProjectTemplates = 16`, Enterprise-only like the rest of Extended
  RBAC. On a role it means "see all templates of the project and everything derived from them";
  on a `project__template_role` row it means "see this template" and exists for read-only
  visibility (history and logs without run).
- Visibility rule, one function reused everywhere:
  `admin || rolePerms has 16 || any template-level bit on this template` — a role that may run
  or edit a template obviously sees it.
- All four built-in roles get bit 16, so community edition and existing projects behave exactly
  as today. A migration ORs 16 into every existing custom role (`update role set permissions =
  permissions | 16`, bitwise OR works on MySQL, PostgreSQL and SQLite); template-level rows are
  untouched because any non-zero value already implies view. The role form checks the new box by
  default; an admin unchecks it to create a "team" role.
- One authoritative `GetEffectiveTemplatePermissions(projectID, templateID, userID)` on the store
  replaces the hand-rolled OR-ing in `ProjectMiddleware` and `GetTaskPermissionsMiddleware`.
- Filtering happens in SQL through an optional `TemplateViewer{UserID, ViewAll}` on the template,
  task, schedule and event queries; `nil` means no filtering for internal callers (scheduler,
  backup, runners, workflows). `ViewAll` is computed in Go from the request context so built-in
  roles stay code-defined and the SQL never joins the `role` table.
- API enforcement: a dedicated read-gate middleware on `/templates/{template_id}/...` (because
  `GetMustCanMiddleware` ignores GET), the task-scoped middleware extended to 403 on hidden tasks,
  and the viewer passed into every project-wide list (templates, view templates, tasks, last
  tasks, stats, schedules, events). Workflows in `pro_impl` hide a workflow that references a
  hidden template.
- Websocket and log streaming: `TaskRunner` keeps only users allowed to see the template in
  `t.users`, computed at task start from the DB on the node that runs the task — no cache, no
  shared state, so HA-safe.
- UI: bit 16 in `USER_PERMISSIONS`, checkbox in `EditRoleForm.vue` and
  `EditTemplatePermissionForm.vue` (which today hard-codes `[1, 2, 4, 8]`), chip in
  `TemplatePermissionsChips.vue`; lists come back already filtered so navigation needs nothing.

## Constraints

- Default-deny on every new read path: a missing role row or a store error is 403, never a
  fall-through to `next`.
- Filter in SQL, never in Go after fetching — task history uses keyset pagination (2.19) with a
  per-page `LIMIT`, so post-filtering would break page sizes.
- Never trust `permissions` sent by the client; API tokens go through the same chain.
- A role with `CanManageProjectResources` but without bit 16 is a misconfiguration (backup and
  export expose everything); the role form should force bit 16 on when bit 4 is checked.
- Admins bypass visibility, consistent with the rest of the API.

## Open questions

- ⚠️ 2026-09-27: the pre-existing bug the plan noted is still live — `GetTaskPermissionsMiddleware`
  (`api/projects/tasks.go:211-214`) writes 400 on a store error and still calls `next`. Fix it
  with or before this feature.
- Migration number: the plan said `v2.20.5`; the next free version is now `v2.20.8` or later.
- Events: hide task/template/schedule events by joining to the template; workflow events are
  simplest to hide entirely for restricted viewers — confirm that is acceptable.
- Guest keeps seeing everything (it is a built-in role with bit 16); hiding requires a custom
  role. Docs FAQ must say so.
- Out of scope, to be stated in docs: several roles per user or LDAP/OIDC group mapping (a member
  of two teams needs a combined role), views as "team folders", visibility of inventories, keys,
  repositories and environments.

## Key files and references

- `db/ProjectUser.go`, `db/Role.go`, `db/sql/template.go` (`getTemplates`,
  `GetTemplatePermission`), `db/sql/task.go` (`getTasks`), `db/sql/schedule.go`, `db/sql/event.go`
- `api/projects/project.go` (`ProjectMiddleware`, `GetMustCanMiddleware`),
  `api/projects/tasks.go` (`GetTaskPermissionsMiddleware`), `api/router.go`
  (`projectUserAPI`, `projectTmplManagement`, `projectTaskManagement`), `api/projects/views.go`
- `services/tasks/TaskRunner.go` (user list for websocket updates), `services/tasks/TaskRunner_logging.go`
- `web/src/lib/constants.js`, `web/src/components/EditRoleForm.vue`,
  `web/src/components/EditTemplatePermissionForm.vue`, `web/src/components/TemplatePermissionsChips.vue`
- Docs: `docs/docs/user-guide/team.md` ("Extended RBAC")
- Plan: `AGENTS/plans/2_20/template-visibility-rbac.md`
