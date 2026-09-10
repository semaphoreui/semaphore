# Plan — SIEM-Ready Audit Events

> Status: **implemented** (16.07.2026, branch `feature/siem-audit-events` in the main
> repo and in `pro_impl`; docs — submodule, branch `feature/siem-audit-events`).
> Deviations from the plan:
> - `integration_id` did not exist in any dialect (it was not merely missing from the
>   INSERT) — the column was added in v2.20.1 for all dialects.
> - `log_writer` in the request context became optional (otherwise handlers cannot
>   be tested in isolation).
> - CR/LF sanitization is done in `helpers.EventLog` (single point), not in the webhook.
> - Graceful shutdown of the webhook: `Close()` exists but is not wired into the server
>   lifecycle (at most the queued events are lost on shutdown).
> - Frontend (showing action/IP in Activity) — not done, optional part of Task 6.

- **Branch:** `develop`
- **Research:** MCP research "SIEM support in Semaphore UI" (RESEARCH@a39c0ef615f486972eaff4),
  16.07.2026: analysis of the codebase, GitHub issues (#158, discussion #2194, #3410),
  competitors (AWX External Log Aggregator, Rundeck Audit Stream Plugin) and
  the requirements of OWASP Logging Vocabulary / SIEM ingestion (Splunk HEC, syslog RFC 5424, CEF).

## 1. Problem

Semaphore has an event log (the `event` table, the Activity page, the `/events`
API), but it is not usable as an audit trail for SIEM/compliance:

1. **Login/logout and failed login attempts are not logged at all** — neither in
   the `event` table nor in the file log. In `api/login.go` there is only a single
   logrus info line on successful LDAP authentication. This is requirement #1
   of OWASP and of any SOC.
2. **IP address and user-agent are not stored** in either `db.Event` or
   `pro_interfaces.EventLogRecord` — even though they are already extracted
   from the request when a session is created (`api/login.go:172-173`: `X-Real-IP`,
   `user-agent`).
3. **The `action` field (create/update/delete) does not reach the DB** — only
   the logrus fields and the file log (`api/helpers/event_log.go:48-62`). In the DB
   an event is distinguishable only by the `Description` text.
4. **Global CRUD of users and API tokens is not logged** —
   `api/users.go` (AddUser/UpdateUser/UpdateUserPassword/DeleteUser) and
   `api/user.go:129-171` (create/deleteAPIToken) do not call
   `helpers.EventLog`. Only membership/roles inside a project are logged.
5. **`integration_id` is lost**: it exists in the model but is omitted from the INSERT
   (`db/sql/event.go:36`).
6. **No push channel for audit**: syslog (`cli/cmd/syslog.go`) forwards the whole
   logrus log mixed with debug output; there is no generic outbound HTTP webhook
   (Splunk HEC etc.); alerts (`services/tasks/alert.go`) cover only
   task status.

## 2. Current State (context for the implementer)

- `db/Event.go:10-24` — the `Event` model; `db/sql/event.go:32-51` — INSERT.
- `api/helpers/event_log.go:29` — `EventLog(r *http.Request, action EventLogType, item EventLogItem)`:
  the single write point; writes to the DB and to `pro_interfaces.LogWriteService.WriteEventLog`.
- `pro_interfaces/log_write_svc.go:5-17` — the `LogWriteService` interface and
  `EventLogRecord`. The OSS implementation is a stub (`pro/services/server/log_write_svc.go`),
  the working one (JSON/raw file + lumberjack) is in `pro_impl/services/server/log_write_svc.go`.
- `util/config.go:290-311` — `EventLogType`/`TaskLogType` (env
  `SEMAPHORE_EVENT_LOG_*`); `util/config.go:320-326` — `SyslogConfig`.
- Migrations: `db/sql/migrations/v2.20.0.sql`, registered in
  `db/Migration.go:GetMigrations`.

## 3. Design

Principle (de-facto standard per AWX/Rundeck + OWASP): **a single vocabulary of
audit events → a single emission point → multiple emitters**.

### 3.1 Event vocabulary

Actions are extended from create/update/delete to:

```go
// api/helpers/event_log.go
const (
    EventLogCreate EventLogType = "create"
    EventLogUpdate EventLogType = "update"
    EventLogDelete EventLogType = "delete"

    EventLogLoginSuccess EventLogType = "login_success"
    EventLogLoginFail    EventLogType = "login_fail"
    EventLogLogout       EventLogType = "logout"
)
```

New object type: `EventSession EventObjectType = "session"` (in
`db/Event.go` next to the other constants).

### 3.2 Event enrichment

`db.Event` gets `Action`, `IP`, `UserAgent`; everything is filled in
`helpers.EventLog` from `*http.Request` — calling code does not change.

### 3.3 OSS / Pro split (product decision)

We follow the Rundeck model and the current repository split:

- **OSS**: enriched events in the DB (action/IP/UA), auth events,
  user/token CRUD. This covers trust/basics.
- **Pro** (`pro_impl/`): the file JSON log already exists; we add a
  **generic HTTP audit webhook** (compatible with Splunk HEC) —
  asynchronous, with retries; a delivery failure does not affect the request.
- Syslog RFC 5424 already exists in OSS for the whole log; we do not build a
  dedicated audit channel in syslog in this version (events already reach the logrus
  stream via `event.ToFields()` — that is sufficient; a dedicated channel/CEF is
  a candidate for the next version if there is demand).

## 4. Tasks

### Task 1 — Migration and model: action, ip, user_agent in event

**Files:**
- Create: `db/sql/migrations/v2.20.1.sql`
- Modify: `db/Migration.go` (add `{Version: "2.20.1"}` to the list),
  `db/Event.go`, `db/sql/event.go` (there is no bolt storage in the repo — SQL only)

**Migration** (`v2.20.1.sql`):

```sql
alter table event add `action` varchar(20) null;
alter table event add `ip` varchar(45) null;
alter table event add `user_agent` varchar(255) null;
```

**Model** (`db/Event.go`, add to `Event`):

```go
Action    *string `db:"action" json:"action"`
IP        *string `db:"ip" json:"ip"`
UserAgent *string `db:"user_agent" json:"user_agent"`
```

`db/sql/event.go:CreateEvent` — extend the INSERT with the new columns and
fix the loss of `integration_id` at the same time:

```go
_, err = d.exec(
    "insert into event(user_id, project_id, integration_id, object_id, object_type, description, created, action, ip, user_agent) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    evt.UserID, evt.ProjectID, evt.IntegrationID, evt.ObjectID, evt.ObjectType,
    evt.Description, created, evt.Action, evt.IP, evt.UserAgent)
```

**Test:** `db/sql/event_test.go` — CreateEvent stores and returns
action/ip/user_agent and integration_id (following the existing sql tests;
testify, `require.NoError`).

### Task 2 — helpers.EventLog: filling the new fields from the request

**Files:**
- Modify: `api/helpers/event_log.go`, `pro_interfaces/log_write_svc.go`,
  `pro/services/server/log_write_svc.go` (stub — signatures do not change),
  `pro_impl/services/server/log_write_svc.go` (write the new fields to JSON/raw)

In `EventLog` before writing:

```go
func extractClientIP(r *http.Request) string {
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return host
}
```

```go
actionStr := string(action)
ip := extractClientIP(r)
ua := r.Header.Get("user-agent")
event.Action = &actionStr
event.IP = &ip
event.UserAgent = &ua
```

Extend `EventLogRecord` (`pro_interfaces/log_write_svc.go`):

```go
IP        string `json:"ip,omitempty"`
UserAgent string `json:"user_agent,omitempty"`
ObjectType *string `json:"object_type,omitempty"`
ObjectID   *int    `json:"object_id,omitempty"`
```

(Currently not even object_type/object_id reach the file log — add them.)

**Test:** `api/helpers/event_log_test.go` — httptest request with `X-Real-IP`
and `User-Agent`; mock store; verify that `CreateEvent` received
populated Action/IP/UserAgent.

### Task 3 — Auth events: login/logout/fail

**Files:**
- Modify: `api/login.go`, `db/Event.go` (the `EventSession` constant),
  `api/helpers/event_log.go` (the login_success/login_fail/logout constants)

Insertion points:

1. `login(w, r)` (`api/login.go:292`) — after successful authentication
   (before/after `createSession`):
   ```go
   helpers.EventLog(r, helpers.EventLogLoginSuccess, helpers.EventLogItem{
       UserID:      user.ID,
       ObjectType:  db.EventSession,
       ObjectID:    user.ID,
       Description: fmt.Sprintf("User %s logged in", user.Username),
   })
   ```
   On every failure branch (wrong password, user not found, LDAP fail) —
   `EventLogLoginFail` with the login from the request in Description. **Never log
   the password.** For "user not found" UserID stays 0 (not written).
2. `logout` (`api/login.go:455`) — `EventLogLogout`.
3. OIDC callback (the handler after `oidcLogin`/redirect, where
   `createSession` is called for the OIDC user) — the same login_success/login_fail.
4. Failed TOTP verification (the verify handler in `api/auth.go` /
   `api/login.go`, look for `SessionVerificationTotp`) — `EventLogLoginFail`
   with Description "MFA verification failed".

Note: instance-level events (ProjectID == 0) are not visible to non-admins in
`/events` — this is correct (`api/events.go:21-23`).

**Test:** `api/login_test.go` — use the existing login test helpers;
verify that after a wrong password an event with
action=login_fail and an IP appeared in the store.

### Task 4 — Events for global CRUD of users and API tokens

**Files:**
- Modify: `api/users.go` (AddUser, UpdateUser, UpdateUserPassword,
  DeleteUser, DeleteUserIdentity), `api/user.go` (createAPIToken,
  deleteAPIToken)
- Modify: `db/Event.go` — new type `EventAPIToken EventObjectType = "api_token"`

Example (AddUser, after successful creation):

```go
helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
    UserID:      editor.ID, // who performed the action
    ObjectType:  db.EventUser,
    ObjectID:    newUser.ID,
    Description: fmt.Sprintf("User %s created", newUser.Username),
})
```

- UpdateUserPassword → Description "Password changed for user %s"
  (never the password itself).
- createAPIToken/deleteAPIToken → ObjectType `api_token`, Description without
  the token value, only the ID.

**Test:** extend `api/users_test.go` (or create it) — AddUser writes a
create event with ObjectType user.

### Task 5 (Pro) — Audit webhook emitter (generic HTTP / Splunk HEC)

**Files (in `pro_impl/`, a separate repo!):**
- Create: `pro_impl/services/server/audit_webhook_svc.go` + `_test.go`
- Modify: `pro_impl/services/server/log_write_svc.go` — after writing to
  the file, hand the record to the webhook emitter
- Modify (OSS): `util/config.go` — config:

```go
// util/config.go, in ConfigLog:
type AuditWebhookConfig struct {
    Enabled  bool              `json:"enabled" env:"SEMAPHORE_AUDIT_WEBHOOK_ENABLED"`
    URL      string            `json:"url" env:"SEMAPHORE_AUDIT_WEBHOOK_URL"`
    // Authorization headers, for example:
    //   Authorization: "Splunk <hec-token>"  → Splunk HEC
    //   Authorization: "Bearer <token>"      → generic
    Headers  map[string]string `json:"headers" env:"SEMAPHORE_AUDIT_WEBHOOK_HEADERS"`
    // splunk_hec | json (default json)
    Format   string            `json:"format" env:"SEMAPHORE_AUDIT_WEBHOOK_FORMAT"`
}
```

Emitter requirements (from research, the AWX/Rundeck standard):

- Asynchronous: a buffered channel (capacity ~1000) + a single sender goroutine;
  `EventLog` never blocks and never returns a delivery error.
- Retries: 3 attempts with backoff (1s/5s/30s); after that the record is dropped with
  a logrus warn. `// ponytail: in-memory buffer; a disk queue — if requested`.
- `splunk_hec` format: envelope `{"time": <unix>, "event": {...}, "sourcetype": "semaphore:audit"}`
  POSTed to `<url>/services/collector/event`.
- `json` format: POST the record as-is.
- Sanitization: CR/LF in Description are replaced with a space.
- Graceful shutdown: flush the buffer on server stop (with a 5s timeout).

**Test:** httptest.Server as the receiver; verify the HEC envelope, retry on
500, and no blocking when the receiver is unavailable.

### Task 6 — Documentation and config schema

**Files:**
- Modify: `config.schema.yaml` — new fields `log.audit_webhook.*`
  (use the `semaphore-config-schema` skill)
- Modify: `api-docs.yml` / swagger description of Event (new fields action/ip/user_agent)
- Modify: docs `admin-guide` — a page about SIEM integration: a table
  "channel → format → how to connect Splunk/Elastic/Wazuh"
- Frontend: `web/src/views/project/Activity.vue` (or wherever Activity is
  rendered) — show action and IP in the event list (optional, can be a
  separate task)

## 5. Out of Scope (deliberately, candidates for 2.21+)

- CEF/LEEF formatter and a dedicated audit channel in syslog — only on demand
  (QRadar customers).
- Before/after (diff) in object change events.
- Disk queue for the webhook (currently an in-memory buffer).
- Retention/archiving of the `event` table.
- Secret read-access events (secret access trail) — on the Enterprise
  roadmap (related to #3410), separate plan.

## 6. Order and dependencies

Task 1 → Task 2 → (Task 3, Task 4 — independent, in parallel) → Task 5
(depends on Task 2 for the EventLogRecord fields) → Task 6.

Each task is a separate commit(s) `feat(audit): ...`, tests per the
rules in `.claude/CLAUDE.md` (testify, table-driven where appropriate).
