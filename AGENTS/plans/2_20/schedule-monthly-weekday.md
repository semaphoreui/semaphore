# Plan — Monthly Weekday Schedules ("Nth weekday + offset")

## Why

Commonly needed when patching Windows servers: updates ship on Patch Tuesday
(2nd Tuesday), so windows are often defined relative to it. Cron can't express
these:

- **Third Wednesday** — no nth-weekday operator; the DOM∧DOW trick fails because
  robfig/cron ORs a restricted DOM with a restricted DOW.
- **First Wednesday after Patch Tuesday** — not a fixed ordinal (2nd Tue + 1 day),
  so even a Quartz `#` can't name it.

## What

A Semaphore cron descriptor stored in the existing `cron_format`:

```
@monthly-weekday <ordinal> <weekday> [offset <days>] [at <HH:MM>]
```

`<ordinal>` = `1`..`5` or `last`; `<weekday>` = `sun`..`sat`; `offset` in
`[-28,28]` (calendar-day shift, may cross a month); `at` = 24h local, default
`00:00`. Example: `@monthly-weekday 2 tue offset 1 at 03:00` = first Wednesday
after Patch Tuesday.

## Design

- **Cron descriptor, not a new column.** A custom `cron.ScheduleParser` parses it
  and delegates every other expression to the standard parser; it backs both the
  pool (`cron.WithParser`) and `ValidateCronFormat`, so what validates is what
  fires. Reusing `cron_format` means no DB migration and export/backup keep
  working. HA-safe (pure function of the string). No new dependency.
- **Timezone** mirrors robfig/cron: ambient pool location, or a `CRON_TZ=`/`TZ=`
  prefix overrides.
- **Preview**: `schedules/validate` returns `next_run` for descriptors the JS
  cron parser can't evaluate.

## Files

- `services/schedules/monthly_weekday.go` — grammar, schedule, parser, `NextRunTimes`.
- `services/schedules/SchedulePool.go` — `WithParser`; `ValidateCronFormat`.
- `api/projects/schedules.go` — `next_run` in the validate response.
- `web/src/components/ScheduleForm.vue` — "Monthly (by weekday)" builder (same
  chip grids as Monthly); `api-docs.yml`.

## Edge cases

Ordinal 5 that doesn't occur → skip that month (scan bounded to 5 years). Offset
spill → scan starts one month back; occurrences are monotonic. DST → `time.Date`
resolves wall-clock correctly; `AddDate` preserves it. Validation is total.

## Tests

testify, table-driven, next to source: both headline cases across a year;
last/5th/offset/negative; DST; `CRON_TZ`; every validation error; a property
suite vs an independent oracle; a differential sweep proving the parser is a pure
superset of `ParseStandard`.
