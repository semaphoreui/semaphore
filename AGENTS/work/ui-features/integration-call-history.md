# Integration call history (PROD-10)

The Integrations page has a **History** tab listing every webhook call the project received, newest
first, filterable by integration and by alias. Linear PROD-10 asked for "the log of calls of each
integration alias"; Denis refined it to one project-wide page, because an alias can be deleted and
its history must survive.

Status (2026-10-21): implemented on branch `prod-10-logs-for-intergrations`, not yet in a PR.

## Decisions and why

- **Its own table, not `audit_event`.** A webhook endpoint is public and high-volume; the audit log
  is a security record with its own retention sweeper (`services/audit.Service.stopRetention`), so
  webhook traffic would both swamp it and be pruned by it. Filtering by alias also needs an indexed
  column, not a JSON `metadata` probe that differs per SQL dialect.
- **One row per (call, integration).** `GetIntegrationsByAlias` returns a list — a project-level
  alias fans one request out to several integrations — and "filter by integration" only means
  something if each one has its own row.
- **Alias and integration name are plain strings; `integration_id` is nullable (`on delete set
  null`), `task_id`/`run_id` carry no foreign key at all.** That is the whole point of the feature:
  history outlives the alias, the integration and the task.
- **Only calls whose alias resolves are recorded.** `/api/integrations/<anything>` is
  unauthenticated, so recording unresolved aliases would let anyone flood the table from the
  internet. Internet noise against a non-existent alias stays a log line; a *known* alias that
  fails HMAC, matches no matcher, or starts nothing at all is recorded — that is the case a user
  actually debugs. `GetIntegrationsByAlias` therefore returns the alias owner's `projectID` even
  when it matched no integration, so that call can still be attributed to a project.
- **Keyset pagination**, reusing `parseTasksPageParams` and the `X-Has-Next` convention — see
  `history-keyset-pagination.md`.
- **One row per (call, integration), including failures.** A per-integration failure `continue`s
  rather than returning, so one integration with an unloadable secret neither hides the others from
  the history nor stops them running — the pre-review code returned and aborted the whole fan-out.
- **Every column written from the request is truncated on a rune boundary** (`truncateUTF8`).
  `handlers.ProxyHeaders` rewrites `RemoteAddr` from `X-Forwarded-For` without validating it
  (`cli/cmd/root.go:371`), so `source_ip` is caller-controlled and unbounded; a reason carries an
  error message that may be non-ASCII. A cut inside a multi-byte rune leaves invalid UTF-8, which
  PostgreSQL rejects — and `recordCall` does not retry, so the row would be lost silently.
- **Failure reasons are plumbed, not re-derived.** `IntegrationRunResult.Reason` carries why nothing
  started ("Integration has no template", "Workflows are not available in this build", …) from the
  run helpers to the recorded row.
- HA: a plain insert on the shared DB, read back by any node. No new authz surface — the list is on
  `projectUserAPI`, the same visibility as the integrations themselves.

## Open, broken or deferred

- ⚠️ 2026-10-21: **no retention.** The table grows without bound, like `task` does. Ask Denis
  whether calls should expire; a sweeper would fit next to the audit one.
- Request bodies and headers are deliberately **not** stored — a webhook payload carries secrets.
  Only metadata is kept. Raise it with Denis if he wants payloads.
- The auth switch in `ReceiveIntegration` compares the token with `!=`, not a constant-time compare.
  Pre-existing, untouched here, worth its own fix.
- A call whose alias lookup *errors* (e.g. a dedicated alias on a `searchable` integration, which
  `db/sql/integration_alias.go:87` rejects with `ErrNotFound`) records nothing: at that point no
  project is known. It stays a log line.
- The alias filter is exact match on an indexed column; aliases are random strings, so the alias
  in each row is clickable to apply it. No substring search.
- Filter changes start a load without awaiting the one in flight, so each request carries a
  sequence number and only the newest applies its response (`requestSeq` in `IntegrationCalls.vue`).
- `projectId` is a prop, so switching project reuses the component and `created()` does not run
  again: the watcher clears both filters and the rows, and reloads the integration options, or they
  would still belong to the previous project. The integration select reloads on `@change`, not a
  watcher, so clearing it in code does not fire a second request.
- The call list and the integration options are fetched independently (`Promise.all`), so a failed
  integration lookup cannot leave the previous project's calls on screen, and each has its own
  sequence counter: A -> B -> A gives overlapping requests with the same project id, so the id
  cannot decide which response is current.
- The ngrok reference also shows Path and Host. Path is `/api/integrations/<alias>`, which the
  alias column already carries; Host is not stored.
- The user docs (`docs/` submodule) are not checked out in this workspace, so the History tab is
  undocumented for end users.

## Key files and references

- `db/sql/migrations/v2.20.17.sql` (+ `.err.sql`), registered in `db/Migration.go`
- `db/IntegrationCall.go`, `db/sql/integration_call.go`, `db/Store.go` (`IntegrationManager`)
- `api/integration.go` (`recordCall`, `sourceIP`, `verifyIntegrationAuth`, `IntegrationRunResult.Reason`)
- `api/projects/integration_call.go`, route in `api/router.go` (`/integrations/calls`, kept out of
  the `{integration_id}` subrouter)
- `db/sql/integration_alias.go` (`GetIntegrationsByAlias` returns `projectID`)
- `web/src/views/project/IntegrationCalls.vue`, tabs in `Integrations.vue`, route in
  `web/src/router/index.js` (above `:integrationId`)
- Tests: `db/sql/integration_call_test.go`, `api/projects/integration_call_test.go`,
  `TestReceiveIntegration_Records*` in `api/integration_test.go`
