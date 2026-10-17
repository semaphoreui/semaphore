# Global roles and permissions

Ported from `semaphore-ex` (its `db/Role_ex.go`, `db/sql/global_role.go`, `api/auth_ex.go`,
`test/edition-contract/enhanced/api/roles.go`) in task `TASK@8C634C1E95` (2026-10-17). A global role
carries a second bit mask, `global_permissions`; assigning the role to a user *globally* (table
`user__global_role`) grants those bits instance-wide. Effective permissions of a user = everything for
`user.admin`, otherwise the union of the assigned global roles.

## Decisions

- **Split as in `feat/group-role-mappings`.** Community owns the contract and the schema: types in
  `db/Role.go`, migration `v2.20.14`, the interface `db.GlobalRoleStore` (not part of `db.Store`), the
  stub `pro/db/sql/global_role.go` (no assignments, effective permissions 0), the factory
  `pro/db/factory.NewGlobalRoleStore`, the stub controller `pro/api/global_roles.go` (404), the
  middleware `globalPermissionMiddleware` (`api/global_permissions.go`) and the user handlers.
  `cli/cmd/root.go` builds the store and hands it to `api.Route`, which puts it on every authenticated
  request (`global_roles` context key). `pro_impl` owns the real store (`db/sql/global_role.go`), the
  factory and the `GlobalRolesController` (`api/global_roles.go`), plus the global bits in the role
  audit metadata and the `iam.global_role/unassign` events of a forced role delete.
- **Community edition = admin flag only.** Without the Pro store the effective mask is the admin flag;
  the community store still refuses to leave the instance without an administrator (its `UpdateUser`,
  `DeleteUser`, `UpdateRole`, `DeleteRoleWithAssignments` read `user__global_role` through the shared,
  exported helpers `LockGlobalRoleState` / `RequireAdministrator` / `GlobalRoleAssignmentSelect` that the
  Pro store reuses). The "Global roles" tab needs `features.custom_roles_management`.
- **No `revision` / string `role_id` from ex.** Roles keep `int id` + `slug`; the row lock
  `global_role_state` serializes administrator-removing transactions across HA nodes.
- **Only four bits** (1 users, 2 roles, 4 system, 8 audit): the other ex bits have no consumer here.
- **The admin flag is break-glass.** `user.admin` passes every check without a store read. Only an
  admin may set `admin`/`pro`; a user manager cannot touch an admin or a user holding a global
  permission the manager lacks (`userMayManageUser`), because a password reset would be a way up.
- Built-in roles cannot be assigned globally (they grant nothing); a project backup never carries
  `global_permissions` (`backup:"-"`), a store-to-store export does. `RoleImpact.members` counts global
  holders too, as in ex.
- The effective mask reaches the UI through `GET /api/user` (`global_permissions`), not system info.
- The Pro endpoints are not in `api-docs.yml` and have no Dredd fixtures, like `/role_mappings`.

## Route → permission

| Routes | Permission |
| --- | --- |
| `/options`, `/admin/info`, `/cache`, `/cluster*`, `/runners*`, `/runner_tags`, `/apps/{id}*`, `/tasks*` | manage system (4) |
| `/roles*`, `/users/{id}/global-roles*`, `/users/{id}/global-permissions` | manage roles (2) |
| `POST /users`, another user's `PUT/DELETE /users/{id}`, `/users/{id}/password`, full `GET /users` | manage users (1) |
| `/audit/events*` (`canReadAuditLog`) | read audit (8) |

## Open

- `pro_impl` work sits on branch `feat/features_from_ex` of the clone at `<worktree>/pro_impl`
  (from `origin/main` de0fc15); `feat/group-role-mappings` changes `DeleteRoleWithAssignments` and
  `RoleAssignments` too, so expect a merge in `api/roles.go` / `recordRemovedAssignments`.
- A Pro stand without a licence hides the "Global roles" tab (`custom_roles_management` is Enterprise);
  the API was checked with curl instead.
