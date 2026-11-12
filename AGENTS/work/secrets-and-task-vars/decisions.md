# Task variables and secrets — decisions

## Task environment variables are default-deny (2026-07-30)

An external researcher reported (received 2026-07-29, embargoed, publication planned around
2026-10-27) that a `task_runner` could pass arbitrary keys in a task's `environment`, which
`populateTaskEnvironment` merges over the template environment and which reach Ansible as
`--extra-vars`; `ansible_ssh_common_args` + `ProxyCommand` then runs commands on the Semaphore
server. `Task.Secret` was the same vector (merged in `getEnvironmentExtraVars`).

Decision (Denis): default-deny for every caller of the task API — no role or admin carve-out — with
an explicit per-template opt-in `allow_any_vars_in_task` (migration v2.20.2, checkbox under
Advanced options in the template form). A permission-based variant (allow undeclared vars for
`CanManageProjectResources`) was rejected.

The gate is `db.Task.ValidateSurveyVars` (returns early when the template opts in), called from
`TaskController.AddTask`. Integrations, schedules and workflow-node tasks call `TaskPool.AddTask`
directly and are deliberately not validated — their variable names are manager-configured, and
validating them would break existing integration extract values.

⚠️ 2026-09-28: the gate exists only on branch `feat/survey_vars_validation` (commit 75ea78caf,
2026-07-30) — not on `develop`, where `v2.20.2.sql` is the workflow-delay migration. Renumber the
migration when merging.

## Survey secrets on remote runners are task-bound access keys (2026-07-23)

Survey `secret` variables were dropped for remote-runner tasks: `TaskPool.AddTask` blanked
`taskObj.Secret` and threaded the value only into `LocalExecutor`; the runner poll payload got
`"{}"`, and runner-side `LocalExecutorProvider.NewExecutor` never set the executor `Secret`.

Rejected: an in-memory fix (must be HA), a new `task.secret` column, Redis storage. Chosen: store
the survey-secret JSON as a task-bound row in the existing `access_key` table — new owner `task`
(hidden from the secrets list by the `owner=''` filter), `access_key.task_id` (FK cascade) and
generic `access_key.expire_at`; expiry enforced centrally in `DeserializeSecret`; decrypt-on-read
in the runner poll handler on any HA node; delete on task finalize plus an expired sweep; populate
executor `Secret` on the runner side. As built, TTL derivation, cleanup and the pro_impl gap:
`survey-secrets-remote-runners.md`. Research: `RESEARCH@d8d16e3974`; related key-separation
research `RESEARCH@6a7e8ecd1b`.

## CyberArk PVWA secret storage (2026-09-25)

Implemented on branch `feat/cyberark-secret-storage` in three repos at once: root (from develop),
`pro_impl` (from the sem-272-host-config-page HEAD, two unpushed commits ahead of main) and the
`docs` submodule (from main), mirroring the Vault/Azure storage pattern. Design notes on that
branch: `AGENTS/plans/2_20/cyberark-secret-storage.md`.

⚠️ 2026-09-25: never tested against a real PVWA, only against the httptest fake in
`pro_impl/services/server/cyberark_client_test.go`. Verify pagination (`nextLink`) and
`Password/Retrieve` for key accounts against a real instance before touching it again.
