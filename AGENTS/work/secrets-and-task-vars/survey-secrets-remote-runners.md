# Survey secrets on remote runners — as built

Survey variables of type `secret` used to arrive empty on remote runners and to survive only in
the memory of the node that accepted the task. They are now persisted as a task-bound, expiring
row in `access_key`, encrypted with the shared keyset, and decrypted on read by whichever HA node
dispatches the task. The decision and the rejected alternatives (task column, Redis, in-memory)
are in `decisions.md`; this doc records what shipped and where.

Status (2026-09-27): implemented — PR #4086 merged 2026-07-25 (commit 081425d2b, develop);
migration `db/sql/migrations/v2.20.1.sql` (`access_key.task_id` FK cascade, `access_key.expire_at`,
index `access_key__task_id`), registered in `db/Migration.go`. Not on `origin/2-19-stable` — the
v2.19.6 tag tree contains it only because that tag was cut from develop; v2.19.7 does not.

## How it works

- `TaskPool.AddTask`: after `CreateTask`, when the submitted secret JSON is a non-empty object
  (`hasSurveySecrets`), `CreateTaskSurveySecrets` stores it as one `AccessKeyString` row with
  owner `task`, `task_id` and `expire_at`. A store failure fails the task — never run with
  silently missing secrets. Plaintext is not retained in memory; both execution branches read
  back through the service.
- TTL (`taskSecretExpireAt`): `MaxTaskDurationSec` + 1 h queue allowance when the limit is set,
  else 24 h. No new config option. Expiry is a backstop; deletion on finish is the lifecycle.
- Expiry is enforced once, in `DeserializeSecret`
  (`services/server/access_key_encryption_svc.go:141`, `ErrAccessKeyExpired`), so every consumer
  is covered; `ExpireAt == nil` keeps all pre-existing keys non-expiring.
- Remote: `RunnerController.prepareRemoteJob` (`api/runners/runners.go`) decrypts on the node
  serving the poll and sets `JobData.Task.Secret`; expired or unreadable → task failed and
  finalised, not dispatched with empty vars. Runner side: `LocalExecutorProvider.NewExecutor`
  sets `Secret: task.Secret`.
- Local: `TaskRunner` reads through the same service before `Prepare` and sets
  `LocalExecutor.Secret` (`TaskRunner.go:255`) — also fixes secrets lost on a node restart while
  the task was queued.
- Cleanup: `DeleteTaskSurveySecrets` in the finish path (`TaskRunner.go:373`);
  `taskSecretSweepLoop` (started in `TaskPool`, once at boot then hourly, on every HA node,
  idempotent) removes expired rows; task deletion cascades through the FK.
- Leak surface: hidden from the secrets list by the `owner=''` filter; `api/projects/keys.go`
  refuses get/update/delete on task-owned keys (lines 44, 108); `services/project/backup.go:192`
  excludes them; `TaskID`/`ExpireAt` are `json:"-" backup:"-"`; `RekeyAccessKeys` re-encrypts
  them with every other key. Compatibility: an old runner ignores `task.secret`; a new runner on an
  old server sees it empty — as before.

## Open or deferred

- ⚠️ 2026-09-27: the Pro Docker/K8s executors still do not receive survey secrets. `New` in
  `pro_impl/services/tasks/docker/executor.go` and `k8s/executor.go` takes `Task`,
  `Environment` and `JWT` but has no `Secret` field, and `containerexec/ansible.go` builds
  `--extra-vars` from `Environment.Secrets` only. The "separate pro_impl PR" the plan promised has
  not landed on `pro_impl` `main`.
- User docs (`docs/docs/user-guide/task-templates/survey-vars.md`) describe the secret type but
  not the runner/HA lifecycle or the TTL; the encryption page does not mention task-owned keys.

## Key files and references

- `db/AccessKey.go` (`AccessKeyTaskSecret`, `TaskID`, `ExpireAt`), `db/Store.go`,
  `db/sql/access_key.go` (`GetTaskAccessKey`, `DeleteTaskAccessKeys`,
  `DeleteExpiredTaskAccessKeys`), `services/server/task_secret_svc.go`,
  `services/server/access_key_encryption_svc.go`, `services/tasks/TaskPool.go`,
  `services/tasks/TaskRunner.go`, `services/tasks/local_executor_provider.go`,
  `api/runners/runners.go`, `api/projects/keys.go`, `services/project/backup.go`.
- PR https://github.com/semaphoreui/semaphore/pull/4086. Research `RESEARCH@d8d16e3974`
  (per-job secret delivery in HA), `RESEARCH@6a7e8ecd1b` (key separation).
- Source plan: `AGENTS/plans/2_20/survey-secrets-remote-runners.md` (rev. 2).
