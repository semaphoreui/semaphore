# tools — index

Working tools of the agent for this project: scripts only the agent runs to carry out an
operation — database queries, audits, probes, one-off checks, and the like. One folder
per tool, each carrying its own `README.md` with full usage. This file is the registry that lets a
tool be found without listing the tree first.

Not to be confused with the harness's built-in tools (Bash, Read, MCP…). These are project-local
scripts checked into the repo.

How to add, document and retire a tool — [`CREATE.md`](CREATE.md).

## Inventory

Empty — no project-local tool has been added yet. The entries below are example data,
illustrating the shape a real entry takes; replace them once an actual tool exists.

### `db-query/` — ad-hoc SQL runners (example)

`query_<db>.sh` — one script per database, `docker exec`s into the DB container and runs SQL as
that database's least-privilege app user, so no client is needed on the host. SQL comes as a
positional string, `-f FILE`, or stdin. Credentials are read out of the app's own `.env`, never a
copy kept in the tool.

### `paddle-probe/` — raw calls to a provider's API, sandbox only (example)

`probe.sh <METHOD> <PATH> [body.json]` — a probe against the sandbox; the response prints as-is,
status to stderr, body as plain JSON to stdout. The base URL is derived from the key, and the
script refuses to run against a live key.
