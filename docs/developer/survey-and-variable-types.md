# Survey variables and typed extra vars

## Survey variables on templates

Templates can define **survey variables** (`survey_vars` on `db.Template`) that
prompt the user at task launch. Each variable has a `type`:

| `type` | UI control | Validation |
| --- | --- | --- |
| *(empty / str)* | Text field | Required check only |
| `int` | Text field | Must match `/^\d+$/` when non-empty |
| `enum` | Select from `values[]` | Required selects must have a choice |
| `text` | Multiline textarea | Required check only |
| `secret` | Masked text field | Stored as encrypted task-bound access key (see below) |

Each variable also has an optional `target`:

| `target` | Effect |
| --- | --- |
| *(empty)* | App-specific default (CLI args for Ansible, env for other apps) |
| `env` | Passed as a process environment variable |

**Code:** `db/Template.go` (`SurveyVar`, `SurveyVarType`), `web/src/components/SurveyVars.vue`,
`web/src/components/TaskParamsForm.vue`, `web/src/components/TaskForm.vue`.

Integer survey vars are stored and passed as strings in the task environment;
the UI enforces numeric input before submit.

### Default values

`default_value` on `SurveyVar` pre-fills the launch form. Scheduled tasks and
workflow nodes can also supply defaults via their `task_params.environment` JSON.

## Variable group extra vars (typed editor)

Variable groups (`db.Environment`) store `json` extra variables. The UI table
editor (`EnvironmentForm.vue`) assigns each row a `type`:

| Editor `type` | JSON output |
| --- | --- |
| `string` | string scalar |
| `number` | parsed number |
| `list` | JSON array (user enters `["a","b"]`) |
| `dict` | JSON object (user enters `{"k":"v"}`) |

`inferVarType()` maps existing JSON back to the correct editor type when loading
a group. `rowToVarValue()` parses user input and throws descriptive errors on
invalid JSON or numbers.

Switching between **JSON** and **table** edit modes preserves structure when the
JSON is compatible with the typed rows.

## Task and workflow params

`db.TaskParams` holds per-run overrides used by:

- Manual task launch (`TaskForm.vue`)
- Schedules (`ScheduleForm.vue`)
- Workflow task nodes (`WorkflowEditor.vue` → `TaskParamsForm.vue`)

Fields include `environment` (survey answers JSON), `git_branch`, `inventory_id`,
`message`, `arguments`, and app-specific `params` (Ansible limits/tags, etc.).
`TaskParams.CreateTask()` copies them into a `db.Task`.

Workflow nodes persist params via `task_params_id` → `project__task_params`
(migration `v2.19.11.sql`). Each task node can pin survey answers and overrides
independent of the template defaults.

## Secret survey variables

Survey variables with `type: secret` are **not** stored in the task row, event log,
or API responses. At task creation (`TaskPool.AddTask`), their JSON object is
persisted as an encrypted access key:

| Property | Value |
| --- | --- |
| `owner` | `task` (`AccessKeyTaskSecret`) |
| `task_id` | Bound to the creating task |
| `expire_at` | `max_task_duration_sec` + 1 hour when set; otherwise 24 hours |

This design lets any HA node decrypt secrets at dispatch time (local or remote
runner) without passing plaintext through in-memory task state.

**Lifecycle**

1. **Create** — `CreateTaskSurveySecrets` encrypts and inserts the key
   (`services/server/task_secret_svc.go`).
2. **Dispatch** — local executors and `prepareRemoteJob` call
   `GetTaskSurveySecrets` and populate `JobData.task.secret` / executor env.
3. **Finish** — `TaskRunner` deletes the key when the task reaches a terminal
   state (`DeleteTaskSurveySecrets`).
4. **Sweep** — `taskSecretSweepLoop` runs hourly to remove expired keys left
   behind when a node crashes before cleanup.

Task-bound keys are hidden from project key APIs (`404` on direct access) and
excluded from project backups (`services/project/backup.go`).

## Secret-backed variable groups

When a variable group links to a secret storage (`secret_storage_id`), extra vars
and secrets can be synced from Vault/OpenBao/Enterprise backends. See
[secret-storages.md](secret-storages.md).

## Related code

- Types: `db/Template.go`, `db/TaskParams.go`, `db/Environment.go`
- UI: `web/src/components/EnvironmentForm.vue`, `web/src/components/TaskParamsForm.vue`
- Task-bound survey secrets: `services/server/task_secret_svc.go`, `services/tasks/TaskPool.go`
- Encryption / external secrets: `services/server/access_key_encryption_svc.go`
