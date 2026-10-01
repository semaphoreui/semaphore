# wf-stand — run and verify workflows on the local stand

Drives the agent's throwaway Semaphore stand (`http://localhost:3100`, admin/admin123, project 1,
Pro build from HEAD) to execute a workflow end to end and follow every node to a terminal status.
Exists because a workflow run is asynchronous: tasks, approvals and delays resolve over time and
only the run-details endpoint shows all of them together.

Call forms (from the repository root):

```
AGENTS/tools/wf-stand/stand.sh {start|stop|restart|build|status|log [n]}
AGENTS/tools/wf-stand/seed.sh
AGENTS/tools/wf-stand/run.sh <workflow_id> [--approve|--reject] [--stop-after SEC] [--expect success|failed|stopped] [--timeout SEC] [--run ID]
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

Traps: `run.sh` exits 1 when `--expect` does not match or `--timeout` (default 180 s) passes;
`--approve` resolves every pending approval it sees, so use `--run ID` to only follow an existing
run. `stand.sh build` rebuilds the Vue app into `api/public` (a tracked-ignored embed input) and
the binary at `/tmp/semaphore-stand-bin`; it needs the `pro_impl` workspace for Workflows to be
enabled. The stand DB is `/tmp/semaphore-stand/database-wf.sqlite` and persists between sessions.
