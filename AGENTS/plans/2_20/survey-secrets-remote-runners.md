# Plan — Survey Secret Variables Lost on Remote Runners (HA-safe fix)

- **Branch:** `develop`
- **Status:** implemented (rev. 2 — task-bound access keys; the `task.secret`
  column and Redis approaches were rejected). TTL decided: derived from
  `MaxTaskDurationSec` (+1h queue allowance; 24h when unlimited). Note: the
  pro_impl docker/k8s providers still need the `Secret: task.Secret` wiring
  (separate PR).

## 1. Problem

Survey variables of type `secret` work when a task is executed by the server
itself, but arrive **empty** on remote runners.

Additionally (same root cause), the secret survives only in the memory of the
node that accepted the task: it does not survive a server restart while the
task is queued, and in HA mode no other node can ever dispatch the task with
its secrets.

## 2. Root Cause

Survey secrets travel in `db.Task.Secret` (`db:"-" json:"secret,omitempty"`,
`db/Task.go:52`) — never persisted and blanked before serialization:

```go
// services/tasks/TaskPool.go:971 (AddTask)
extraSecretVars := taskObj.Secret
taskObj.Secret = "{}"
```

- **Local branch** (`TaskPool.go:1048`): `extraSecretVars` is threaded into
  `LocalExecutor{Secret: extraSecretVars}` → works, but only in the accepting
  node's memory.
- **Remote branch** (`TaskPool.go:1036`): `RemoteJob{Task: taskRunner.Task}` —
  the secret is already `"{}"` and `extraSecretVars` is dropped. The runner
  poll payload (`api/runners/runners.go:121`, `JobData{Task: tsk.Task}`) ships
  `"secret": "{}"`.
- **Second gap, runner side**: even if `Task.Secret` arrived, the runner never
  uses it — `LocalExecutorProvider.NewExecutor`
  (`services/tasks/local_executor_provider.go:37`) does not populate
  `LocalExecutor.Secret`, and `mergeExtraVars` reads only that field
  (`local_executor.go:159`), not `t.Task.Secret`.

## 3. Design

Industry consensus (research above; AWX is the closest analog): per-run
secrets are persisted **encrypted at rest in the shared DB** with a key held
in config identically on every node, decrypted on read by whichever node
serves the worker request, hidden from user-facing surfaces.

**Chosen approach: store survey secrets as a task-bound `access_key` row** —
the table that already holds every other secret, already encrypted with the
shared keyring (`util.Config.EncryptAccessSecret`, key-id envelope, rotation
support) and already covered by the vault/rekey tooling.

Principles:

1. The secret is stored in `access_key` like any other secret (keyring
   encryption via the existing `SerializeSecret` path).
2. The key is **bound to the task** (new `task_id` column) and carries a new
   owner value, so it is **invisible in the project secrets list** (the list
   endpoint filters `owner=''` — `db/sql/access_key.go:28-29`,
   `api/projects/keys.go:69` — non-empty owners are excluded automatically,
   same as `environment`/`variable`/`vault` owners today).
3. New **`expire_at`** column limits the secret's lifetime.
4. **Expired secrets cannot be used**: enforced centrally at deserialization.

Rejected alternatives: extra column on `task` (rejected by maintainer);
encrypted values in Redis (area `AREA@d489ed1d499f99bd81a192`: acceptable only
with hardening, but Redis can silently lose the value and joins the dispatch
critical path).

## 4. Implementation

### 4.1 DB migration — `db/sql/migrations/v2.20.1.sql`

```sql
alter table `access_key` add `task_id` int null references task(`id`) on delete cascade;
alter table `access_key` add `expire_at` datetime null;
create index `access_key__task_id` on `access_key`(`task_id`);
```

Register `{Version: "2.20.1"}` in `db/Migration.go` (after `2.20.0`, line 134).

