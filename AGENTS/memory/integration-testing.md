# integration-testing — SMTP, Dredd and the Bookwright repo

## SMTP integration tests (`util/mailer/integration/`)

- Behind `//go:build integration`, run with `task test:be:integration` (CI job `test-integration`
  in `dev.yml`). Testcontainers + Mailpit `axllent/mailpit:v1.28`. Plan on branch
  `feat/project-alerts-mail-refactor`: `AGENTS/plans/2_20/mailer-smtp-integration-tests.md`.
- Mailpit takes auth via env `MP_SMTP_AUTH="user:pass"`, not a flag. Its `Received:` header says
  `with SMTP` even over STARTTLS — TLS is proven by the `Username` field (AUTH is advertised only
  after TLS unless `MP_SMTP_AUTH_ALLOW_INSECURE`).
- The container bridge IP is not routable from the host on macOS, so scenario E2 (own SMTP host
  through `guardedDialContext`) is Linux-only and skips locally — do not chase it.
- testcontainers-go v0.44: `wait.ForListeningPort`/`WithPort`/`MappedPort` take plain strings; the
  Docker client is `github.com/moby/moby/client` with option structs.
- `mailer.Options.RootCAs` exists only so tests can trust a private CA; no user-facing
  `email_tls_ca_file` yet.
- License regen: `go-licenses` in `~/go/bin`; `license-checker` is not installed — a PATH shim
  `exec npx --yes license-checker "$@"` makes the collect script work. Test-only deps behind a
  build tag never appear in `THIRD-PARTY-LICENSES.md`.

## Dredd contract tests (`.dredd/`)

- Alert API hooks in `.dredd/hooks/main.go` (fixtures in `capabilities.go`/`helpers.go`).
- `.dredd/config.json` is gitignored and CI does not use it: the integrate-* jobs generate config
  via `semaphore setup` fed from a `config.stdin` heredoc answering "no" to every channel, so the
  CI server has no `telegram_token`. A local token masks failures that still hit CI.
- Alert create/update fixtures send a Slack body: Slack's `Secret()==nil`, so `clearUnusedFields`
  nulls `key_id` and `ValidateAlert` only checks the webhook URL. Telegram/Gotify would need a
  `string`-type project key, and `store.CreateAccessKey` inserts the already-encrypted `Secret`.
- `body_format` enum in `api-docs.yml` is `[text, html]`; a CI failure mentioning `json` or
  `server_enabled` is a stale binary.
- ⚠️ 2026-09-27: `task dredd:docker` and the `server` service of `deployment/compose/dredd/base.yml`
  exist only on `feat/project-alerts` (commit `43e5054b8`), not on `develop` or
  `feat/refactor_workflows`, where the compose fails with "service server has neither an image
  nor a build context". Cherry-pick `43e5054b8` (or merge that branch) to get the Docker runner.
  What CI actually does: `task dredd:hooks` (needs the Pro workspace) and Dredd against an OSS
  server, where `isProBuild` skips every workflow endpoint. Locally, `task dredd:hooks` plus a
  `js-yaml` parse of `api-docs.yml` is the cheap check. `.dockerignore` excludes `go.work` and
  `pro_impl` so images build the OSS tree.
- Local run without Docker: `docker run postgres:15` matching `.dredd/config.json` (db
  `semaphore_test_000071`, user `postgres`, pass `pwd`, port 5432), `npm install dredd@13.1.2` in
  `web`, `go install .../goodman`, build `bin/semaphore` and `.dredd/compiled_hooks` with
  `GOWORK=off`, then `./web/node_modules/.bin/dredd --config .dredd/dredd.testing.yml`. The
  dredd↔goodman startup is racy (ECONNREFUSED 61321 → every txn 401): retry.
