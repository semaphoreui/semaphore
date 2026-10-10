# Workflow outputs and inputs (ex "artifacts")

Hand-off of structured data between task nodes of one workflow run. A task **produces outputs**
— a JSON object it writes to a file; a downstream node **consumes inputs** — an explicit mapping
of one of its template's survey variables to one output key of one ancestor node. The UI and the
docs say "Outputs" / "Inputs"; the DB column and the API field keep the name `artifacts`.

**Status (2026-10-09): the hand-off works end to end on the server and on remote runners with
the local executor: outputs are captured and stored, edges store and validate `input_mode` /
`input_mappings`, the engine fills the survey variables of the next task from them, and the
editor configures it on the connection; survey values and inputs reach Ansible as literals
(no Jinja evaluation) and the task API cannot forge workflow fields. Left: docs and Dredd.** A workflow task gets the file, its
outputs are validated and stored on `task.artifacts` — on the server directly, from a runner
through `JobProgress.Outputs` with a second validation on the server; `startWorkflowNode`
resolves the inputs of the next task before it is enqueued. The seeded stand for checking it by hand is described in
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
  Implemented 2026-10-08 (migration `v2.20.10`): `input_mode` is `not null default 'by_name'`,
  so a stored edge reads back and is returned by the API with `"input_mode":"by_name"` (the
  contract's "omitted in JSON" holds only for what a client sends — same as `kind` and
  `convergence_mode` on nodes); `input_mappings` is NULL for by_name and a JSON list (possibly
  `[]`) for explicit. Validation: `by_name` passes on any edge; `explicit` or a non-empty list
  is rejected on an edge into an approval/delay node and on a by_name edge; in explicit, `var`
  must be a non-secret survey variable of the destination template, `key` must match the
  output-name pattern (`db.IsValidWorkflowOutputName`), one mapping per `var`.

Fan-in and pass-through (implemented 2026-10-08 as the research `AREA@7ec03538fb` recommended;
supersedes the 2026-10-07 assumptions):

- **Fan-in.** A node with several incoming edges takes contributions from every edge whose
  source task succeeded. If two edges feed the same variable, an `explicit` mapping beats a
  `by_name` one; between two of the same kind **the edge with the lower id wins** — never the
  task that finished last, so a run gives the same result every time. The task log names the
  winner and the number of edges that lost. Two explicit edges mapping the same variable of one
  node are rejected at save when the node's `convergence_mode` is `all` (both would fire); with
  `any` they are accepted (one branch runs).
- **Pass-through nodes.** An approval or delay node has no outputs of its own. An edge leaving
  it is transparent: it offers the **raw** outputs that reached that node through its own
  incoming edges (merged by the same edge-id order, any depth), so `task → approval → task`
  still passes data, and the mode and mappings are applied once, on the last edge (the one into
  the task). The alternative — data stops at an approval — would make the most common "deploy
  after approval" workflow unable to use outputs.

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

