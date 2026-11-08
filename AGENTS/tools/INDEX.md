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
`INPUT` lines of a run, `shoot-run.cjs` screenshots the run view, `seed-docs.sh` +
`shoot-docs.cjs` seed the "Demo Stand" project and shoot the screenshots of the user guide's
Outputs and inputs section. Use it instead of
hand-written curl polling whenever a change to the workflow engine or run view must be checked
against a real run.

### `redis-sentinel-stand/` — prove HA failover on Redis Sentinel

`stand.sh` builds the Pro binary for linux, brings up Redis master + replica + Sentinel, PostgreSQL
and two Semaphore HA nodes (`:3110`, `:3111`) in docker compose, kills and restores the master;
`seed.sh` creates bash templates, `run.sh` starts a task on one node and follows it through the
other, `wsprobe.go` prints the WebSocket events a node receives. Use it whenever the HA Redis
connection, `ha.engine` / `ha.redis_sentinel` or the Pub/Sub broadcaster change — unit tests run
on miniredis, which has no Sentinel.

### `agent-stand/` — check the AI agent UI on a throwaway stand

`stand.sh` builds this work copy and starts it on `:3102` with its own SQLite, `seed.sh` adds an
mcp inventory and Claude Code / Codex templates, `verify.cjs` saves through the UI and checks the
API and task logs, `shoot.cjs` takes the screenshots. Agents run a fake CLI unless started with
`--real`. Use it after changing the mcp inventory form, the agent fields of the template form or
the task prompt; unlike `agent-cli/`, it goes through the server and the browser.

### `agent-cli/` — run the real Codex / Claude Code CLI like AgentApp

`probe.sh codex|claude` writes the same per-task config as `db_lib/AgentApp.go`, sends the
prompt on stdin, and starts a fake MCP server that shows whether the bearer token arrived. API keys
come from the tool's git-ignored `.env`. Use it after changing how AgentApp builds the CLI
config or flags, or to check that a test API key still works. Unit tests use a fake binary and
cannot catch a CLI that rejects the config.