- Local Dredd recipe that works on this Mac (2026-10-14): the `my-postgres` Docker container already
  matches `.dredd/config.json`; drop/create `semaphore_test_000071` before each run (fixtures assume
  a fresh DB), `GOWORK=off go build -o bin/semaphore ./cli` and the hooks, `PATH=$HOME/go/bin:$PATH`
  for goodman, then `dredd --config .dredd/dredd.testing.yml` (server-wrapper starts the server on
  58427 itself). 54 skips (ws, login, OIDC, 400s, Pro workflow endpoints) are the normal baseline.
- A field the server nulls (e.g. `Inventory.repository_id` for non-file types since `6a4a04ba`)
  must be `x-nullable: true` in `api-docs.yml`; the openapi2 parser Dredd uses honours it.
- Pro Dredd run (2026-10-15, exercises the 14 workflow endpoints the OSS run skips): build server
  and hooks *with* `go.work` (`go build -o bin/semaphore-pro ./cli`, hooks to `compiled_hooks_pro`),
  point a copy of `dredd.testing.yml` at them with absolute paths (keep it outside the repo) and
  a wrapper that `exec`s `bin/semaphore-pro`; same DB recreate and `ECONNREFUSED 61321` retry.
  Baseline: 127 passing, 0 failing, 32 skipped. A worktree without `web/node_modules` can use
  the main checkout's `web/node_modules/.bin/dredd` and its `.dredd/config.json`.
- The dredd↔goodman race also shows as `complete: 2 passing, 130 failing … 27 skipped` with every
  transaction 401 and no `ECONNREFUSED` in the log (2026-10-16): retry until the summary says
  `0 failing` instead of grepping for the error. A `run_in_background` Bash inherits whatever cwd the
  session last had — use absolute paths for `go build ./cli` and the Dredd config.

## Bookwright `integration-tests/` repo

- Separate untracked git repo inside the checkout: Bookwright v1.4.0 (Java 21, JUnit 5,
  Retrofit, Guice, Allure). Overview research: `RESEARCH@84146f310f`.
- It builds the app image from `develop` HEAD (or an `Application-PR:` line in the test PR) via
  `scripts/app-source.sh`; a local branch cannot be resolved. Build the image yourself:
  `docker buildx build --file deployment/docker/server/Dockerfile --tag semaphore-test/app:<tag>
  --provenance false --load .` from a shallow `git clone file://<repo>` in /tmp (Taskfile runs
  `git name-rev`, and the build must not see `go.work`; the working-tree context is ~4 GB), then
  `APP_IMAGE=semaphore-test/app:<tag> test-environment/profile up <profile>` /
  `profile test <profile> [gradle args]` (`--tests io.bookwright.tests.semaphore.SomeTest`).
- Manifests may set `test_class` (comma-separated). To run locally, change `"3000:3000"` in
  `test-environment/compose.base.yml` and `readiness_url`/`api_base_url` to a free port (3100),
  revert afterwards.
- Architecture tests when adding a domain: `FixtureArchitectureTest` (scenario literals only in
  fixtures; secret-bearing fixtures redact in `toString`), `DomainArchitectureTest` (Api
  interfaces ≥3 levels under `api/`), `StateOwnershipArchitectureTest` (TestStore public
  methods). New fixtures → `FixtureCatalog`; new APIs → `ApiModule` + `SemaphoreSteps`
  (+ `SemaphoreSessionApis`).
- `./gradlew frameworkTest` includes `BrowserContextIsolationTest`, which needs a Playwright
  Chromium download and times out offline. Run `./gradlew spotlessApply` before `qualityGate`.
  CHANGELOG bullets must not start with the section word.
- Inventories used with host-config mappings must put host-key options in
  `ansible_ssh_extra_args`: inventory `ansible_ssh_common_args` overrides the
  `--ssh-common-args -F <generated ssh config>` the executor passes.
- Known red (2026-09-24): `HostConfigSshTaskTest.urlMappingRewritesHttpsRepositoryToSsh` on PR
  sem-272, because `urlRewrite` in `pkg/ssh/host_config.go` hard-codes `git@` and overrides the
  `User` written to the generated ssh config.