Implemented 2026-10-08 in `pro_impl/services/server/workflow_inputs.go`
(`resolveWorkflowNodeInputs`, called by `startWorkflowNode` with the pass's task snapshot):
the fallback chain and the required check apply to **every** non-secret survey variable of the
destination template, not only to mapped ones — the server never applied survey defaults
before, now a workflow task gets them. A failed resolution does not enqueue anything:
`WorkflowTaskEnqueuer.AddFailedTask` stores the task already finished with status `error`
(start = end = now, no alerts, creation and completion audit events) and the lines in its log;
the progression pass that created it re-reads the run and follows `on_failure`. For an enqueued
task the lines go through `WorkflowTaskEnqueuer.LogTask`, which writes through the pool's log
channel even though the task is not registered in the pool yet (registration goes through the
pool loop). A `target: env` variable reaches the process under its own name (`env_name`, not
`ENV_NAME`).

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

Output values come from a task process and are data, never code: they are never evaluated as
Jinja by Ansible, are never printed in full to a log, and are re-validated on the server when
they arrive from a runner. A task created through the API cannot set `artifacts`,
`workflow_run_id`, `workflow_node_id` or `runner_id` — the engine and the outputs store are the
only writers.

Ansible never evaluates a Jinja2 expression inside a value Semaphore passes, and there is no
switch to turn evaluation back on (owner's decision 2026-10-09, NOTE@ca2d7bb0cc and its follow-up;
AWX's `ALLOW_JINJA_IN_EXTRA_VARS=never`). The rule applies to **every** Ansible task, not only
to workflow tasks, because `ansible-playbook --extra-vars '<json>'` evaluates
`{{ lookup('pipe', …) }}` inside any value lazily (ansible-core 2.21.2 probe, 2026-10-09). All
extra vars go through temporary YAML files in which every string scalar carries the `!unsafe`
tag, recursively through lists and objects (`services/tasks/extra_vars_file.go`); numbers and
booleans need no tag. Since 2026-10-10 (TASK@CEA87DDD08) there are two files per task, both in
the project's tmp directory, 0600, chowned to the task user and deleted in `Cleanup`:

- `task_<id>_extra_vars_*.yml` — variable groups, survey answers, workflow inputs,
  `semaphore_vars`; in clear.
- `task_<id>_secret_vars_*.yml` — the survey secrets and the variable-group secrets of type
  `var`, encrypted in the Ansible Vault 1.1 format (`pkg/ansible_vault`, stdlib only) with a
  one-off password. The password and the vault id are random GUIDs that live in the executor's
  memory only: ansible-playbook gets `--vault-id <guid>@prompt` and the password is typed on its
  prompt through the same pty mechanism that answers the template's own vault prompts
  (`db_lib/AnsiblePlaybook.go`). The file comes last on the command line, so a secret still
  overrides a group variable of the same name. No secret is on disk in clear while the task
  runs, and none is in `ps`. A task without secrets writes the open file only.

The vault header carries no id, so Ansible tries every `--vault-id` it was given on it: a
template's own password vault and the secrets vault coexist (stand task 171, 2026-10-10). Files
of the repository that Ansible loads itself (`vars_files`, `group_vars`) are the playbook
author's code and are still templated — the `!unsafe` rule is for data Semaphore passes.
`SEMAPHORE_DEBUG_FILTER=extra_vars` logs the split, the files and their removal by name
(`AGENTS/work/logging/debug-filter.md`). Side effects to document: Jinja typed into a survey
answer or a variable group is no longer evaluated, and secrets no longer appear in the
`ansible-playbook` command line or on disk in clear.

## What exists in the code (checked 2026-10-09)

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
- **UI (2026-10-09).** The edge panel (`WorkflowEdgeProperties.vue`) shows an "Inputs" section
  only when the destination is a task node with a template: the checkbox "Map inputs explicitly"
  switches `input_mode`; unchecked, it lists the destination's non-secret survey variables as
  chips (a check on those the source produced in the latest finished run of this revision);
  checked, one row per variable with a combobox for the output key (suggestions from that run,
  pattern validation, a hint when the key was `skipped`). Key hints come from the latest run
  whose `revision_id` equals the loaded template's (`WorkflowEditor.loadOutputHints`) — after a
  save the node ids change and there are no hints until the next run; an approval/delay source
  has no hints. The canvas keeps `input_mode`/`input_mappings` per edge next to the condition
  (`WorkflowGraph.inputs`, `setEdge` applies condition and inputs as one undo step) and marks an
  explicit edge with a swap icon on its pill and the `WorkflowGraph__conn--explicit` class. The
  task dialog's Details tab shows an "Outputs" table (`values`) and "Not captured" chips
  (`skipped` with the reason); the run view card says "N outputs" for a producer. Screenshots
  and the Playwright checks: `mocks/README.md` (`shoot-inputs.cjs`, `interact-inputs.cjs`).
- Still as before: `GET /project/{p}/workflows/{w}/runs/{r}/artifacts` (`api/router.go`,
  `api-docs.yml`) → `WorkflowService.GetWorkflowRunArtifacts`, used by nothing in the UI.
- ⚠️ 2026-10-04: the server never applies survey defaults or validates survey types — only the
  Vue forms do (`TaskParamsForm.vue`; see `../secrets-and-task-vars/schedule-survey-defaults.md`).
  The resolver therefore applies `default_value` itself for mapped variables; `int` values are
  strings in `task.Environment` because the form stores them so.
- ⚠️ 2026-09-27: `docs/docs/user-guide/workflows.md` § "Workflow artifacts (set_stats)" tells
  users the feature works for local tasks and describes the superseded implicit merge. Until
  the last stage rewrites it the doc overclaims.

- **Security (2026-10-09).** `LocalExecutor.getPlaybookArgs` always writes the extra-vars file
  (`writeExtraVarsFile`; `extraVarsYAML` builds a `yaml.v3` node tree, `markUnsafe` tags every
  string); the inline JSON builder and the separate `--extra-vars name=secret` arguments are
  gone. `api/projects/tasks.go` `sanitizeClientTask` zeroes `WorkflowRunID`, `WorkflowNodeID`,
  `Artifacts`, `RunnerID` on `POST …/tasks`. Stand evidence (workflows 13 "Inputs: jinja
  injection" and 14 "Inputs: jinja vs vars group", server and runner): `INPUT image_tag={{
  lookup('pipe', 'echo INJECTED-$(id -un)') }}` and `INPUT group_expr={{ 2*21 }}` printed
  literally. The same build with the earlier inline form printed `INPUT image_tag=INJECTED-fiftin`
  and `region=/Users/fiftin`, which is the evaluation the file prevents. No `task_*_extra_vars_*`
  files remain after the runs.

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

- Still to be built: docs and Dredd (stage 9 of the task), including the behaviour change that
  Jinja in survey answers and variable groups is no longer evaluated. The two ⚠️
  assumptions in § Consumer (fan-in tie-break, pass-through nodes) are open until the owner
  confirms them; the research direction `AREA@7ec03538fb` (2026-10-07) recommends a
  deterministic tie-break by edge id, rejecting conflicting explicit mappings at save under
  `convergence_mode: all`, and stating that pass-through carries raw outputs.
- Docker/K8s executors are out of v1. The Kubernetes termination message (4 KB per pod) cannot
  carry 256 KB; the path is a file on a shared volume read by the runner side. Until then a
  workflow task on those executors only gets the task-log line from `JobPool.taskOutputs`.
- Key hints in the mapping form come from the latest finished run of the loaded revision only:
  node ids change on every save, so right after a save there are no hints until the workflow
  runs again. An identity that survives a save (e.g. a stable node key) would lift this.
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
  `formatVarValue`; Jinja rule: `getPlaybookArgs`), `services/tasks/extra_vars_file.go`
  (`extraVarsYAML`, `markUnsafe`, `writeExtraVarsFiles`, `takeSecretExtraVars`),
  `pkg/ansible_vault/vault.go` (`Encrypt`, `Decrypt`), `api/projects/tasks.go`
  (`sanitizeClientTask`); stand checks —
  `AGENTS/tools/wf-stand/seed-outputs.sh` workflows 13–14, `services/tasks/TaskRunner.go` (`saveOutputs`),
  `services/tasks/TaskPool.go` (`StoreRemoteTaskOutputs`, `applyDBPersistedTaskSnapshot`),
  `services/tasks/local_executor_provider.go`, `services/tasks/executor.go`
  (`OutputsProvider`), `db_lib/TerraformApp.go` (`Outputs`), `pro_interfaces/task_outputs.go`
- Runner protocol: `services/runners/types.go` (`JobProgress.Outputs`),
  `services/runners/running_job.go` (`finish`), `services/runners/job_pool.go`
  (`taskOutputs`, `SetTaskOutputsCollector`), `api/runners/runners.go` (`UpdateRunner`),
  `cli/cmd/runner.go`; stand check — `AGENTS/tools/wf-stand/runner.sh`
- `db/Workflow.go` (`WorkflowEdge`, `WorkflowEdgeInputMode`, `WorkflowInputMapping`,
  `IsValidWorkflowOutputName`), `db/sql/migrations/v2.20.10.sql`, `pro_impl/db/sql/workflow.go`
  (edge insert / select per revision, `fillWorkflowEdge`, `workflowEdgeInputMappingsJSON`),
  `pro_impl/db/Workflow.go` (`validateWorkflowEdgeInputs`),
  `services/project/{types,backup,restore}.go`
- Resolver: `pro_impl/services/server/workflow_inputs.go` (`resolveWorkflowNodeInputs`,
  `outputsReaching`, `coerceWorkflowInput`), `pro_impl/services/server/workflow_svc.go`
  (`startWorkflowNode`), `pro_interfaces/workflow_svc.go` (`WorkflowTaskEnqueuer.AddFailedTask`,
  `LogTask`), `services/tasks/TaskPool.go` (their implementation); stand checks —
  `AGENTS/tools/wf-stand/seed-outputs.sh` workflows 9–12, `inputs.sh`
- `pro_interfaces/workflow_svc.go`, `pro/services/server/workflow_svc.go` (stub),
  `pro_impl/services/server/workflow_svc.go` (`startWorkflowNode`, `mapLatestNodeTask`),
  `pro_impl/db/Workflow.go` (`ValidateWorkflowTemplate`), `pro_impl/services/tasks/artifacts/`
  (incl. `ansible/callback_plugins/semaphore_outputs.py`)
- `web/src/components/workflow/WorkflowEdgeProperties.vue` (the edge panel: condition, input
  mode checkbox, mapping table, key hints), `web/src/components/WorkflowGraph.vue` (`inputs`,
  `setEdge`, explicit class and label flag), `web/src/components/workflow/WorkflowEdgeLabel.vue`,
  `web/src/views/project/WorkflowEditor.vue` (`loadOutputHints`, `applyEdgeEdit`),
  `web/src/components/TaskDetails.vue` (Outputs panel), `web/src/views/project/WorkflowRun.vue`
  + `WorkflowNodeCard.vue` (output count), `web/src/lang/en.js` (`workflowEdgeInputs*`,
  `workflowOutputs*`)
- `api-docs.yml` (artifacts path), `docs/docs/user-guide/workflows.md`
- https://github.com/semaphoreui/semaphore/pull/3488 (origin of the feature)
