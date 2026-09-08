# Semaphore UI - Agent Instructions

See `.github/copilot-instructions.md` for comprehensive development instructions including build commands, project structure, and troubleshooting.

## Cursor Cloud specific instructions

### Environment basics

- **Go 1.24.6** and **Node.js 22+** are available system-wide.
- **Task runner** (`task`) is installed at `~/go/bin/task`. Ensure `~/go/bin` is on `PATH`.
- Dependencies are vendored (`vendor/` for Go, `web/node_modules/` for frontend) after running `task deps`.

### Running the server

1. Build: `task build` (builds frontend into `api/public/` then compiles Go binary to `bin/semaphore`)
2. If `config.json` does not exist, create one with SQLite dialect:
   ```bash
   ./bin/semaphore user add --admin --login admin --name Admin --email admin@example.com --password changeme --config ./config.json
   ```
   (This also runs migrations and creates the DB file.)
3. Start: `./bin/semaphore server --config ./config.json`
4. Verify: `curl http://localhost:3000/api/ping` → `pong`
5. Web UI at http://localhost:3000 (login: `admin` / `changeme`)

If the config.json already exists, just start the server directly.

### Key commands

| Action | Command | Timeout |
|--------|---------|---------|
| Install deps | `task deps` | 5 min |
| Build all | `task build` | 3 min |
| Backend tests | `task test:be` | 2 min |
| Frontend lint | `cd web && npm run lint` | 1 min |
| Ping server | `curl http://localhost:3000/api/ping` | — |

### Gotchas

- The `task` binary is installed via `go install` into `~/go/bin/`. If `task` is not found, run: `go install github.com/go-task/task/v3/cmd/task@latest` and ensure PATH includes `~/go/bin`.
- Frontend lint (`npm run lint`) has **pre-existing errors** (webpack import, template-root in AboutDialog.vue, etc.). These are known; do not attempt to fix unless specifically asked.
- Backend lint (`golangci-lint run`) also has pre-existing issues due to module import patterns — ignore existing issues.
- The Go binary serves the built frontend from `api/public/`. There is no separate frontend server needed for production-style testing. Use `cd web && npm run serve` only when you need Vue hot-reload during frontend development (proxies API to `:3000`).
- SQLite is the simplest database for development (no external service needed). Use `"dialect": "sqlite"` in config.json.
- `task deps` runs `go mod vendor` which can take a while on first run — never cancel it.
