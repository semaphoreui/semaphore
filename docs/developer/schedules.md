# Schedules

## Purpose

Schedules trigger template runs on a timer. Each schedule belongs to a project,
references a template, and can be **cron** (recurring) or **run_at** (one-shot).

The server evaluates cron expressions with `github.com/robfig/cron/v3`
`ParseStandard` — five fields (minute, hour, day-of-month, month, day-of-week),
no seconds field.

## Schedule types

| `type` value | Constant | Behaviour |
| --- | --- | --- |
| `""` (empty) | `ScheduleTypeCron` | Recurring; `cron_format` required |
| `run_at` | `ScheduleTypeRunAt` | One-shot; `run_at` must be in the future; `cron_format` cleared |

`validateSchedulePayload` (`api/projects/schedules.go`) enforces these rules on
create and update.

## Timezone

Cron evaluation uses `schedule.timezone` from server config
(`SEMAPHORE_SCHEDULE_TIMEZONE`, default `UTC`). The UI preview in
`ScheduleForm.vue` parses expressions with the same timezone via `cron-parser`.

## Cron validation API

The UI validates cron expressions against the server before save so users see the
same errors the API will return.

| Method | Path | Body |
| --- | --- | --- |
| `POST` | `/api/project/{project_id}/schedules/validate` | `{ "cron_format": "0 * * * *" }` |

Responses:

- `200` — expression is valid (empty body)
- `400` — `{ "error": "Cron: <parser message>" }`

`ScheduleForm.vue` and `TemplateForm.vue` (schedule-type templates) call this
endpoint on cron changes. Users can opt into **raw cron** mode in the UI
(`schedule__raw_cron` localStorage flag) to skip client-side preview while still
relying on server validation at save time.

Shared validator: `schedules.ValidateCronFormat` in
`services/schedules/SchedulePool.go`.

## REST API

Routes under `/api/project/{project_id}/schedules` (`api/router.go`):

| Method | Path | Action |
| --- | --- | --- |
| `GET` | `/schedules` | List project schedules |
| `POST` | `/schedules` | Create schedule |
| `POST` | `/schedules/validate` | Validate cron expression |
| `GET/PUT/DELETE` | `/schedules/{schedule_id}` | Read, update, delete |
| `GET` | `/templates/{template_id}/schedules` | Schedules for one template |

Create/update run `validateSchedulePayload` which delegates cron checks to the
same `ValidateCronFormat` helper used by the validate endpoint.

## Execution

`SchedulePool` (`services/schedules/SchedulePool.go`) loads active schedules,
computes next run times, and enqueues tasks through the task pool. Survey secret
variables on scheduled runs follow the same task-bound storage path as manual
launches (see [survey-and-variable-types.md](survey-and-variable-types.md)).

## Related code

- Model: `db/Schedule.go`
- API handlers: `api/projects/schedules.go`
- Pool and cron parsing: `services/schedules/SchedulePool.go`
- UI: `web/src/components/ScheduleForm.vue`, `web/src/components/TemplateForm.vue`
- Config: `util/config.go` (`ScheduleConfig.timezone`)
