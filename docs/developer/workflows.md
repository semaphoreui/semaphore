# Workflows

## Purpose

**Workflows** chain task templates into a directed graph: each node runs a
template (or pauses for human approval), and edges define when downstream nodes
start based on upstream outcomes. They mirror CI/CD pipeline concepts (Concourse
/ AWX workflow jobs) inside a Semaphore project.

Workflows are a **Pro / Enterprise** feature. The open-source tree compiles the
full API surface and database schema, but:

- `pro/api/projects/workflows.go` and `pro/services/server/workflow_svc.go` are
  no-op stubs that return empty lists or `404`.
- Real orchestration lives in `pro_impl` when the proprietary module is linked.
- The UI and routes are gated by `features.workflows` from `GET /api/info`.

## Data model

Tables are created in migration `db/sql/migrations/v2.18.15.sql`:

| Table | Role |
| --- | --- |
| `project__workflow_template` | Named workflow definition (`name`, `start_version`, graph) |
| `project__workflow_node` | Node in the graph (`kind`, `template_id`, positions, overrides) |
| `project__workflow_edge` | Directed edge with condition (`on_success`, `on_failure`, `always`) |
| `project__workflow_run` | One execution of a template (`status`, `version`, timestamps) |
| `project__workflow_approval` | Pending/resolved approval for an approval node |
| `task.workflow_run_id` | Links each spawned task back to its run |

Go types: `db/Workflow.go`. Persistence: `pro/db/sql/workflow.go`.

### Node kinds

| `kind` | Behaviour |
| --- | --- |
| `task` | Runs the linked task template (with optional per-node inventory, environment, limit) |
| `approval` | Pauses the run until a user approves or rejects (`approval_message`, optional `approval_timeout`) |
| `note` | Canvas annotation only — not executed, not connected by edges, excluded from validation and the runner |

### Edge conditions

| `condition` | Fires when upstream task finishes with |
| --- | --- |
| `on_success` | Success |
| `on_failure` | Failure |
| `always` | Any terminal status |

`convergence_mode` on a node (`all` or `any`) controls how multiple inbound edges
are satisfied before the node becomes ready.

### Run statuses

| Status | Terminal? | Meaning |
| --- | --- | --- |
| `running` | No | Tasks are executing |
| `approval` | No | Waiting on a human approval node |
| `success` | Yes | All terminal nodes succeeded |
| `failed` | Yes | A failure path completed the run |
| `stopped` | Yes | User stopped the run |

`WorkflowRunStatus.IsFinished()` in `db/Workflow.go` treats `success`, `failed`,
and `stopped` as terminal.

## Execution model

1. **Start** — `POST /api/project/{project_id}/workflows/{workflow_id}/run`
   creates a `WorkflowRun`, computes `version` from `start_version` and prior
   runs (`db.GetNextBuildVersion`), and launches root node(s).
2. **Progression** — When any workflow task reaches a terminal state,
   `TaskRunner.finishRun` calls `WorkflowService.HandleWorkflowTaskCompletion`,
   which evaluates edges, enqueues ready nodes, creates approval records, and
   updates run status. The run view also polls `GetWorkflowRun` every 5s as a
   backstop.
3. **Stop** — `POST …/runs/{run_id}/stop` force-stops all active tasks for the
   run, rejects pending approvals, and marks the run `stopped`. A stopped run is
   never revived by later progression.
4. **Approvals** — `POST …/runs/{run_id}/approvals/{node_id}` with body
   `{ "status": "approved" | "rejected" }` resolves the gate and resumes
   progression.

Interface contracts: `pro_interfaces/workflow_svc.go`,
`pro_interfaces/workflow_ctl.go`.

## Workflow artifacts

Downstream nodes can consume key/value pairs produced by upstream tasks (AWX
`set_stats` parity). See `AGENTS/plans/2_19/workflow-artifacts.md` for recipes.

Summary:

- Tasks receive `SEMAPHORE_ARTIFACTS_FILE` pointing at a JSON file to write.
- Ansible: use `ansible.builtin.set_stats` (callback plugin `semaphore_artifacts`).
- Shell: `echo '{"key":"value"}' > "$SEMAPHORE_ARTIFACTS_FILE"`.
- Downstream tasks get merged artifacts in Ansible extra vars, under
  `semaphore_workflow_artifacts`, and as `SEMAPHORE_WF_<KEY>` env vars.
- `GET …/runs/{run_id}/artifacts` returns the merged map for the run.

**Limitation:** remote runners do not yet stream artifacts back to the server.

## REST API

Swagger definitions: `api-docs.yml` (`workflow` tag). Routes are mounted under
`/api/project/{project_id}/workflows` in `api/router.go`.

| Method | Path | Action |
| --- | --- | --- |
| `GET` | `/workflows` | List workflow templates |
| `POST` | `/workflows` | Create template (nodes + edges in body) |
| `GET/PUT/DELETE` | `/workflows/{workflow_id}` | Read, update, delete template |
| `POST` | `/workflows/{workflow_id}/run` | Start a run |
| `GET` | `/workflows/{workflow_id}/runs` | List runs |
| `GET` | `/workflows/{workflow_id}/runs/{run_id}` | Run details (nodes, approvals, task status) |
| `POST` | `/workflows/{workflow_id}/runs/{run_id}/stop` | Stop an in-progress run |
| `GET` | `/workflows/{workflow_id}/runs/{run_id}/artifacts` | Merged artifact map |
| `GET` | `/workflows/{workflow_id}/runs/{run_id}/approvals` | Pending approvals |
| `POST` | `/workflows/{workflow_id}/runs/{run_id}/approvals/{node_id}` | Approve or reject |

Permissions follow project task permissions (`run_project_tasks` for run/stop,
template management for CRUD).

## UI entry points

| Route | Component |
| --- | --- |
| `/project/{id}/workflows` | `Workflows.vue` — list |
| `/project/{id}/workflows/new`, `…/edit` | `WorkflowEditor.vue` — Drawflow graph editor |
| `/project/{id}/workflows/{workflowId}/runs/{runId}` | `WorkflowRun.vue` — live status graph |

Shared graph component: `web/src/components/WorkflowGraph.vue`. Layout helper:
`web/src/lib/workflowLayout.js`.

## Backup and export

Workflow templates participate in project backup/restore
(`services/project/backup.go`, `cli/cmd/project_export.go`). Nodes and edges
carry `backup` tags on struct fields in `db/Workflow.go`.

## Related code

- DB: `db/Workflow.go`, `db/WorkflowStore_pro.go`, `pro/db/sql/workflow.go`
- API middleware: `api/projects/workflows.go`
- Pro API/service: `pro/api/projects/workflows.go`, `pro/services/server/workflow_svc.go`
- Task integration: `services/tasks/TaskRunner.go`, `services/tasks/TaskPool.go` (`StopTasksByWorkflowRun`)
- Plans: `AGENTS/plans/2_19/graphical-workflow-editor.md`, `AGENTS/plans/2_19/workflow-artifacts.md`
