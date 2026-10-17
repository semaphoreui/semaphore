# Global roles and permissions

Ported from `semaphore-ex` (its `db/Role_ex.go`, `db/sql/global_role.go`, `api/auth_ex.go`,
`test/edition-contract/enhanced/api/roles.go`) in task `TASK@8C634C1E95` (2026-10-17). A global role
carries a second bit mask, `global_permissions`; assigning the role to a user *globally* (table
`user__global_role`) grants those bits instance-wide. Effective permissions of a user = everything for
`user.admin`, otherwise the union of the assigned global roles.

## Decisions

- **Community side holds everything but role CRUD.** Model, migration (`v2.20.14`), store
  (`db/sql/global_role.go`), `globalPermissionMiddleware` (`api/global_permissions.go`), the
  assignment and effective-permission handlers (`api/global_roles.go`), UI and docs are here. Role
  CRUD stays in the private `pro_impl`; it needs no change for the new field because
  `helpers.Bind(&db.Role)`, `db.ValidateRole` and the store already carry `global_permissions`.
- **No `revision` / string `role_id` from ex.** Roles keep `int id` + `slug`. Instead of optimistic
  locking, every transaction that can remove an administrator first updates the single row of
  `global_role_state` (`lockGlobalRoleState`), which serializes those transactions across HA nodes.
- **Only four bits** (1 users, 2 roles, 4 system, 8 audit): the other ex bits have no consumer here.
- **The admin flag is break-glass.** `user.admin` passes every check without a store read. Only an
  admin may set `admin`/`pro`; a user manager cannot touch an admin or a user holding a global
  permission the manager lacks (`userMayManageUser`), because a password reset would be a way up.
- **Last administrator invariant.** `requireAdministratorTx` refuses a change that would leave no
  user with `admin` or `CanManageGlobalRoles` (delete/demote user, remove assignment, delete role or
  strip the bit). An instance with zero such users is not protected (fresh DB, tests).
- Built-in roles cannot be assigned globally (they grant nothing); a project backup never carries
  `global_permissions` (`backup:"-"`), a store-to-store export does.
- The effective mask reaches the UI through `GET /api/user` (`global_permissions`), not system info.

## Route → permission

| Routes | Permission |
| --- | --- |
| `/options`, `/admin/info`, `/cache`, `/cluster*`, `/runners*`, `/runner_tags`, `/apps/{id}*`, `/tasks*` | manage system (4) |
| `/roles*`, `/users/{id}/global-roles*`, `/users/{id}/global-permissions` | manage roles (2) |
| `POST /users`, another user's `PUT/DELETE /users/{id}`, `/users/{id}/password`, full `GET /users` | manage users (1) |
| `/audit/events*` (`canReadAuditLog`) | read audit (8) |

## Open

- `pro_impl/api/roles.go` is behind this branch's `RoleRepository` (`ValidateRole(role, isNew)`,
  `DeleteRole(id)`) and could add `global_permissions` to `audit.RoleMetadata`.
- The Dredd fixtures (`global_role` capability, `assignment_id` x-example 23) were only compiled, not
  run locally.
