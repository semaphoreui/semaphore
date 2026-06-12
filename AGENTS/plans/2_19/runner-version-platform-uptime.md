# Plan — Show Runner Version, Platform and Uptime

## Goal

Make the runners page actually useful for diagnosing a fleet. Today the
runners table tells the operator that a runner is "alive" (via the `touched`
timestamp) and not much else. When something goes wrong — a task hangs on a
specific runner, a runner stops picking up jobs after a server upgrade, a
runner silently runs an outdated binary — there is no way to triage from the
UI. This plan adds three concrete pieces of information that answer the most
frequent questions an operator has:

- **What version is this runner running?** (matters during/after server
  upgrades, when the wire protocol changes)
- **Where is it running?** (OS and CPU architecture — matters for picking the
  right runner for a job, and for diagnosing platform-specific failures)
- **How long has it been up?** (distinguishes "freshly restarted, lost local
  state" from "stable for weeks")

## Scope

In scope:

- Runner reports its **version**, **OS**, **architecture** and **start time**
  to the server as part of the existing polling protocol.
- Server persists the latest reported values per runner and exposes them in
  the runner list / details API.
- Runners table on the frontend gains three new cells: Version, Platform,
  Uptime. Uptime is rendered relative to "now" on the client.

Out of scope:

- Hostname, IP address, container ID, kernel version, libc, Ansible/Tofu
  binary versions installed on the runner. Useful but a separate decision —
  every additional field is a new piece of state to maintain and a small
  privacy surface.
- Resource usage (CPU/RAM/disk). Genuinely useful but requires sampling and
  a different storage model; tracked as a follow-up.
- Historical version timeline / audit log. Only the *current* values are
  stored.
- Any change to how tasks are routed. Platform is informational only; the
  task scheduler does not (yet) consider it.

## Design Summary

1. **Protocol.** The runner already calls `GET /api/runner` on every poll
   (handled by `RunnerController.GetRunner` in `api/runners/runners.go`).
   That call is the natural place to attach the metadata. The runner sends
   `version`, `os`, `arch` and `started_at` as request headers (or, if a
   request body is preferred, in a small JSON payload on the existing
   `UpdateRunner` progress call). Headers are preferred because they require
   no change to the encryption path used for the `GetRunner` response.
2. **Storage.** Add four columns to the `runner` table: `version`, `os`,
   `arch`, `started_at`. All nullable — older runners simply don't report
   them. The server overwrites these fields on every successful poll, next to
   the existing `touched` update.
3. **API.** Extend the `Runner` JSON shape returned by `/api/runners` and
   `/api/project/:id/runners` with the new fields. They are omitted when
   null so an older runner appears as "unknown" in the UI rather than as a
   zero value.
4. **UI.** Add three columns to the runners table in `web/src/views/Runners.vue`:
   - **Version** — plain text; a small warning chip when the runner's version
     differs from the server's own version (already available via
     `systemInfo.version`).
   - **Platform** — `os/arch` (e.g. `linux/amd64`, `darwin/arm64`).
   - **Uptime** — relative duration from `started_at` to now. Tooltip shows
     the absolute timestamp.

## Steps

### 1. Runner — collect and send metadata

- At startup, capture `runtime.GOOS`, `runtime.GOARCH`, the embedded build
  version, and `time.Now()` as the runner's `started_at`.
- Attach them as request headers on the existing poll request:
  - `X-Runner-Version`
  - `X-Runner-OS`
  - `X-Runner-Arch`
  - `X-Runner-Started-At` (RFC3339)
- No new endpoint, no extra round trip.

### 2. Server — accept, validate, persist

- In `RunnerController.GetRunner`, read the four headers. Treat missing
  headers as "unknown" (do not error — older runners must keep working).
- Validate: version ≤ 64 chars, os/arch from a small allowlist or
  `[a-z0-9_]+` regex, `started_at` parses as RFC3339 and is not in the
  future.
- Extend `db.RunnerManager.TouchRunner` (or add a sibling
  `TouchRunnerWithInfo`) to write the new columns in the same UPDATE that
  bumps `touched`. One write, no extra DB load.
- Add the migration adding `version`, `os`, `arch`, `started_at` columns to
  the `runner` table for all three SQL dialects (MySQL, Postgres, SQLite).
  Add equivalent fields to the Bolt model.

### 3. API — expose the new fields

- Extend the `db.Runner` struct with the four new fields (JSON-tagged, all
  `omitempty` / pointer-typed where appropriate).
- No new endpoint. The existing list/detail endpoints return them
  automatically once the struct is updated.
- Confirm Swagger / `api-docs.yml` reflects the additions.

### 4. Frontend — render the new columns

- In `web/src/views/Runners.vue` `getHeaders()`, add `version`, `platform`,
  `uptime` columns. Slot templates:
  - `item.version`: text + warning chip when `item.version !== version`
    (the existing `version` computed property already strips the build
    suffix from the server version).
  - `item.platform`: `${item.os}/${item.arch}` or `—` when missing.
  - `item.uptime`: relative duration (`dayjs(item.started_at).fromNow(true)`
    or equivalent helper already in the codebase). Tooltip with the absolute
    timestamp.
- Keep all three columns hidden-friendly: render `—` when the field is
  missing so mixed fleets (old + new runners) look clean.

### 5. Verification

- Start a runner against an updated server: confirm version, platform and
  uptime appear within one poll cycle.
- Restart the runner: confirm uptime resets, version/platform unchanged.
- Run an *old* runner (without the new headers) against the updated server:
  confirm the row still shows the runner as alive and the three new cells
  render `—`.
- Run a *new* runner against an old server: confirm the runner still polls
  successfully (server ignores unknown headers — verify in a staging build).
- Bump the server version locally and confirm the warning chip appears on
  runners running the previous version.
- Run with SQL and Bolt backends; confirm parity.

## Rollout

- One schema migration adding four nullable columns. No backfill needed.
- Backend and frontend ship together. Mismatched versions degrade cleanly:
  - Old frontend + new backend → extra JSON fields ignored.
  - New frontend + old backend → fields missing, cells render `—`.
  - Old runner + new server → headers absent, cells render `—`.
  - New runner + old server → headers ignored by the server's HTTP layer.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Runner forging headers to misrepresent version/platform | The runner is already authenticated via its token; this is informational, not a security boundary. Same trust level as `touched`. |
| `started_at` clock skew between runner and server | Stored as the value reported by the runner; rendered relatively. Acceptable: a few seconds of skew is invisible at "uptime" granularity. |
| Schema churn when we later add more fields (hostname, etc.) | New columns are independent and additive; each future field follows the same pattern without disturbing this one. |
| UI clutter on narrow screens | Existing table already has a horizontal scroll affordance; the three new columns are short. If needed, hide Platform behind a tooltip on Version. |
| Header size on hosts behind aggressive proxies | All four headers together are well under 256 bytes — no realistic risk. |

## Follow-ups (not part of this plan)

- **Resource usage** (CPU / RAM / disk free, cache size). Requires a
  sampling loop on the runner and a small time-series story on the server.
- **Hostname / container ID / IP** for "which box is this actually running
  on" — straightforward extension of the same header mechanism, deferred
  until there is a concrete operator ask.
- **Version-aware compatibility warnings** — server refuses or warns when a
  runner's version is below a declared minimum. The version field added
  here is the prerequisite.
- **Platform-aware task routing** — let templates require a specific
  OS/arch and let the scheduler honour it. The platform fields added here
  are the prerequisite.
