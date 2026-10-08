# memory — index

Loaded every session, so it stays small. A short fact lives here as one or two lines. Facts that
cluster around one topic live in a topic file in this folder, listed in the table below and read
only when the task touches that topic. Keep this file under 150 lines.

## Rules

- Every solution must work in High Availability mode and be secure by design; no global variables
  in Go code — these are the owner's standing constraints for this repo.
- Never run tasks on the public demo stand `https://semaphore.orbantix.com` (login `demo`/`demo`,
  role task_runner) — it is public; details in `ui-stands.md`.
- Before repeating a release-notes claim, `git grep` the tag tree (`git grep <term> vX.Y.Z -- '*.go'`) —
  the published v2.19.7 GitHub notes were written from develop and advertise 2.20-only features
  (Prometheus `/api/metrics`, SSH host-key checking, `ldap_tls_skip_verify`, task-bound survey
  secrets, embedded docs, Argon2id, runner token hashing, runner version/platform columns).
- Template form UX (agreed 2026-09-09): hide only fields most users leave at default, one labelled
  inline toggle per hidden group (not accordion, not tabs), auto-expand on a non-default value or
  validation error. `TemplateForm.vue` already has two reveal patterns (chip → card, checkbox →
  `DropdownCard`); `web/src` uses no `v-expansion-panels`.
- Verify a claim about Chromium wheel/scroll behaviour against the facts in
  `AGENTS/work/workflow-editor/` before changing wheel handling in the workflow editor — three
  rounds of fixes were built on guesses.

## Gotchas

- Parse a cron format with `schedules.ParseCronSchedule`, never `cron.ParseStandard` or
  `Cron.AddJob`: robfig/cron v3.0.1 panics on a `TZ=`/`CRON_TZ=` prefix with nothing after it.
- `default:` struct tags in `util/config.go` are applied at config load (when the section exists),
  so they — not in-package fallback constants — are the effective documented defaults.
- `config.schema.yaml` is not a source of truth; it drifts (gave wrong k8s executor image
  defaults). Generate reference docs from the Go struct with `task docs:gen`.
- The `docs/` submodule is checked out by no CI job except `docs-reference` in
  `.github/workflows/dev.yml` — a build step that reads `docs/` must fail loudly when it is absent.
- Anything the Vue build emits to `api/public` is `go:embed`-ed into the binary and served by
  `serveFile()`; ship static content that way, not through a new Go handler. Use relative paths
  (`publicPath: './'`) — `serveFile()` rewrites `<base href>` from `util.WebHostURL`.
- `go.mod` has `replace .../pro => ./pro`; `go.work` and `pro_impl/` are gitignored but present
  locally. Any Dockerfile layer running `go mod download` must also COPY `pro/go.mod` + `pro/go.sum`;
  build images from a shallow `git clone file://<repo>` so they never see `go.work`.
- `go mod why` in workspace mode blames `pro/services/tasks/docker`; use `GOWORK=off` to see the
  real dependency chain. A version-specific `replace` (not a wildcard) is used to pin grpc, because
  a wildcard downgrades the Pro workspace.
- `integration-tests/` inside the checkout is a separate untracked git repo (Java/Gradle,
  Bookwright); see `integration-testing.md`.
- Docker Desktop on this Mac is not running by default: `open -a Docker` starts it in ~2 s.
- The shell `grep` is ugrep and silently returns 0 matches on Cyrillic patterns in built HTML;
  verify locale pages with Python.
- Port 3000 is usually taken by the Docusaurus dev server; 8080 by the dev stand; local stands use
  3100 (see `ui-stands.md`).
- Research MCP: `query_search_run` returned zero sources on several days (2026-09-08/09) and many
  fetches return `error`; check `sources_list` before assuming a search worked, and expect to fall
  back to WebSearch/WebFetch with URLs cited in notes.
- Workbench MCP: every session starts unbound and can silently lose the binding mid-session —
  re-run `workspace_use` on the first "not bound" error.
- Workbench MCP: `task_update` refuses to change context/constraints/criteria of a human-created
  task (only type, priority, group/parent change) — propose a brief via `note_add(type="decision")`.
- Workbench MCP: body limits refuse instead of trimming — task body 8192 chars, note body 2048;
  stages exist only on `extended` tasks and have their own bodies, so move detail there.
- Subagents on the Fable model hit 403 twice during the docs audit and had to be rerun on Opus.
- `docs.cyberark.com` returns 404 to WebFetch/curl for every page; use the Ansible collection
  `cyberark/ansible-security-automation-collection` and the Go client `infamousjoeg/cybr-cli`
  as the PVWA API reference.
- Pre-2026-09-27 plan, task and research-digest files (`AGENTS/plans/…`, `AGENTS/tasks/…`,
  `AGENTS/research/*.md`) cited as sources in `AGENTS/work/` docs were removed from the tree; read them
  with `git show 539d98ad:<path>`. Some plans exist only on feature branches (`feat/cyberark-secret-storage`,
  `feat/project-alerts-mail-refactor`).

## Topic files

| When you need it | Read |
| --- | --- |
| Editing or building the docs submodule: build/verification commands, webpack cache trap, generated reference pages, locale rules, MDX/image conventions, where executor docs live, docs research pointers | `docs-site.md` |
| Spinning up a throwaway Semaphore stand, shooting UI screenshots with Playwright, using the public demo stand, running a hot-reload dev server against a stand | `ui-stands.md` |
| Running or extending integration tests: SMTP/Mailpit tests, Dredd alert hooks and local Dredd runs, the Bookwright `integration-tests/` repo | `integration-testing.md` |