`on delete cascade` ties the secret's lifetime to the task row: task retention
cleanup (`MaxTasksPerTemplate`) deletes the key automatically.

### 4.2 Model (`db/AccessKey.go`)

```go
const AccessKeyTaskSecret AccessKeyOwner = "task"

// on AccessKey:
TaskID   *int       `db:"task_id" json:"-" backup:"-"`
ExpireAt *time.Time `db:"expire_at" json:"-" backup:"-"`
```

- Add both columns to `CreateAccessKey` inserts (`db/sql/access_key.go`).
- `GetAccessKeys`: add `case db.AccessKeyTaskSecret` filtering by
  `pe.task_id` (mirror of the `environment_id` cases).
- New store method `GetTaskAccessKey(projectID, taskID int) (AccessKey, error)`
  (lookup by `owner='task' and task_id=?`).
- New store method `DeleteTaskAccessKeys(projectID, taskID int) error`.

### 4.3 Expiry enforcement (generic, central)

In `accessKeyEncryptionServiceImpl.DeserializeSecret`
(`services/server/access_key_encryption_svc.go:128`):

```go
if key.ExpireAt != nil && tz.Now().After(*key.ExpireAt) {
    return ErrAccessKeyExpired // "access key expired"
}
```

One check point covers every consumer (local executor, runner dispatch,
anything future). `ExpireAt == nil` → no expiry (all existing keys). The field
is generic on purpose: any access key can later be given a TTL.

### 4.4 Creating the secret — `TaskPool.AddTask` (`TaskPool.go:960`)

After `p.store.CreateTask` succeeds and if the submitted secret is a
non-empty JSON object:

```go
key := db.AccessKey{
    Name:      fmt.Sprintf("task-%d-survey-secrets", newTask.ID),
    Type:      db.AccessKeyString,
    ProjectID: &projectID,
    Owner:     db.AccessKeyTaskSecret,
    TaskID:    &newTask.ID,
    ExpireAt:  &expireAt,
    String:    extraSecretVars, // whole survey-secrets JSON as one string secret
}
// SerializeSecret (keyring-encrypt) + CreateAccessKey via AccessKeyEncryptionService
```

- `expireAt = now + TTL` (decided: derive from `MaxTaskDurationSec`, no new
  config option): TTL = `util.Config.MaxTaskDurationSec` + 1h queue allowance
  when set; 24h when `MaxTaskDurationSec == 0` (expire_at is always set —
  expiry is a hard requirement; deletion on finalize remains the primary
  lifecycle, expiry is the backstop).
- Failure to create the key fails the task (never run with silently-missing
  secrets).
- `taskObj.Secret = "{}"` stays as today; plaintext no longer retained in
  memory at all — both execution branches read back through the service.

### 4.5 Reading the secret at dispatch

New service method (e.g. on `AccessKeyEncryptionService` or a small
`TaskSecretService`): `GetTaskSurveySecrets(projectID, taskID) (string, error)`
— loads via `GetTaskAccessKey`, calls `DeserializeSecret` (which enforces
`expire_at`), returns the plaintext JSON. `db.ErrNotFound` → task simply has
no survey secrets (normal case) → `""`.

Consumers:

- **Remote**: runner poll handler (`api/runners/runners.go`, where `JobData`
  is built): set `jobData.Task.Secret` from the service. Works on any HA node
  — ciphertext in the shared DB, keyring in config. Same trust channel that
  already carries decrypted access keys and vault passwords (TLS + runner
  token). On `ErrAccessKeyExpired` or decrypt failure: fail the task with a
  clear log/status message, do not dispatch with empty vars.
- **Local**: in `AddTask`'s local branch (and/or `LocalExecutor.Prepare`),
  populate `LocalExecutor.Secret` via the same service call instead of the
  in-memory `extraSecretVars`. This also fixes local secrets lost on node
  restart while queued.

### 4.6 Runner-side consumption

