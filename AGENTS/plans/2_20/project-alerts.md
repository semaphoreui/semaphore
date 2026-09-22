# Spec — Project alerts

Branch `feat/project-alerts`, PR #4272. Migration `2.20.6`.

## Goal

Let a project decide where each task reports, instead of every configured
messenger receiving every task of every project. Keep everything that worked
before working the same way.

## Terms

- **Server channel** — a messenger configured by the administrator in
  `config.json` (Telegram bot, Slack webhook, SMTP, ...). Shared by all projects.
- **Project alert** — a named destination that belongs to one project: a
  Telegram chat, a Slack webhook URL, a list of e-mail addresses, and so on.
- **Event** — a task moment that produces a message: `success`, `error`,
  `waiting_confirmation`.

## Behaviour

### Server channels (unchanged)

- Read from `config.json` at send time. Changing the config keeps working.
- A project receives them only when its "send to server channels" switch is on.
  This is the old *Allow alerts for this project* flag under a new name.
- Telegram may use a per-project chat ID, as before.
- Chat channels report success, error and waiting-for-confirmation. E-mail
  reports errors only. Both as before.

### Project alerts (new)

Each alert has a name, a channel type, the destination fields that channel
needs, a list of events, a "project default" flag, an enabled flag and an
optional custom message template.

- **Events** default to the channel defaults (chat: all three; e-mail: error)
  and can be changed per alert.
- **Project default** alerts are sent by every template that uses project
  defaults.
- Disabled alerts are never sent.
- Webhook URLs supplied by users must be http(s) and may not point at
  localhost, link-local, unspecified or cloud-metadata addresses. Server
  channel URLs are exempt.

### Secrets

Telegram, Gotify and e-mail need a secret (bot token, application token,
SMTP credentials). An alert never stores it. It chooses one of:

- **Use the server settings** — the secret from `config.json`. Offered only
  when the server has it configured.
- **Use my own** — an access key from the project Key Store, stored
  encrypted like every other key. Tokens use the new key type *Secret token*
  (`string`); SMTP credentials use *Login with password*. The key must
  belong to the project and be of the type the channel declares.

An e-mail alert with its own credentials may also set its own SMTP host,
port, sender and encryption; settings it leaves empty fall back to the
server ones. A key used by an alert can not be deleted; the Key Store shows
the alerts that use it.

### Where a template sends

- **Project defaults** (default for existing templates): server channels, when
  the project switch is on, plus project alerts marked as default.
- **Custom list**: only the selected alerts. Empty list means silent.

The existing *suppress success* / *suppress error* flags still apply to both.

### Where a schedule sends

- **Inherit** (default): whatever the template resolves to.
- **Custom list**: only the selected alerts, replacing the template choice.

### Delivery

- The routing decision is computed when the task is created and stored on the
  task. Later edits do not change what an already created task sends.
- Every delivery is recorded per task, destination and event. In HA, only the
  node that records it first sends the message.
- Sending runs in the background and never blocks the task.

### Testing

- "Test" on one alert sends a test message to it, ignoring its events.
- "Test all" sends through every server channel enabled for the project and
  every enabled project alert. Nothing to send returns 409.

### UI

- New **Alerts** page in the project sidebar: a server channels card (switch,
  Telegram chat ID, list of configured channels) and the project alert table
  with test, clone, edit and delete.
- **New Alert** is a drop-down listing the channels, like templates,
  inventories and schedules. The type is fixed once chosen; the form has no
  type selector and its fields are driven by the channel description from
  the API.
- Template form: project defaults / custom list plus the two suppress flags.
- Schedule form: inherit / custom list.
- Alert fields are removed from project settings.

### API

- `GET/POST /project/{id}/alerts`, `GET/PUT/DELETE /project/{id}/alerts/{alert_id}`
- `GET /project/{id}/alerts/channels` — supported channels, their fields,
  default events, default template, the secret they need (access key type)
  and whether the server-wide secret exists.
- `GET /project/{id}/alerts/{alert_id}/refs`, `POST /project/{id}/alerts/{alert_id}/test`
- Templates and schedules gain `alert_mode` and `alert_ids`.
- Backups gain an `alerts` section; templates and schedules reference alerts
  by name, alerts reference their access key by name.

## Extensibility

A messenger is one `Channel` implementation in `services/alerting` registered
in `NewRegistry`. It declares its type, title, icon, destination fields
(columns or JSON params, with a kind for the UI), the secret it needs and
its access key type, default events, body format, built-in template,
validation, how to read the server-wide destination from config, and how to
send. The API, the UI form, validation and delivery all follow that
description, so nothing else changes.

## Database schema

New tables:

| Table | Columns | Notes |
|---|---|---|
| `project__alert` | `id`, `project_id`, `name`, `type`, `enabled`, `is_default`, `events`, `chat_id`, `thread_id`, `url`, `recipients`, `key_id`, `params`, `body` | Unique `(project_id, name)`. `events` is a comma-separated list, empty means channel defaults. `key_id` → `access_key` (encrypted secret; null = server-wide secret). `params` is JSON with channel override settings (SMTP host, port, sender, secure, tls). Cascade on project delete. |
| `project__template_alert` | `project_id`, `template_id`, `alert_id` | PK `(template_id, alert_id)`. Cascade on template delete, restrict on alert delete. |
| `project__schedule_alert` | `project_id`, `schedule_id`, `alert_id` | PK `(schedule_id, alert_id)`. Cascade on schedule delete, restrict on alert delete. |
| `task__alert_send` | `task_id`, `destination`, `event`, `created` | PK `(task_id, destination, event)`. `destination` is `alert:<id>` or `instance:<type>`. Cascade on task delete. |

New columns:

| Table | Column | Values |
|---|---|---|
| `project__template` | `alert_mode` | `default` (default) or `ids` |
| `project__schedule` | `alert_mode` | `inherit` (default) or `ids` |
| `task` | `alert_snapshot` | JSON: `instance`, `alert_ids`, `on_success`, `on_error`; null for tasks created before the upgrade |

Unchanged and still used: `project.alert`, `project.alert_chat`,
`project__template.suppress_success_alerts`, `project__template.suppress_error_alerts`,
`access_key` (new keys of type `string` hold single tokens).

Rollback: `v2.20.6.err.sql` drops the columns and tables above.

## Not included

- Per-schedule success/error overrides. Events live on the alert and on the
  template flags; a schedule only picks destinations.
