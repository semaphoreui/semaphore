# AGENTS.md

## Cursor Cloud specific instructions

Semaphore UI is a single product: a Go backend (REST API + WebSockets, serving the
built Vue.js SPA) plus a Vue 2 frontend. The same binary runs the built-in task runner,
so no separate runner process is needed for local development. Build orchestration uses
[go-task](https://taskfile.dev) via `Taskfile.yml` (there is no Makefile).

Standard build/lint/test commands live in `Taskfile.yml` and `CONTRIBUTING.md`; prefer
those over duplicating them. Notable ones: `task build` (builds frontend into
`api/public` then the `bin/semaphore` binary), `task lint:fe`, `task test` / `task test:be`.

### Startup caveats (non-obvious)

- The `task` binary is installed to `$(go env GOPATH)/bin`, which is not on `PATH` by
  default. Either run it as `$(go env GOPATH)/bin/task ...` or add that dir to `PATH`
  first. `bin/semaphore`, `vendor/`, `web/node_modules/`, and `api/public/` are all
  git-ignored build artifacts; regenerate them with `task build` if missing.
- The server needs a config file and at least one admin user, which are NOT created by
  the update script (`config.json` and `database.sqlite*` are git-ignored). If they are
  missing, create a SQLite dev config and seed the admin user before starting the server:
  ```bash
  cat > config.json <<'EOF'
  {
    "sqlite": { "host": "/workspace/database.sqlite" },
    "dialect": "sqlite",
    "cookie_hash": "5WJjXCLpvf3Cn5t+C/IV9F0asZUQLakOhCT+eSdIwP0=",
    "cookie_encryption": "6x6mmQWGn6YcsHN1rN0HiQjhYA+7HukcbCxUGHuT2CE=",
    "port": ":3000"
  }
  EOF
  ./bin/semaphore user add --admin --login admin --name Admin \
    --email admin@example.com --password changeme --config ./config.json
  ```
  The `user add` command also runs DB migrations. Default dev login: `admin` / `changeme`.
- Run the server (serves API + UI at http://localhost:3000):
  ```bash
  ./bin/semaphore server --config ./config.json
  ```
  For active UI work with hot reload, additionally run `cd web && npm run serve`
  (Vue dev server on :8080, proxies `/api` to :3000).
- Use SQLite for local dev (BoltDB was removed in 2.19). MySQL/PostgreSQL are for
  production and require a running DB server.
- `task lint:be` needs `golangci-lint` + `swagger`, which are intentionally NOT installed
  by `task deps` (commented out in `Taskfile.yml`). `task lint:fe` currently reports
  pre-existing lint errors in the repo; they are not caused by environment setup.
- Running real Ansible/Terraform task templates requires those CLIs; the devcontainer
  installs Ansible into a `.venv`. The "demo project" auto-populates runnable resources
  (e.g. a "Ping" Ansible template) useful for a quick end-to-end smoke test.
