# Workflows

Workflows chain multiple task templates into a **directed acyclic graph (DAG)** with conditional edges, manual approvals, and shared artifacts between steps. They are a **Pro feature** (`features.workflows`).

## Overview

A workflow consists of:

- **Workflow template** — the graph definition (nodes + edges).
- **Workflow run** — one execution of that graph.
- **Tasks** — each `task` node launches a normal Semaphore task linked to the run via `workflow_run_id` and `workflow_node_id`.

The UI provides a graphical editor (`WorkflowEditor.vue`), a list page (`Workflows.vue`), and a full-screen run view (`WorkflowRun.vue`) with live status on the graph.

## Node kinds

| Kind | Executes | Purpose |
|------|----------|---------|
| `task` | Yes | Runs a project template (playbook, Terraform, etc.) |
| `approval` | Gates | Pauses the run until a user approves or rejects |
| `note` | No | Canvas annotation only; excluded from validation and execution |

Task nodes can override inventory, environment, and Ansible limit per node.

## Edge conditions

Edges connect nodes and control when downstream nodes start:

| Condition | Fires when |
|-----------|------------|
| `on_success` | Source task succeeds (default) |
| `on_failure` | Source task fails |
| `always` | Source task reaches any terminal state |

Approval nodes use `convergence_mode`: `all` (every inbound edge satisfied) or `any` (first satisfied edge).

## Run lifecycle

### Statuses

| Status | Terminal | Meaning |
|--------|----------|---------|
| `running` | No | Tasks executing or waiting to start |
| `approval` | No | Waiting on a manual approval |
| `success` | Yes | All terminal nodes succeeded |
| `failed` | Yes | A task failed with no matching failure edge, or approval rejected |
| `stopped` | Yes | User stopped the run |

### Progression

Run progression is **server-driven**. When any workflow task finishes (success or failure), `TaskRunner.finishRun` calls `HandleWorkflowTaskCompletion`, which:

1. Evaluates which downstream nodes are ready.
2. Launches ready task nodes via the task pool.
3. Creates approval records for approval nodes.
4. Updates the run status.

The run view polls every 5 seconds as a backstop, but progression does not depend on the UI being open.

### Stopping a run

`POST /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id}/stop` (requires `run_project_tasks`):

1. Force-stops all non-finished tasks of the run.
2. Rejects pending approvals.
3. Marks the run `stopped`.

Stopped runs are never revived by later progression.

### Versioning

`start_version` on the template seeds build-style versioning. Each run gets a `version` derived from prior runs (same mechanism as build templates).

## Artifacts

Tasks can publish workflow artifacts (Ansible `set_stats` parity). Downstream nodes in the same run receive merged artifacts as extra variables (`semaphore_workflow_artifacts`). See [`AGENTS/plans/2_19/workflow-artifacts.md`](../AGENTS/plans/2_19/workflow-artifacts.md).

`GET /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id}/artifacts` returns the merged artifact map.

## API endpoints

All routes are under `/api/project/{project_id}/workflows`. Documented in `api-docs.yml` under the `workflow` tag.

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/workflows` | List workflow templates |
| POST | `/workflows` | Create workflow |
| GET | `/workflows/{id}` | Get workflow with nodes and edges |
| PUT | `/workflows/{id}` | Update workflow (returns 204) |
| DELETE | `/workflows/{id}` | Delete workflow |
| POST | `/workflows/{id}/run` | Start a run |
| GET | `/workflows/{id}/runs` | List runs |
| GET | `/workflows/{id}/runs/{run_id}` | Run details (node statuses) |
| POST | `/workflows/{id}/runs/{run_id}/stop` | Stop run |
| GET | `/workflows/{id}/runs/{run_id}/approvals` | Pending approvals |
| POST | `/workflows/{id}/runs/{run_id}/approvals/{node_id}` | Approve or reject |
| GET | `/workflows/{id}/runs/{run_id}/artifacts` | Merged artifacts |

### Validation rules

`db.ValidateWorkflowTemplate` enforces before every write:

- Non-empty name, at least one node.
- Exactly one root (zero in-degree node).
- Graph is a DAG (no cycles).
- Task nodes require `template_id`; approval nodes require positive `approval_timeout` when set.
- No self-edges or dangling edge references.

## Architecture (open vs Pro)

Workflows use the same Pro gating pattern as `terraform_backend` and `project_runners`:

| Layer | Open source | Pro (`pro_impl`) |
|-------|-------------|------------------|
| DB schema + CRUD | `db/sql/workflow.go` | same |
| HTTP handlers | stubs return `[]` or `404` | `pro_impl/api/projects/workflows.go` |
| Orchestration | no-op `WorkflowService` | `pro_impl/services/server/workflow_svc.go` |
| Feature flag | `features.workflows: false` | `true` when `IsPro()` |

The open `TaskPool` delegates `HandleWorkflowTaskCompletion` and `GetWorkflowRunArtifacts` to the injected service (no-ops in open builds).

Wiring in `cli/cmd/root.go`:

```
taskPool := NewTaskPool(...)
workflowService := proServer.NewWorkflowService(store, &taskPool)
taskPool.SetWorkflowService(workflowService)
```

## Persistence notes

- Nodes and edges are stored in child tables; the template row holds metadata only.
- **Writes are delete-and-reinsert**: node database IDs change on every save. The editor re-fetches after save to rebind IDs. Historical runs may reference node IDs that no longer exist in the current template (the run view tolerates missing nodes).
- Node positions (`position_x`, `position_y`) persist per node and survive the reinsert.

## Related code and plans

- `db/Workflow.go` — data model
- `pro_interfaces/workflow_ctl.go`, `workflow_svc.go` — interfaces
- `web/src/components/WorkflowGraph.vue` — shared Drawflow renderer
- [`AGENTS/plans/2_19/graphical-workflow-editor.md`](../AGENTS/plans/2_19/graphical-workflow-editor.md) — implementation details
- [`AGENTS/plans/2_19/workflow-artifacts.md`](../AGENTS/plans/2_19/workflow-artifacts.md) — artifact design