`LocalExecutorProvider.NewExecutor`
(`services/tasks/local_executor_provider.go:37`): add `Secret: task.Secret`.
Mirror in pro docker/k8s providers (pro_impl, separate PR).

Compatibility: old runner + new server deserializes `task.secret` and ignores
it (same behavior as today, no error); new runner + old server sees an empty
field. Fully backward compatible.

### 4.7 Cleanup

- **On terminal status**: call `DeleteTaskAccessKeys` from the finish path
  (`TaskRunner.finishRun`), covering both local and remote
  (`finalizeRemoteTaskLocked` → `finishRun`) completion; HA orphan cleaner as
  backstop where it force-fails stale tasks.
- **Expired sweep**: periodic cleanup deleting `access_key` rows where
  `owner='task' and expire_at < now` (attach to an existing periodic job,
  e.g. the task-pool/orphan maintenance loop). Idempotent — safe if several
  HA nodes race.
- **Task deletion/retention**: covered by `on delete cascade`.

### 4.8 Guard rails / leak surface

- Secret value never serialized: `AccessKey.Secret` is already `json:"-"`;
  new `TaskID`/`ExpireAt` also `json:"-" backup:"-"`.
- Secrets list: hidden automatically by the `owner=''` filter (verified,
  `db/sql/access_key.go:28`).
- Generic key endpoints (`api/projects/keys.go` get-by-id/update/delete):
  reject keys with `Owner == AccessKeyTaskSecret` (mirror whatever guard
  applies to other owned keys; verify during implementation — metadata-only
  exposure today, but mutations must be blocked).
- Project backup (`services/project/backup.go`): verify task-owned keys are
  excluded (exclude non-shared owners if not already).
- Rekey: `RekeyAccessKeys` iterates with `IgnoreOwner: true`
  (`access_key_encryption_svc.go:208`) → task secrets are re-encrypted on key
  rotation for free.
- Never log the decrypted value.

## 5. Tests

Per `.claude/CLAUDE.md` rules (`testify`, table-driven):

1. `AddTask` with survey secrets → `access_key` row created: `owner='task'`,
   `task_id` set, `expire_at` set, ciphertext decrypts to the submitted JSON;
   empty/`"{}"` secret → no row.
2. `DeserializeSecret` with `ExpireAt` in the past → `ErrAccessKeyExpired`;
   nil / future → OK (table-driven).
3. Runner poll handler (httptest): task with secret key →
   `new_jobs[0].task.secret` contains decrypted JSON; expired key → task
   failed, no job dispatched.
4. `LocalExecutorProvider.NewExecutor` populates `Secret`; `mergeExtraVars`
   merges it (extend `local_executor_test.go`).
5. Cleanup: after finalize the key row is gone; expired sweep removes stale
   rows; project secrets list API does not contain the task key.

## 6. Files Touched (summary)

| File | Change |
|------|--------|
| `db/sql/migrations/v2.20.1.sql` | `task_id`, `expire_at` columns + index |
| `db/Migration.go` | register version |
| `db/AccessKey.go` | `AccessKeyTaskSecret` owner, `TaskID`, `ExpireAt` |
| `db/Store.go` + `db/sql/access_key.go` | insert columns, owner filter case, `GetTaskAccessKey`, `DeleteTaskAccessKeys` |
| `services/server/access_key_encryption_svc.go` | expiry check in `DeserializeSecret`; `GetTaskSurveySecrets` |
| `services/tasks/TaskPool.go` | create task key in `AddTask`; read via service for local branch |
| `services/tasks/TaskRunner.go` | delete key in `finishRun` |
| `api/runners/runners.go` | inject decrypted secret into `JobData` |
| `api/projects/keys.go` | block mutations on task-owned keys |
| `services/project/backup.go` | verify/exclude task-owned keys |
| `services/tasks/local_executor_provider.go` | populate `Secret` |
| pro_impl docker/k8s providers | populate `Secret` (separate PR) |
