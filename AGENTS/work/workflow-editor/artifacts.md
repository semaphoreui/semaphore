# Workflow artifacts (set_stats parity)

AWX-style hand-off of key/value data between task nodes of one workflow run. Design: a task gets
`SEMAPHORE_ARTIFACTS_FILE` pointing at a per-task JSON file; Ansible fills it through an
embedded aggregate callback plugin (`semaphore_artifacts`) from `ansible.builtin.set_stats`, any
other app writes the JSON itself. After exit the server validates the file and stores it on
`task.artifacts`; when a downstream node starts, artifacts of every finished upstream task are
merged (later task id wins, the current task excluded) and injected as top-level Ansible extra
vars, under `semaphore_workflow_artifacts`, and as `SEMAPHORE_WF_<KEY>` env for scalars. Before
this the only cross-node value was the Build template `version` (`incoming_version`).

**Status (2026-09-27): partially implemented — storage, merge and read API exist; the
producer and consumer wiring is absent on this branch.** Evidence:

- Present: `task.artifacts` column (`db/sql/migrations/v2.18.15.sql:92`), `db.Task.Artifacts`,
  `Store.UpdateTaskArtifacts` (`db/sql/task.go`), `GET
  /project/{p}/workflows/{w}/runs/{r}/artifacts` (`api/router.go:512`, `api-docs.yml:3401`) →
  `WorkflowService.GetWorkflowRunArtifacts` → `artifacts.CollectFromTasks`
  (`pro_impl/services/server/workflow_svc.go`). The package
  `pro_impl/services/tasks/artifacts/` carries `LoadFile`, `Parse`, `Merge`, `ToShellEnv`,
  `AnsibleCallbackEnv`, the embedded callback plugin and `artifacts_test.go`. UI: the task
  dialog shows a pretty-printed `task.artifacts` panel (`web/src/components/TaskDetails.vue`),
  the run view shows only the remote-runner warning (`WorkflowRun.vue`).
- ⚠️ 2026-09-27: nothing calls `UpdateTaskArtifacts`, `LoadFile`, `ToShellEnv`,
  `AnsibleCallbackEnv` or `TaskPool.GetWorkflowRunArtifacts` (`grep -rn` over the open module and
  `pro_impl`, tests excluded); `LocalExecutor.WorkflowArtifacts`
  (`services/tasks/local_executor.go:57`) has no reader; the body of
  `pro_impl/services/tasks/LocalJob_artifacts_test.go` is commented out. So no task receives the
  file path or the callback env, no task row is ever filled, and no downstream task sees
  anything — the endpoint always returns `{}`. The Copilot commits that had the wiring
  (`47f725321`, `58d11c192`, PR #3488; re-imported as `30bb3e194`) are not ancestors of HEAD;
  `f29bb3f40 Workflows` (on HEAD) brought only the plan. The producer side was lost when
  `LocalJob.go` became `local_executor.go` and the engine moved to `pro_impl`.
- ⚠️ 2026-09-27: `docs/docs/user-guide/workflows.md` § "Workflow artifacts (set_stats)" tells
  users the feature works for local tasks. Until the wiring is restored the doc overclaims.

## Decisions and why

- **File hand-off via env var, not stdout parsing** — app-agnostic (Bash, Python, Terraform
  write the JSON themselves); Ansible needs no playbook change thanks to the callback plugin.
- **Three projections at once** (flat extra vars, namespaced key, `SEMAPHORE_WF_*` env) — flat
  vars give AWX parity for existing playbooks, the namespace serves explicit callers, env serves
  shell/Terraform; only scalars become env, nested values stay in extra vars.
- **Later task overrides earlier**, current task excluded — mirrors AWX `set_stats` merge and
  stops a running task from feeding itself.
- **Per-task text column, merge on read** — no run-level blob to keep consistent; the merged map
  is derived, so HA nodes need no coordination.
- **Validation at ingest**: JSON object only, keys `^[A-Za-z_][A-Za-z0-9_]*$`, 256 KB max,
  reserved keys `semaphore_vars`, `semaphore_workflow_artifacts`, `task_details`,
  `incoming_version` refused — prevents clobbering Semaphore-injected variables.
- **Pro-gated with the engine**: the package and merge live in `pro_impl`; the open stub returns
  an empty map so `TaskRunner` can call it unconditionally.

## Open, broken, deferred

- Restore the producer/consumer wiring (env var + callback env on launch, read + validate +
  `UpdateTaskArtifacts` on finish, merge into extra vars/env on downstream launch), on the Pro
  side; then un-comment or rewrite `LocalJob_artifacts_test.go`.
- Remote runners (`useRemoteRunner: true`) have no artifact channel in the runner→server
  protocol; the docs warning and `workflowArtifactsRemoteRunnerWarning` alert stay until then.
- Run-level artifacts panel was removed with the run-view table (see
  `editor-architecture.md`); only the per-task panel remains.
- Automatic capture of `terraform output -json` as a sub-key — never started.
- Dredd skips the artifacts endpoint (`.dredd/hooks/main.go:179`), so the API contract is not
  exercised in CI.

## Key files and references

- `db/Task.go`, `db/Store.go`, `db/sql/task.go`, `db/sql/migrations/v2.18.15.sql`
- `services/tasks/TaskPool.go` (`GetWorkflowRunArtifacts` delegator),
  `services/tasks/local_executor.go` (`WorkflowArtifacts` field)
- `pro_interfaces/workflow_svc.go`, `pro/services/server/workflow_svc.go` (stub),
  `pro_impl/services/server/workflow_svc.go`, `pro_impl/services/tasks/artifacts/`
  (incl. `ansible/callback_plugins/semaphore_artifacts.py`)
- `web/src/components/TaskDetails.vue`, `web/src/views/project/WorkflowRun.vue`
- `api-docs.yml` (artifacts path), `docs/docs/user-guide/workflows.md`
- https://github.com/semaphoreui/semaphore/pull/3488 (origin of the feature)
