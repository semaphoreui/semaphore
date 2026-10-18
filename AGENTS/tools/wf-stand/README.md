# wf-stand — run and verify workflows on the local stand

Drives the agent's throwaway Semaphore stand (`http://localhost:3100`, admin/admin123, project 1,
Pro build from HEAD) to execute a workflow end to end and follow every node to a terminal status.
Exists because a workflow run is asynchronous: tasks, approvals and delays resolve over time and
only the run-details endpoint shows all of them together.

Call forms (from the repository root):

```
AGENTS/tools/wf-stand/stand.sh {start|stop|restart|build|status|log [n]}
AGENTS/tools/wf-stand/seed.sh
AGENTS/tools/wf-stand/seed-outputs.sh          # PROJECT_NAME="…" to seed another project
AGENTS/tools/wf-stand/run.sh <workflow_id> [--approve|--reject] [--stop-after SEC] [--expect success|failed|stopped] [--timeout SEC] [--run ID]
AGENTS/tools/wf-stand/runner.sh {start|stop|status|tag|untag}   # PID=<project> TAG=<tag>; remote runner for the stand
AGENTS/tools/wf-stand/inputs.sh <workflow_id> [run_id]           # outputs stored per producer + "INPUT name=value" lines per consumer
NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/wf-stand/shoot-run.cjs <workflow_id> [run_id] [out.png] [--dark]
AGENTS/tools/wf-stand/seed-docs.sh                                # project "Demo Stand" + workflow "Build and deploy" for the docs screenshots
AGENTS/tools/wf-stand/seed-survey.sh                              # workflow "Survey: workflow variables" (survey + node bindings) in the seed-outputs project
NODE_PATH=… node AGENTS/tools/wf-stand/shoot-editor.cjs <workflow_id> [node_title] [out.png] [--dark]   # editor with a node selected: survey panel, canvas, bindings
NODE_PATH=… node AGENTS/tools/wf-stand/shoot-run-dialog.cjs <workflow_id> [out_dir] [--dark]       # starts a run through the survey dialog; shoots the dialog and the run view's Variables strip
PID=<project> WF=<workflow> NODE_PATH=… node AGENTS/tools/wf-stand/shoot-docs.cjs   # the four screenshots of workflows.md § Outputs and inputs
```

Examples:

```
$ AGENTS/tools/wf-stand/stand.sh build && AGENTS/tools/wf-stand/stand.sh start
stand: up at http://localhost:3100 (pid 40366)
$ AGENTS/tools/wf-stand/seed.sh
workflows happy=5 branches=6                      # ids also in /tmp/semaphore-stand/wf-exec-ids.txt
$ AGENTS/tools/wf-stand/run.sh 5 --approve --expect success
run 4 started (workflow 5) — UI: http://localhost:3100/project/1/workflows/5/runs/4
17:30:14 approval version=0.1.0 n18/task=task#6:success n19/approval=approval:pending n20/delay=- n21/task=-
17:30:14 node 19 -> approved
17:30:24 success version=0.1.0 n18/task=task#6:success n19/approval=approval:approved n20/delay=delay:success n21/task=task#7:success
final: success (run 4)
$ AGENTS/tools/wf-stand/run.sh 6 --stop-after 8 --expect stopped
final: stopped (run 6)
```

Seeded workflows: "Exec happy" (task → approval → delay 5 s → task, ends `success` with `--approve`)
and "Exec branches" (ok → fail; on_failure → recover, on_success → never started; always → 30 s
slow task; ends `failed`). Templates are `bash` scripts from the local repo
`/tmp/semaphore-stand/wf-repo` (`ok.sh`, `fail.sh`, `slow.sh`), so no network or Ansible is needed.

