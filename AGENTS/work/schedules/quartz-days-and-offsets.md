# Quartz day forms and day offsets

Cron schedules accept the Quartz day forms `2#2` (second Tuesday), `5L` (last Friday), `L` (last
day), `LW` (last weekday) and `15W` (weekday nearest the 15th), and `offset_days` moves every run by
whole calendar days. Together they cover patch windows such as the day after Patch Tuesday
(`0 3 * * 2#2`, offset 1) or the Monday after the last Friday (`0 22 * * 5L`, offset 3), which
plain cron cannot express because it ORs day of month and day of week.

## Decisions and why

- **robfig/cron keeps parsing everything it accepts**, with its own DST handling, so existing
  schedules run as before. It has been unmaintained since 2020 and lacks `#`, `L` and `W`, but
  replacing it changes existing schedules: on 20 000 random robfig-valid expressions,
  netresearch/go-cron rejected 44 % (stricter step rules such as `*/60`) and hashicorp/cronexpr
  computed different runs for 3 % (`*/1` and `0-0` in a day field).
- **hashicorp/cronexpr evaluates only the Quartz day forms**, and only through `parseExpression`:
  - Day fields must be one of the forms above (weekday lists such as `1#1,1#3` are fine), with the
    other day field `*` or `?`. cronexpr misreads others (`02#2` is a Sunday) and panics on
    reversed ranges.
  - robfig reads the minute, hour and month fields and cronexpr gets plain lists, so those fields
    mean what they mean in any other schedule.
  - cronexpr goes backwards where clocks change at midnight (Africa/Cairo, Asia/Beirut: the same
    run every second, or an endless loop with an offset), so its runs are computed on the wall
    clock in UTC and then placed in the schedule's zone (`cronSchedule.Next`).
- **Wall-clock rules** apply to Quartz forms and to every schedule with an offset: a run in the hour
  skipped by DST happens after the change (02:30 becomes 03:30) and a run in the repeated hour
  happens once. An offset keeps the time of day; it is limited to ±31 days (`MaxOffsetDays`,
  mirrored in `cronPresets.js`) and refused for `@every`, which has no time of day.
- **`TZ=Local`** means the schedule zone, as it did in robfig/cron, not the process zone.
- **The server computes the preview.** `POST /schedules/validate` returns `next_runs` from the same
  `ParseCronSchedule` the scheduler uses. cron-parser cannot read `LW`, `15W` or lists of `#`
  items, so it only fills the builder.
- **HA**: runs derive from the DB row only; the `TryLockExecution` dedup path is unchanged.

## Schedule form

- Timing "Monthly by weekday": Week of the month (1st–4th, Last), Weekdays, Offset, Hours,
  Minutes. Both day rows are multi-select like the other timings, and the day of week lists every
  combination (`1#1,1#3`). The 5th is left to raw cron, as in most calendar apps.
- An expression the timings cannot show — another Quartz form, or an offset on a format that is
  not "monthly by weekday" — keeps raw mode, which always shows the Offset field.
- `beforeLoadData` drops the previous schedule's preview, because the dialog reuses the form.

## Follow-ups

- User guide: `docs/docs/user-guide/schedules.md` in the docs submodule does not describe the
  Quartz day forms or `offset_days` yet; a change there needs all 11 locales.

## Key files

- `services/schedules/cron_schedule.go` (`ParseCronSchedule`, `NextRuns`), `SchedulePool.go`
- `api/projects/schedules.go` (`ValidateScheduleCronFormat`), `api-docs.yml`
- `db/Schedule.go` (`OffsetDays`), migration `v2.20.11`
- `web/src/components/ScheduleForm.vue`, `ScheduleOffsetField.vue`, `web/src/lib/cronPresets.js`
