# Workflow outputs and inputs (ex "artifacts")

Hand-off of structured data between task nodes of one workflow run. A task **produces outputs**
— a JSON object it writes to a file; a downstream node **consumes inputs** — an explicit mapping
of one of its template's survey variables to one output key of one ancestor node. The UI and the
docs say "Outputs" / "Inputs"; the DB column and the API field keep the name `artifacts`.

**Status (2026-10-08): contract agreed, consumer side lives on the edge; the producer side works
for tasks run on the server and on remote runners with the local executor, nothing else does.** A
workflow task gets the file, its outputs are validated and stored on `task.artifacts` — on the
server directly, from a runner through `JobProgress.Outputs` with a second validation on the
server; edges carry no `input_mode` / `input_mappings`, the engine resolves nothing, the UI shows
the raw document. The seeded stand for checking it by hand is described in
`AGENTS/tools/wf-stand/README.md` (`seed-outputs.sh`, `runner.sh`). Work is tracked in
workbench `TASK@1845d9d4d7`; the practices behind the contract are in `RESEARCH@a97b871b1d`.

## Contract

This section is the source of truth for the implementation stages; where a stage body in the
task differs, this wins.

### Producer: the outputs file

- A task that belongs to a workflow run (`WorkflowRunID != nil`) gets the environment variable
  `SEMAPHORE_OUTPUTS_FILE` — the path of an empty file Semaphore creates for that task alone
  (mode 0600, inside the task's temporary directory, removed in `Cleanup`). Tasks outside a
  workflow get neither the variable nor the file. `SEMAPHORE_ARTIFACTS_FILE` never worked for
  anyone, so there is no alias.
- The process writes one UTF-8 JSON object: output name → any JSON value.
  `echo '{"image_tag":"1.4.2"}' > "$SEMAPHORE_OUTPUTS_FILE"`.
- Output names match `^[A-Za-z_][A-Za-z0-9_-]*$`. The hyphen is allowed because Terraform output
  names may contain it and, with explicit mapping, a name is never turned into a variable name.
- Limits: the file ≤ 256 KB, ≤ 100 outputs, one value ≤ 32 KB measured as compact JSON. The
  value limit follows from delivery: a value ends up inside a single process argument
  (`--extra-vars`, `-var k=v`, `k=v`) and Linux refuses an argument over 128 KB.
- There are no reserved names: nothing is injected by output name any more, so an output called
  `task_details` cannot clobber anything.
- Numbers keep their written form (decode with `UseNumber`) — IDs above 2^53 must survive.
- The file is read only when the task **succeeded**. A failed or stopped task has no outputs;
  an `on_failure` branch gets nothing from the node that failed.
- A missing or empty file is "no outputs", not an error. A file that is not a JSON object, has
  a bad name, or breaks a limit **fails the task** with a line in its log naming the reason —
  the mistake shows at the producer, not later at a consumer.
- The file must be a regular file when read (no symlink following, size-capped read): the
  process is less trusted than the server and must not be able to make it read another file.

### Producer adapters by application

| App | How outputs get into the file |
| --- | --- |
| Ansible | `ansible.builtin.set_stats` with `per_host: false` (the default). The bundled `semaphore_outputs` callback writes the `_run` scope as Ansible aggregated it; per-host stats are not outputs. `set_stats` itself rejects names that are not variable names. |
| Terraform, OpenTofu, Terragrunt | After a successful run Semaphore captures `output -json` from the state. A process-written file wins over a captured output of the same name. |
| Bash, Python, PowerShell, Pulumi | The script writes the file itself. |

Captured Terraform outputs never fail the task — the user did not write them. An output marked
`sensitive`, one whose name does not match the pattern, one over the value limit, and anything
that would break the count or file limit (added in name order) is skipped; the task log lists
each skipped name with its reason, never a value.

### Storage and the API shape

`task.artifacts` holds one JSON document per task, written before
`HandleWorkflowTaskCompletion` so the engine on any HA node reads it from the DB:

```json
{
  "values": {"vpc_id": "vpc-0a1b", "subnet_ids": ["subnet-1", "subnet-2"]},
  "skipped": {"db_password": "sensitive", "kubeconfig": "too_large"}
}
```

`skipped` exists so the mapping form can warn when a mapped key is one Semaphore did not
capture; reasons are `sensitive`, `too_large`, `invalid_name`, `limit`. A remote runner sends
the same document in `JobProgress.Outputs`; the server validates it again with the same rules
before storing it — the runner is not trusted.

Outputs are stored and shown in plain text. They are **not for secrets**: no encryption, no
masking, no secret outputs in v1.

### Consumer: input mode and mappings on the edge

Decided 2026-10-07 (supersedes the node-level `input_mappings` of 2026-10-04): the connection
between two nodes says how the source's outputs feed the destination's survey variables.

```json
{
  "source_node_id": 1,
  "destination_node_id": 4,
  "condition": "on_success",
  "input_mode": "explicit",
  "input_mappings": [
    {"var": "vpc_id", "key": "vpc_id"},
    {"var": "subnets", "key": "subnet_ids"}
  ]
}
```

- `input_mode` — `by_name` (default; omitted in JSON) or `explicit`. In the UI it is one checkbox
  on the edge, "Map inputs explicitly".
- **`by_name`**: every output of the source whose name equals the name of a non-`secret` survey
  variable of the destination template feeds that variable. Outputs with no variable of that
  name are ignored, so a hyphenated Terraform output (never a valid variable name) is simply
  not delivered. Nothing to configure, nothing to validate.
- **`explicit`**: only the listed pairs are delivered. `var` — a survey variable of the
  destination template, at most one mapping per `var` on one edge, type `secret` not allowed;
  `key` — one top-level output name (same pattern as the producer). An empty list means the
  edge passes nothing. The source is always the edge's own source node, so a mapping carries no
  node id and needs no `nodeIDMap` remap — edges are reinserted with every revision anyway.
- Only an edge whose destination is a `task` node has an input mode. Everything not fed by an
  edge keeps coming from the node's static `task_params.environment`, exactly as before.
- Stored as two columns on `project__workflow_edge` (`input_mode`, `input_mappings` JSON),
  both with `backup` tags; `ValidateWorkflowTemplate` checks explicit mappings against the
  destination template's survey; `api-docs.yml` documents them on `WorkflowEdge`.

⚠️ Assumptions still to be confirmed by the owner (2026-10-07):

- **Fan-in.** A node with several incoming edges takes contributions from every edge whose
  source task succeeded. If two edges feed the same variable, an `explicit` mapping beats a
  `by_name` one; between two of the same kind the source task that finished last wins. The
  task log names the winner. Conflicting explicit mappings on two edges are not rejected at
  save: with `convergence_mode: any` only one of them may ever fire.
- **Pass-through nodes.** An approval or delay node has no outputs of its own. An edge leaving
  it is transparent: it offers the outputs that reached that node through its own incoming edges,
  so `task → approval → task` still passes data, and the mapping is configured on the last edge
  (the one into the task). The alternative — data stops at an approval — would make the most
  common "deploy after approval" workflow unable to use outputs.

### Resolution when the node starts

The engine walks the node's incoming edges of the run's revision. For each edge it takes, from
the DB, the latest task of the source node in this run (through pass-through nodes, see above).
If that task succeeded, the edge's mode is applied to its `values`: by name, or the explicit
list. A value is used when it is non-null and fits the variable's type. A variable fed by no
edge falls back, in order: the node's static value for that variable → the variable's
`default_value` → nothing. If nothing is left and the
variable is `required`, the node's task is created and immediately failed with a log line naming
the variable, the source node and the key, so the run follows `on_failure`. For a `required`
variable an empty string or empty list counts as no value.

With `convergence_mode: any` an ancestor may be unfinished or never run when the node starts —
that is "no value", resolved by the same fallback.

The resolved values are written into `task.Environment`, so the task shows what it ran with and
the existing delivery (extra vars, `-var`, `k=v`, `target: env`) needs no per-app code. The task
log gets one line per mapped variable saying where its value came from (output of which node and
task / static value / default) — the source, never the value. The total size of mapped values
for one task is capped at 64 KB; over it the task fails with a clear message instead of an
"argument list too long" from the OS.

### Type coercion

The result has the shape the survey form itself produces, so a playbook behaves the same
whether the template is launched by hand or from a workflow.

| Variable type | Accepted output value | Stored in `task.Environment` as |
| --- | --- | --- |
| string (`""`), `text` | string; number or boolean; object or array | the string; its JSON literal (`42`, `true`); compact JSON text |
| `int` | integer number, or a string matching `^-?\d+$` | decimal string (`"42"`) — what the form stores |
| `enum` | scalar whose string form is one of the variable's `values[].value` | that string |
| `select` | array of scalars all in `values[].value`, or one such scalar | array of strings |
| `secret` | — | not mappable |

A value that does not fit is treated as no value (fallback), with the reason in the log line.

### Trust

Output values come from a task process and are data, never code: they must not be evaluated as
Jinja by Ansible (stage 8 of the task decides between `!unsafe` and ansible-core data tagging),
are never printed in full to a log, and are re-validated on the server when they arrive from a
runner. A task created through the API must not be able to set `artifacts`,
`workflow_run_id` or `workflow_node_id`.

## What exists in the code (checked 2026-10-08)

- **Producer, local executor — implemented.** `pro_interfaces.TaskOutputsCollector` /
  `TaskOutputsCapture` (`pro_interfaces/task_outputs.go`) is the open/pro seam: the stub factory
  `pro/services/tasks.NewTaskOutputsCollector` returns nil, the Pro one returns
  `artifacts.Collector`; `cli/cmd/root.go` injects it into `TaskPool`
  (`SetTaskOutputsCollector`), which hands it to each `LocalExecutor`.
  `LocalExecutor.beginOutputs` (in `Prepare`, only when `WorkflowRunID != nil`) creates
  `<project tmp>/task_<id>_outputs_*/outputs.json` (dir 0700, file 0600, chowned to the process
  user) and adds the env; `collectOutputs` (in `Run`, after the app returned without error, the
  task is not killed and its status is still `running`) reads it; `Cleanup` removes it.
- **Stored before the success status.** `TaskRunner.saveOutputs` calls `UpdateTaskArtifacts`
  before `SetStatus(success)`: in HA any node may progress the run once it sees the terminal
  status in the DB, so the outputs must already be there. For the same reason
  `TerraformApp.Run` no longer sets `success` itself on "no changes" / plan-only — the caller
  does it after `Run` returns (both `TaskRunner.run` and the runner's `job_pool` already did).
- **Package `pro_impl/services/tasks/artifacts`** now implements the contract: `Parse` /
  `ReadFile` (strict), `Build` (lenient Terraform capture), `Document` (`Encode`, `Notes`,
  `Decode` — the validation a runner-sent document will go through), `Collector`. `Merge`,
  `ToShellEnv`, the reserved names and `SEMAPHORE_ARTIFACTS_FILE` are gone.
  `CollectFromTasks` stays only to feed `GET …/runs/{r}/artifacts` with merged `values`.
- **Ansible wiring.** The callback is `semaphore_outputs.py`; it has no
  `CALLBACK_NEEDS_ENABLED`, so Ansible loads it without `ANSIBLE_CALLBACKS_ENABLED` and the
  user's `callbacks_enabled` is untouched. `ANSIBLE_CALLBACK_PLUGINS` is set to the value the
  task already had (task env, `Config.EnvVars`, forwarded server env) or else Ansible's default
  path, plus the plugin dir (`LocalExecutor.ansibleCallbackPlugins`).
  ⚠️ 2026-10-04: the variable overrides `callback_plugins` of an `ansible.cfg` — a path set
  only there is lost for workflow tasks (AWX and ansible-runner behave the same). A
  `callback_plugins/` directory next to the playbook is unaffected.
- **Terraform capture.** `TerraformApp.Outputs` runs `<binary> output -json` after a successful
  run and keeps the result out of the task log (it carries sensitive values in plain text).
  Checked by hand against Ansible core 2.21, Terraform and OpenTofu.
- **Producer, remote runner — implemented (2026-10-08).** `JobProgress.Outputs *string`
  (`services/runners/types.go`, `json:"outputs,omitempty"`, ignored by an old server, never sent
  by an old runner) carries the same document as `task.artifacts`. On the runner
  `runningJob.finish` sets the success status and the outputs under one lock, so a progress
  snapshot never reports success without them, and `JobPool.taskOutputs` reads them from any
  executor implementing `tasks.OutputsProvider` (`LocalExecutor.Outputs`). The collector reaches
  the runner's `LocalExecutorProvider` through `JobPool.SetTaskOutputsCollector`, wired in
  `cli/cmd/runner.go` from the same open/pro factory as the server. On the server
  `UpdateRunner` (`api/runners/runners.go`) calls `TaskPool.StoreRemoteTaskOutputs` **before**
  `SetStatus(success)`: `TaskOutputsCollector.Validate` (pro: `artifacts.Decode` + `Encode`)
  re-checks the document, `UpdateTaskArtifacts` persists it; an invalid document fails the task
  with a line in its log. Outputs are accepted only with the success status and only for a task
  with `WorkflowRunID`; `applyDBPersistedTaskSnapshot` carries `Artifacts`, so an HA node that
  hydrates the task sees them. Checked end to end on the stand with `runner.sh` (bash,
  Terraform and `set_stats` producers stored, the six invalid producers failed on the runner).
- A task of a workflow run on a Docker/K8s executor gets a task-log line saying this executor
  does not capture outputs (`JobPool.taskOutputs`); the blanket "remote runner" alert in the run
  view and its `workflowArtifactsRemoteRunnerWarning` string are gone — the server cannot tell a
  runner's executor type, and the local executor on a runner now works.
- Still as before: `GET /project/{p}/workflows/{w}/runs/{r}/artifacts` (`api/router.go`,
  `api-docs.yml`) → `WorkflowService.GetWorkflowRunArtifacts`; the task dialog pretty-prints
  `task.artifacts` (`web/src/components/TaskDetails.vue`) — now the `values`/`skipped`
  document.
- ⚠️ 2026-10-04: the server never applies survey defaults or validates survey types — only the
  Vue forms do (`TaskParamsForm.vue`; see `../secrets-and-task-vars/schedule-survey-defaults.md`).
  The resolver therefore applies `default_value` itself for mapped variables; `int` values are
  strings in `task.Environment` because the form stores them so.
- ⚠️ 2026-09-27: `docs/docs/user-guide/workflows.md` § "Workflow artifacts (set_stats)" tells
  users the feature works for local tasks and describes the superseded implicit merge. Until
  the last stage rewrites it the doc overclaims.

## Decisions and why

- **Mapping lives on the edge; by name by default, explicit per edge** (owner, 2026-10-07).
  Supersedes the node-level `input_mappings` with `source_node_id` (2026-10-04). The common
  case — the producer names its outputs like the consumer's survey variables — needs zero
  configuration; the checkbox on the connection is where a user naturally looks for "what
  flows along this arrow". The AWX problems that killed the first implicit design stay
  bounded: outputs only ever land on declared survey variables (never undeclared `-var`s or
  stray extra vars), the scope of a by-name match is one edge, every delivered value is logged
  with its source, and the fan-in tie-break is fixed. Accepted cost: in `by_name` mode an output
  overrides the node's static value of the same name — that is the point of the feature, and
  the task log shows it.
- **Explicit mapping, not implicit merge** (2026-09-28, partly superseded above). The first
  design copied AWX: outputs of all finished upstream tasks merged (later task wins) and
  injected as flat extra vars, a `semaphore_workflow_artifacts` namespace and `SEMAPHORE_WF_*`
  env. Dropped: the merge order at a fan-in is arbitrary, an output can silently override a
  survey value, and Terraform rejects an undeclared `-var`. Mapping onto survey variables
  reuses the delivery every app type already has.
- **File named by an env var, not stdout markers** — app-agnostic, and a log stream is the wrong
  channel for data (GitHub retired `::set-output`; Kestra leaked encrypted outputs through its
  stdout marker).
- **Outputs only from a successful task** (2026-10-04) — one rule, no half-written data from a
  crashed process.
- **A bad file fails the producer** (2026-10-04) — as Tekton and Argo do; a warning would move
  the error to a consumer far from its cause.
- **Count and value limits on top of 256 KB, no per-run total** (2026-10-04) — the numbers (100
  outputs, 32 KB per value) are sized by the 128 KB single-argument limit of Linux, not copied
  from a product.
- **Sensitive Terraform outputs are skipped but named** (2026-10-04) — in the task log and, via
  `skipped`, as a warning in the mapping form; opt-in encrypted transfer is a later feature.
- **Hyphen allowed in names, reserved names dropped** (2026-10-04) — both restrictions served
  the implicit projection into variable names; under explicit mapping they would only make a
  valid Terraform output unusable or fail a task for a harmless name.
- **Fallback before failure** — output → static node value → default → fail if required. Softer
  than Spacelift or Tekton (missing input is an error there) but every substitution is logged.
- **Per-task column, resolved on read** — no run-level blob to keep consistent; HA nodes need no
  coordination.
- **Pro-gated with the engine**: the package lives in `pro_impl`; the open module calls it
  through the open/pro interface, without globals.

## Open, broken, deferred

- Still to be built: the edge columns and validation, the resolver in `startWorkflowNode`, the
  UI on the edge, the security hardening, docs and Dredd (stages 5–9 of the task). The two ⚠️
  assumptions in § Consumer (fan-in tie-break, pass-through nodes) are open until the owner
  confirms them; the research direction `AREA@7ec03538fb` (2026-10-07) recommends a
  deterministic tie-break by edge id, rejecting conflicting explicit mappings at save under
  `convergence_mode: all`, and stating that pass-through carries raw outputs.
- Docker/K8s executors are out of v1. The Kubernetes termination message (4 KB per pod) cannot
  carry 256 KB; the path is a file on a shared volume read by the runner side. Until then a
  workflow task on those executors only gets the task-log line from `JobPool.taskOutputs`.
- Key hints in the mapping form come from a previous run, but node IDs change on every revision
  save — the hint source needs an identity that survives a save (decide in the UI stage). With
  mappings on the edge the form knows both ends: destination survey variables from the
  template, candidate keys from the last run of the source node.
- `GET …/runs/{r}/artifacts` returns a merged map, meaningless under explicit mapping; the run
  view needs outputs per node.
- Related gaps found on the way, not part of this contract: secret survey values set on a node
  never reach the task; Docker/K8s executors ignore survey secrets and `target: env`; no unique
  index on `task(workflow_run_id, workflow_node_id)` — all recorded as findings on the task.
- Dredd skips the artifacts endpoint (`.dredd/hooks/main.go`).

## Key files and references

- `db/Task.go`, `db/Store.go`, `db/sql/task.go`, `db/sql/migrations/v2.18.15.sql`
- `db/Workflow.go` (`WorkflowNode`), `db/Template.go` (`SurveyVar`, `SurveyVarDefaultValue`),
  `db/TaskParams.go`
- `services/tasks/local_executor.go` (outputs: `beginOutputs`, `collectOutputs`,
  `ansibleCallbackPlugins`; survey delivery: `getEnvironmentExtraVars`, `getSurveyEnvVars`,
  `formatVarValue`), `services/tasks/TaskRunner.go` (`saveOutputs`),
  `services/tasks/TaskPool.go` (`StoreRemoteTaskOutputs`, `applyDBPersistedTaskSnapshot`),
  `services/tasks/local_executor_provider.go`, `services/tasks/executor.go`
  (`OutputsProvider`), `db_lib/TerraformApp.go` (`Outputs`), `pro_interfaces/task_outputs.go`
- Runner protocol: `services/runners/types.go` (`JobProgress.Outputs`),
  `services/runners/running_job.go` (`finish`), `services/runners/job_pool.go`
  (`taskOutputs`, `SetTaskOutputsCollector`), `api/runners/runners.go` (`UpdateRunner`),
  `cli/cmd/runner.go`; stand check — `AGENTS/tools/wf-stand/runner.sh`
- `db/Workflow.go` (`WorkflowEdge`), `pro_impl/db/sql/workflow.go` (edge insert / select per
  revision), `services/project/{types,backup,restore}.go`
- `pro_interfaces/workflow_svc.go`, `pro/services/server/workflow_svc.go` (stub),
  `pro_impl/services/server/workflow_svc.go` (`startWorkflowNode`, `mapLatestNodeTask`),
  `pro_impl/db/Workflow.go` (`ValidateWorkflowTemplate`), `pro_impl/services/tasks/artifacts/`
  (incl. `ansible/callback_plugins/semaphore_outputs.py`)
- `web/src/components/workflow/WorkflowEdgeProperties.vue` (the edge panel: condition today,
  the input-mode checkbox and mapping table next), `web/src/views/project/WorkflowEditor.vue`
  (`onConnectionSelected`), `web/src/components/TaskParamsForm.vue`,
  `web/src/components/TaskDetails.vue`, `web/src/views/project/WorkflowRun.vue`
- `api-docs.yml` (artifacts path), `docs/docs/user-guide/workflows.md`
- https://github.com/semaphoreui/semaphore/pull/3488 (origin of the feature)