`seed-outputs.sh` seeds the project "Artifacts QA" (project 1 since the 2026-10-07 rebuild) for
testing workflow outputs by hand: 16 templates (bash producers for every rule of the contract —
valid, empty, bad JSON, bad name, 101 outputs, 40 KB value, outputs-then-exit-1, symlink; bash and
Ansible consumers with one survey var of every type; Terraform / Tofu with a `sensitive` and a
hyphenated output; Ansible `set_stats`) and four workflows: 1 "producers -> consumers" (expect
`success`, see `task.artifacts` of tasks of nodes 2–4), 2 "invalid producers" (expect `failed`,
each bad producer `error` with the reason in its log, "empty" `success` without artifacts),
3 "failed producer" (expect `failed`, no artifacts stored, `on_failure` consumer ran), 4 "tofu ->
consume" (expect `success`). Scripts: `/tmp/semaphore-stand/wf-repo/produce*.sh`, `consume.sh`,
`tf/main.tf`, `playbooks/*.yml`. Ids land in `/tmp/semaphore-stand/wf-outputs-ids.txt`.

Since 2026-10-08 the same seed adds the mapping checks: templates "bash: produce renamed outputs" and
"ansible: set_stats renamed outputs" (output names `tag`, `count`, `flag`, `subnets`, `cfg`, `envname` that
match no survey variable, so only an explicit edge mapping can deliver them) and workflows 9 "Inputs: explicit
mapping" (renamed producers → bash and ansible consumers over `explicit` edges, `image_tag←tag`,
`replicas←count`, …), 10 "Inputs: by_name through approval" (produce → approval → consumers, run with
`--approve`), 11 "Inputs: fan-in by_name vs explicit" (produce by name and renamed explicitly into one consumer
with convergence all: expect `image_tag` from the explicit edge, `replicas` by name), 12 "Inputs: explicit empty
list" (the edge passes nothing). Consumers print one `INPUT name=value` line per survey variable (bash: parsed
`k=v` args plus `env:env_name` — a `target: env` variable keeps its own name, it is not upper-cased; ansible: a debug list), so `inputs.sh <wf>` shows producer outputs next to what
the consumer received. Expected since the resolver landed (2026-10-08): workflow 1 and 10 — the consumer's
`INPUT` lines repeat the producer's `image_tag`, `replicas`, `enabled`, `subnet_ids`, `config` (by name;
`env-name` with a hyphen is not delivered); workflow 9 — the renamed values (`image_tag=renamed-…`,
`replicas=7`, `env:env_name=prod`); workflow 11 — `image_tag=renamed-…` from the explicit edge and `replicas=3`
by name; workflow 12 — only `image_tag=static-from-node`, `replicas=1` (the default) and `env:env_name=static-env`;
workflow 3 — the on_failure consumer is `error` before start with "Required input(s) without a value: image_tag"
in its log. The task log of every consumer starts with the engine's `Input "<var>" <- output "<key>" of node N
(task #M)` lines, so a value's source is always one line above the value.

Since 2026-10-09 the seed also covers the Jinja rule (Ansible never evaluates an expression inside an
extra var — every value reaches it through a file, the ones with a Jinja2 delimiter tagged `!unsafe`): template "bash: produce jinja outputs"
(`produce_jinja.sh` writes `image_tag`, `config` and `replicas` holding `{{ lookup('pipe', …) }}`,
`{{ lookup('env', 'HOME') }}`, `{{ 6*7 }}`), the variable group "vars group with jinja" (`group_expr: "{{ 2*21 }}"`),
template "ansible: consume + vars group" and workflows 13 "Inputs: jinja injection" (→ ansible and bash
consumers) and 14 "Inputs: jinja vs vars group". Expected on the server and on the runner: `INPUT image_tag={{
lookup('pipe', 'echo INJECTED-$(id -un)') }}`, `config` with the expressions verbatim, `replicas=1` (the
default — `{{ 6*7 }}` is not an int) and `INPUT group_expr={{ 2*21 }}` — the variable group is data too. A
consumer log containing `INJECTED-` or `group_expr=42` means Ansible evaluated a value: a regression.
`inputs.sh` right after `run.sh` on the runner can miss the last log lines; re-run it with the run id.
Trap: the template API ignores `environment_id`, pass `environment_ids`.

Since 2026-10-10 the stand also checks the secret extra vars (TASK@CEA87DDD08): the survey secret
`token` and the variable-group secret `db_password` (type `var`, added to "vars group with jinja" by hand
through `PUT /project/1/environment/1` with `secrets:[{type:"var",name,secret,operation:"create"}]`) reach
Ansible through a second, vault-encrypted `--extra-vars @task_<id>_secret_vars_*.yml` with a one-off
`--vault-id <guid>@prompt`. `playbooks/consume.yml` prints `INPUT token=…` and `INPUT db_password=…`;
template 37 "ansible: consume + own vault" (access key "tpl vault", type login_password, vault `default`)
runs `playbooks/consume_vault.yml`, which loads `playbooks/vaulted.yml` (encrypted by `ansible-vault` with
that key's password) through `vars_files` and prints `INPUT vaulted_var=…`. Expected on the server and on
the runner: `INPUT token=s3cr3t {{ 1+1 }}`, `INPUT db_password=p@ss {{ 7*6 }}` literally; `ps -axo args=`
shows `--vault-id=<guid>@prompt` and two `--extra-vars @`, never a secret or the password; `grep -rl
s3cr3t /tmp/semaphore-stand/tmp` (or `runner-tmp`) finds nothing while the task runs; no `task_*_vars_*`
file is left after. With `SEMAPHORE_LOG_LEVEL=DEBUG SEMAPHORE_DEBUG_FILTER=extra_vars` exported before
`stand.sh start` / `runner.sh start`, the server and runner logs show seven `context=extra_vars` lines per
task (split, two files written, vault added, two files removed, password forgotten). Checked 2026-10-10:
server task 162, runner task 164, own vault task 171, regression runs 39 (wf 13) and 40 (wf 14).
Traps: a vault-encrypted file must not sit in `playbooks/group_vars/` — Ansible loads it for every playbook
of the directory and a template without that vault fails with "Decryption failed"; `vars_files` of the
repository are still templated (`{{ 3*3 }}` prints `9`) — only values Semaphore passes are `!unsafe`;
`runner.sh start | tail` hangs because the runner inherits the pipe — redirect to a file instead; a PUT of a
template without the `vaults` field deletes its vaults, which is why `runner.sh tag/untag` re-reads each
template by id before the PUT.

Docs screenshots (2026-10-11): `seed-docs.sh` seeds a separate project "Demo Stand" (the name the published
editor screenshot already shows) with "Provision network" (bash, outputs `vpc_id`, `subnet_ids`), "Build image"
(bash, outputs `image_tag`, `replicas` — the values quoted in the user guide), "Deploy to cluster" (ansible
`playbooks/deploy.yml`, survey `image_tag`, `replicas`, `subnets`) and the workflow "Build and deploy":
Provision → Build → Deploy by name plus Provision → Deploy explicit (`subnets ← subnet_ids`). Run it once
(`PID=<p> run.sh <wf> --expect success`), then `shoot-docs.cjs` writes `workflow-inputs-explicit.png`,
`workflow-inputs-by-name.png`, `workflow-run-outputs.png` and `task-outputs.png` (dialog clipped) to
`/tmp/semaphore-stand/shots-docs`; convert with `cwebp -q 88` into `docs/static/assets/`. Traps: a workflow must
have exactly one root node (the explicit edge alone would make Provision and Build two roots); key hints need a
run of the current revision, so re-run after any re-seed; the editor fits a three-node graph at ~45 % on a 1440 px
viewport — the script collapses the palette (`.WorkflowEditor__sideToggle`) and clicks the fit control after the
side panel opens instead of zooming in, which pushes nodes under the panel.

Seeding traps learned there: a Terraform-family template must be POSTed without `inventory_id`
(the API creates a `terraform-workspace` inventory; a shared one gets claimed and hidden) and with
`task_params.auto_approve`, or a plan with changes parks the task in `waiting_confirmation`; a
workflow must have exactly one root node; `survey_vars[].default_value` is a plain string; a JSON
`\"` inside `$( )` inside a double-quoted bash string is mangled — build such fragments in a
variable first.

Traps: `run.sh` exits 1 when `--expect` does not match or `--timeout` (default 180 s) passes;
`--approve` resolves every pending approval it sees, so use `--run ID` to only follow an existing
run. `stand.sh build` rebuilds the Vue app into `api/public` (a tracked-ignored embed input) and
the binary at `/tmp/semaphore-stand-bin`; it needs the `pro_impl` workspace for Workflows to be
enabled. The stand DB is `/tmp/semaphore-stand/database-wf.sqlite` and persists between sessions.

`runner.sh` registers a global runner "wf-runner" (tag `wf`, local executor, `/tmp/semaphore-stand/runner-config.json`,
log `runner.log`) and starts it from the same binary; `tag` sets `runner_tag` on every template of project
`PID` (default 1) so their tasks go to it, `untag` sends them back to the server. The remote-runner check
of workflow outputs (2026-10-08): `runner.sh start && runner.sh tag && run.sh 1 --expect success && run.sh 2
--expect failed`, then `sqlite3 /tmp/semaphore-stand/database-wf.sqlite "select id,status,runner_id,length(artifacts)
from task order by id desc limit 13"` shows `runner_id` of the runner and `artifacts` on the producers only.
Traps: the token is returned once at creation, so a runner whose config file is gone is deleted and recreated;
stop the runner and `untag` before a session that expects tasks to run on the server, or they wait forever;
`status` keeps saying `online` for up to 120 s after `stop` (the server's heartbeat timeout), trust the
`process:` line.

`seed-survey.sh` (2026-10-16, TASK@B197C72B7F) needs the templates of `seed-outputs.sh` and adds workflow
"Survey: workflow variables": survey `env` (enum, default `staging`), `image_tag` (required), `replicas` (int,
default 2); node 2 "bash: consume" bound `image_tag←image_tag`, `replicas←replicas`, `env_name←env` on top of
static `image_tag=static-from-node`; node 3 the same template without bindings (must receive nothing although
the names match); node 4 "bash: produce renamed outputs" → node 5 "bash: consume" over an explicit edge
`replicas←count` plus the binding `image_tag←image_tag`. Start a run with a body:
`POST …/workflows/<id>/runs {"environment":{"image_tag":"1.2.3","env":"prod"}}` (ids in
`/tmp/semaphore-stand/wf-survey-ids.txt`). `shoot-editor.cjs` logs in, opens the editor, clicks the first card
whose title matches, scrolls the first bound survey field into view and writes the page, a crop of the task
params card (`<out>-fields.png`) and the open assign menu (`<out>-assign.png`). `shoot-run-dialog.cjs` opens the
workflow page, presses Run, checks that an empty required answer is refused, fills `image_tag` (`IMAGE_TAG`,
default 1.2.3) and the first enum option, starts the run and shoots `run-dialog-invalid.png`, `run-dialog.png` and
`run-view-variables.png` in `out_dir`; it prints the run id and the Variables chips.

⚠️ 2026-10-16: macOS purges files (not directories) under `/tmp` that were not touched for ~3 days —
`config-wf.json`, `database-wf.sqlite`, the admin user and every script in `wf-repo` vanished between sessions.
`stand.sh build` + a fresh `config-wf.json` (sqlite at `$DIR/database-wf.sqlite`, port `:3100`, three
`openssl rand -base64 32` keys, `use_remote_runner: true`), `migrate`, `user add --admin --login admin
--password admin123` and the seeds bring it back; `wf-repo` scripts must be recreated before any run.
