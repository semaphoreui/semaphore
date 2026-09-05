# AGENTS.md

## Cursor Cloud specific instructions

Semaphore UI is a single Go binary (`bin/semaphore`, entry `cli/main.go`) that serves a
Vue 2 frontend (`web/`, built into `api/public/`) plus a REST API + task orchestrator on
port `3000`. SQLite is the default DB and needs no separate service. See `README.md`,
`CONTRIBUTING.md`, and `Taskfile.yml` for canonical commands.

### Toolchain / environment (already provisioned by the snapshot + update script)
- Base `go` is 1.22.2 but the module targets Go 1.26; `go` auto-downloads `1.26.4` via the
  toolchain directive, so plain `go build` / `go test` work.
- Dev tools are installed under `$(go env GOPATH)/bin` (`~/go/bin`), which is added to
  `PATH` via `~/.bashrc`: `task` (go-task), `golangci-lint`, `goreleaser`.
- Non-obvious: `golangci-lint` must be built with Go 1.26+ or it fails with
  "Go language version (go1.25) ... is lower than the targeted Go version (1.26.4)". The
  snapshot copy is built correctly; if you ever reinstall it, use
  `GOTOOLCHAIN=go1.26.4 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

### Build
- Full build: `task build` (`build:fe` = `npm run build` in `web/` → `api/public/`;
  `build:be` = `go build ... -o bin/semaphore ./cli`).

### Lint / test
- Lint: `task lint` (frontend `npm run lint`, backend `golangci-lint run` + `swagger validate`).
  Note `swagger` is not installed, so run `golangci-lint run` directly instead of `lint:be`.
  `.golangci.yml` sets `issues-exit-code: 0`, so `golangci-lint` exits 0 even with findings;
  the frontend lint currently reports pre-existing errors. Neither is caused by your changes.
- Tests: `go test ./...` (or `task test`). Frontend unit tests are disabled in the Taskfile.

### Run the server (dev)
1. A SQLite `config.json` lives at the repo root (gitignored, like `database.sqlite` and
   `bin/`). If absent, create one with `"dialect": "sqlite"`, a `sqlite.host` path, and
   `cookie_hash` / `cookie_encryption` values (see `.devcontainer/config.json`).
2. Seed an admin once (runs migrations automatically):
   `./bin/semaphore user add --admin --login admin --name Admin --email admin@example.com --password changeme --config ./config.json`
3. Start: `./bin/semaphore server --config ./config.json` → http://localhost:3000
   (default login `admin` / `changeme`).
- Running tasks against real infra needs the relevant CLI (Ansible is installed; Terraform/
  OpenTofu/PowerShell are not). The server has an embedded local runner; a separate
  `semaphore runner` is only needed for remote-runner flows.
