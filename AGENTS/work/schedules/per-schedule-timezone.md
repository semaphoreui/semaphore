# Per-schedule timezone

Cron schedules are evaluated in one global timezone (`schedule.timezone` /
`SEMAPHORE_SCHEDULE_TIMEZONE`, default `UTC`). The feature adds a `timezone` field per schedule,
honours the robfig/cron `CRON_TZ=`/`TZ=` prefix inside the expression, and shows the effective
timezone in the form preview and the list. This is an open problem: designed, not built.

Status (2026-09-27): not implemented — `db.Schedule` has no `Timezone` field, no migration
touches `project__schedule`, `services/schedules` has no `BuildCronSpec`/`ValidateTimezone`,
`ScheduleForm.vue` has no timezone selector and no `CRON_TZ` handling, `api-docs.yml` has no
`timezone` on `Schedule`. `git log --all -S Timezone -- db/Schedule.go` finds only the plan
commit (3678a3c08). Branch `origin/fix_schedule_tz` only adds "UTC" labels to the form and is
already merged into develop.

Not implemented; to be created as a workbench task when picked up.

## Motivation

Users in different regions share one instance and want "20:22 Hong Kong time" without
re-deriving UTC. The global setting is a config value, so in HA each node could in principle be
configured differently; a per-schedule value in the DB makes all nodes agree.

## Agreed approach

- **Precedence:** `CRON_TZ=`/`TZ=` prefix in `cron_format` → `schedule.timezone` column →
  global `util.Config.Schedule.Timezone`. Empty column means today's behaviour, so existing
  rows need no data migration and upgrade is a no-op.
- **Let robfig/cron do the location handling.** Its `ParseStandard` and `AddJob` already parse
  the prefix, so validation and firing of prefixed expressions work today. The engine change is
  a pure helper that prepends `CRON_TZ=<tz> ` when the column is set and no prefix exists, used
  in `SchedulePool.Refresh()`; `cron.New(cron.WithLocation(loc))` stays as the tier-3 fallback.
- **Prefix + column conflict is tolerated by the API** (prefix wins); the UI prevents it by
  disabling the selector and showing a hint when the expression carries a prefix.
- **Validation:** `time.LoadLocation` on the column (empty passes); cron-type schedules validated
  in `validateSchedulePayload` and in the `POST /schedules/validate` handler, 400 on a bad zone.
- **`run_at` schedules are absolute instants** — timezone affects only how the UI interprets and
  displays the input, never the firing logic.
- **Storage:** `alter table project__schedule add timezone varchar(64) not null default ''`;
  `CreateSchedule`/`UpdateSchedule` use explicit column lists and must be extended; reads use
  `select *`. No `backup:"-"` tag so backup/restore carries it.
- **UI:** `v-autocomplete` from `Intl.supportedValuesOf('timeZone')` plus a "Server default"
  empty option; an effective-timezone computed property drives next-run preview, checkbox
  extraction and `run_at` parsing. JS `cron-parser` (v5, `CronExpressionParser`) does **not**
  understand `CRON_TZ=`: strip the prefix and pass it as the `tz` option. List view appends the
  timezone to the cron column.
- HA: timezone comes from the DB row, so nodes agree regardless of config; the
  `TryLockExecution` dedup path is untouched. DST semantics are robfig/cron's own.

## Constraints

- Next free migration on this branch is past `v2.20.7` (`db/Migration.go`); verify at
  implementation time and add an `.err.sql` undo companion (convention since `v2.20.0`).
- Docs: `docs/docs/user-guide/schedules.md` describes the global timezone only; it must gain the
  field, the prefix and the precedence rules, and `SEMAPHORE_SCHEDULE_TIMEZONE` must be
  re-described as the *default* for schedules without their own.

## Open questions

- Should `POST /schedules/validate` return the computed `next_run` so the UI preview is
  backend-authoritative and `CRON_TZ`-proof? Recommended; the UI would keep the local value as
  an instant preview and reconcile.
- Browser `Intl` zone list vs Go tzdata may differ slightly; acceptable since the backend
  validates with `time.LoadLocation`.
- List display: separate column or suffix on the cron expression.
- Whether `formatRunAt` in the list should show the schedule's zone or stay global.

## Key files and references

- `db/Schedule.go`, `db/sql/schedule.go`, `db/Migration.go`, `db/sql/migrations/`
- `services/schedules/SchedulePool.go` (`Refresh`, `ValidateCronFormat`, `WithLocation`),
  `services/schedules/SchedulePool_test.go`
- `api/projects/schedules.go` (`ValidateScheduleCronFormat`), `api/system_info.go`
  (`schedule_timezone`), `api-docs.yml`
- `web/src/components/ScheduleForm.vue`, `web/src/views/project/Schedule.vue`
- `util/config.go` (`Schedule.Timezone`), `docs/docs/user-guide/schedules.md`
- Source plan: `AGENTS/plans/2_20/schedule-timezone.md`
