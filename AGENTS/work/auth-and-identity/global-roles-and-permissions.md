# Global roles and permissions

Ported from `semaphore-ex` (its `db/Role_ex.go`, `db/sql/global_role.go`, `api/auth_ex.go`,
`test/edition-contract/enhanced/api/roles.go`) in task `TASK@8C634C1E95` (2026-10-17), then split into its
own table in `TASK@EE7CBE6E1F` (2026-10-17). A **global role** (`global_role`: id, slug, name,
`permissions` = `db.GlobalPermission` mask, Go type `db.GlobalRole`) grants instance-wide permissions to the
users it is assigned to (`user__global_role`, `role_id` → `global_role.id`). A **project role** (`role`:
built-in, shared = `project_id is null`, or per project) grants project permissions only; the two never mix
in one row. Effective permissions of a user = everything for `user.admin`, otherwise the union of the
assigned global roles.

## Decisions

- **Split as in `feat/group-role-mappings`.** Community owns the contract and the schema: types in
  `db/Role.go`, migration `v2.20.16` (`global_role`, the `global_role_state` lock row and
  `user__global_role`; the interim 2.20.14/2.20.15 pair that first put a column on `role` was folded into
  it before release), the interface `db.GlobalRoleStore` (not part of `db.Store`: global role CRUD +
  assignments), the stub `pro/db/sql/global_role.go` (no roles, no assignments, effective permissions 0),
  the factory `pro/db/factory.NewGlobalRoleStore`, the stub controller `pro/api/global_roles.go` (empty
  lists, 404 on writes), the middleware `globalPermissionMiddleware` (`api/global_permissions.go`) and the
  user handlers. `cli/cmd/root.go` builds the store and hands it to `api.Route`, which puts it on every
  authenticated request (`global_roles` context key). `pro_impl` owns the real store
  (`db/sql/global_role.go`, wrapping `*coresql.SqlDb`) and the `GlobalRolesController`
  (`api/global_roles.go`: `/global-roles*`, `/users/{id}/global-roles*`, `/users/{id}/global-permissions`),
  with `iam.global_role/create|update|delete` (target `global_role`, `RoleMetadata` /
  `RoleDeleteMetadata`) and `assign|unassign` events.
- **Names.** "Global role" means only the new entity. The project-level roles with `project_id is null`
  are **shared roles** in code (`GetSharedRoles`, `GetSharedRoleBySlug`, `GetProjectOrSharedRoleBySlug`,
  `GetProjectAndSharedRoles`, UI scope `shared`) and stay on the Roles page. Global roles are managed on
  the "Global roles" tab of the Users page (`GlobalRolesList.vue`, owner's choice 2026-10-17) at its own
  address `/users/roles` (same `Users.vue`, the tab is a computed over `$route.path`; 2026-10-19) and are
  addressed by id in the API (`/global-roles/{role_id}`), project roles by slug.
- Global and project slugs are independent (a global role may share a slug with a project role).
  Store-to-store export (`services/export`) does not carry global roles or their assignments. A dev
  database that ran the interim 2.20.14/2.20.15 migrations must be recreated: 2.20.16 creates the same
  tables and is not applied on top of them.
- **Community edition = admin flag only** (tightened 2026-10-19, `TASK@7C74934683`). The community store has no
  global-role SQL: `db/sql/administrator.go` holds only `LockAdministratorState` (the `global_role_state` row
  lock) and an admin-flag count, so `UpdateUser` / `DeleteUser` refuse to remove the **last admin-flag user**
  whatever global roles exist. The Pro store (`pro_impl/db/sql/global_role.go`) takes the same lock and owns
  `globalRoleAssignmentSelect`, `countAdministrators` (admins + holders of `CanManageGlobalRoles`) and
  `requireAdministrator`, used by `UpdateGlobalRole`, `DeleteGlobalRoleWithAssignments` and
  `DeleteGlobalRoleAssignment`; with the community rule in place they only matter on an instance that has no
  admin-flag user. Project role updates and deletes do not take the lock. The Roles page, the Users page
  "Global roles" tab and the user's "Roles" tab exist only in a Pro build (`isPro`) and need
  `features.global_roles_management` — a **Pro** feature of **Pro users** (`GlobalRolesManagement:
  planDetails.IsPro() && user.Pro || planDetails.IsEnterprise()`, 2026-10-19), unlike the Enterprise
  `custom_roles_management` that gates shared/project roles. The Pro `GlobalRolesController`
  answers 403 on every route but the catalog when the plan lacks it, so Dredd skips those transactions
  unless `globalRolesEnabled()`.
- **No `revision` / string `role_id` from ex.** Roles keep `int id` + `slug`; the row lock
  `global_role_state` serializes administrator-removing transactions across HA nodes.
- **Only four bits** (1 users, 2 roles, 4 system, 8 audit): the other ex bits have no consumer here.
- **The middleware is the only gate.** A handler behind `systemAPI` / `rolesAdminAPI` never re-checks
  `user.Admin`: such a check silently shuts out every global-role holder (`api/cache.go` and `api/options.go`
  carried one until 2026-10-19). `api/global_routes_test.go` drives the table below through the real
  `api.Route` with a session cookie, for an admin, one holder per bit, one holder of every other bit and a
  user without roles; a second test walks the router and fails when a route that refuses the plain user
  with a global permission name is missing from the table. Add a row there with every new system route.
- **The admin flag is break-glass.** `user.admin` passes every check without a store read, and one admin-flag
  user always remains (see above). Only an
  admin may set `admin`/`pro`; a user manager cannot touch an admin or a user holding a global
  permission the manager lacks (`userMayManageUser`), because a password reset would be a way up.
- Only global roles can be assigned globally; built-in, shared and project roles belong to project
  membership. A project backup never sees global roles. `RoleImpact.members` of a global role counts its
  holders; of a project role, its project members only.
- The effective mask reaches the UI through `GET /api/user` (`global_permissions`), not system info.
- Unlike `/roles` CRUD, the global role endpoints are in `api-docs.yml` with Dredd fixtures (capability
  `global_role`, path example 24 = global role id, 23 = assignment id) registered only under
  `isProBuild()` and listed in `proOnlyTests`, so a community build skips them.

## Route → permission

Kept in step with `globalRoutes()` in `api/global_routes_test.go`.

| Routes | Permission |
| --- | --- |
| `/options`, `/admin/info`, `/cache`, `/cluster*`, `/runners*`, `/runner_tags`, `/apps/{id}*`, `/tasks*` | manage system (4) |
| `/roles*`, `/global-roles*`, `/users/{id}/global-roles*`, `/users/{id}/global-permissions` | manage roles (2) |
| `POST /users`, another user's `PUT/DELETE /users/{id}`, `/users/{id}/password`, full `GET /users` | manage users (1) |
| `/audit/events*` (`canReadAuditLog`) | read audit (8) |

## Open

- `pro_impl` work sits on branch `feat/features_from_ex` of the clone at `<worktree>/pro_impl`
  (from `origin/main` de0fc15); `feat/group-role-mappings` changes `DeleteRoleWithAssignments` and
  `RoleAssignments` too, so expect a merge in `api/roles.go` / `recordRemovedAssignments`.
- A Pro stand without a licence hides the "Global roles" tab (`custom_roles_management` is Enterprise);
  the API was checked with curl instead.
