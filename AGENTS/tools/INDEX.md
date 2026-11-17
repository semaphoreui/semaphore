# tools — index

Working tools of the agent for this project: scripts only the agent runs to carry out an
operation — database queries, audits, probes, one-off checks, and the like. One folder
per tool, each carrying its own `README.md` with full usage. This file is the registry that lets a
tool be found without listing the tree first.

Not to be confused with the harness's built-in tools (Bash, Read, MCP…). These are project-local
scripts checked into the repo.

How to add, document and retire a tool — [`CREATE.md`](CREATE.md).

## Inventory

### `wf-stand/` — run and verify workflows on the local stand

`stand.sh` starts/stops/builds the `:3100` stand, `seed.sh` creates known-outcome bash
workflows, `run.sh` starts a run and follows every node (task, approval, delay) to a terminal
status with optional auto-approve/stop, `runner.sh` registers and runs a remote runner for the
stand and routes a project's templates to it, `inputs.sh` prints producer outputs and consumer
`INPUT` lines of a run, `shoot-run.cjs` screenshots the run view, `check-run-panel.cjs` walks the run panel (`?w=<workflow>/<run>`)
in a browser and prints PASS/FAIL per check, `seed-docs.sh` +
`shoot-docs.cjs` seed the "Demo Stand" project and shoot the screenshots of the user guide's
Outputs and inputs section. Use it instead of
hand-written curl polling whenever a change to the workflow engine or run view must be checked
against a real run.

### `help-stand/` — check the built-in help panel on a real binary

`stand.sh` builds the community binary, starts it on `:3100` with a second instance under a
sub-path on `:3102` and seeds inventories of every kind; four Playwright checks drive the help of
the Ansible inventory dialog and of the project Settings page: requests to the server only and
work with other hosts blocked, the panel following the focus, the field marked by the active
section, disabled sections, the layout of a page next to the panel, and ten dialogs of every kind
in any language. Use it after any change to
`EditDialog.vue`, `HelpPage.vue`, `HelpPanel.vue`, `HelpDocsHost.js` or `web/vite/help` — unit
tests run in jsdom, which lays nothing out and has no focus trap.

### `redis-sentinel-stand/` — prove HA failover on Redis Sentinel

`stand.sh` builds the Pro binary for linux, brings up Redis master + replica + Sentinel, PostgreSQL
and two Semaphore HA nodes (`:3110`, `:3111`) in docker compose, kills and restores the master;
`seed.sh` creates bash templates, `run.sh` starts a task on one node and follows it through the
other, `wsprobe.go` prints the WebSocket events a node receives. Use it whenever the HA Redis
connection, `ha.engine` / `ha.redis_sentinel` or the Pub/Sub broadcaster change — unit tests run
on miniredis, which has no Sentinel.
