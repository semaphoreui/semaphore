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
stand and routes a project's templates to it, `shoot-run.cjs` screenshots the run view. Use it instead of
hand-written curl polling whenever a change to the workflow engine or run view must be checked
against a real run.
