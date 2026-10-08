# work — project documentation by zone

What the project's subsystems do, why they were decided that way, and what is still broken. One
folder per zone, each with its own `INDEX.md`. Project-specific — unlike `docs/`, this folder does
not migrate to a new project.

## Zone index

| When you need it | Read |
| --- | --- |
| Changing login, LDAP/OIDC providers, external identity linking, password hashing, task JWT tokens, or per-template RBAC | `auth-and-identity/INDEX.md` |
| Touching runner liveness and lost tasks, runner tokens, the runners page, git working-copy locking, or host-config SSH credentials | `runners-and-execution/INDEX.md` |
| Touching task environment/survey variables, survey secrets on remote runners, survey defaults in schedules, or secret storages (CyberArk) | `secrets-and-task-vars/INDEX.md` |
| Changing how encryption keys are sourced, identified, rotated or verified (`vault rekey` / `check`) | `encryption-keys/INDEX.md` |
| Returning to the project-alerts PR #4272 or its un-pushed docs branch | `alerts/INDEX.md` |
| Changing the workflow editor canvas (Drawflow visual layer, wheel, zoom), workflow artifacts, or planning Workflows-as-Code | `workflow-editor/INDEX.md` |
| Touching user options (pinned menu), task history pagination, or the contextual help panel idea | `ui-features/INDEX.md` |
| Changing cron schedules: syntax, day offsets, timezone, survey defaults for scheduled tasks | `schedules/INDEX.md` |
| Adding debug logging or touching the debug filter | `logging/INDEX.md` |
| Touching the task pool, output streaming, runner polling, DB indexes, or planning a load test | `performance/INDEX.md` |
