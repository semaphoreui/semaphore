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
NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/wf-stand/shoot-run.cjs <workflow_id> [run_id] [out.png] [--dark]
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
