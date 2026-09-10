# Implementation Plan — Per-Schedule Timezones

## Goal

Today all cron schedules are evaluated in a single global timezone
(`schedule.timezone` config / `SEMAPHORE_SCHEDULE_TIMEZONE` env, default `UTC`).
This feature adds **per-schedule timezone support**:

- New `timezone` column/field on `project__schedule` / `db.Schedule`.
- Respect the robfig/cron timezone prefix in the cron expression itself:
  `CRON_TZ=Asia/Hong_Kong 22 20 * * 1-5`.
- UI: timezone selector in the schedule form, correct "Next run time" preview,
  timezone shown in the schedule list.

## Timezone Precedence (design decision)

For a cron schedule, the effective timezone is resolved in this order:

1. `CRON_TZ=` / `TZ=` prefix inside `cron_format` (highest priority — explicit in
   the expression).
2. `schedule.timezone` column (new field), if non-empty.
3. Global `util.Config.Schedule.Timezone` (existing behavior, default `UTC`).

Empty `timezone` column ⇒ behavior identical to today, so **existing rows need no
data migration**.

For `run_at` schedules the stored value is an absolute instant (UTC) — the
timezone field only affects how the UI *interprets/displays* the date-time input,
not the backend firing logic.

## Current State (reference paths)

- `db/Schedule.go` — `Schedule` struct (no timezone field).
- `db/Store.go:664` — `ScheduleProps`.
- `db/sql/schedule.go` — `CreateSchedule` / `UpdateSchedule` use **explicit column
  lists** (must be extended); all reads use `select *` (no change needed).
- `services/schedules/SchedulePool.go`:
  - `init()` — `cron.New(cron.WithLocation(loc))` with the single global location.
  - `Refresh()` — `addRunner(runner, schedule.CronFormat)` →
    `p.cron.AddJob(cronFormat, runner)`.
  - `ValidateCronFormat()` — `cron.ParseStandard(cronFormat)`.
- **robfig/cron v3 already parses the `CRON_TZ=`/`TZ=` prefix** in both
  `ParseStandard` and `AddJob`. So validation and firing of prefixed expressions
  largely work already — the work is in the new column, precedence wiring, and UI.
- `api/projects/schedules.go` — `validateCronFormat`, `validateSchedulePayload`,
  `ValidateScheduleCronFormat` (backend of `POST /schedules/validate`).
- `api/router.go:337` — validate route.
- `api/system_info.go` — exposes global `schedule_timezone` to the UI.
- `web/src/components/ScheduleForm.vue` — receives global `timezone` as a prop;
  uses JS `cron-parser` (v5) for next-run preview + checkbox extraction; validation
  already delegated to the backend endpoint.
- `web/src/views/project/Schedule.vue` — passes
  `systemInfo.schedule_timezone` prop, formats `run_at` in the global tz.
- `api-docs.yml` — `Schedule` / `ScheduleRequest` definitions.
- `db/sql/migrations/` — latest is `v2.20.0.sql`; registry in
  `db/Migration.go` (`GetMigrations`, last entry `{Version: "2.20.0"}`).

**Note:** JS `cron-parser` does NOT understand the `CRON_TZ=` prefix — the UI must
strip it before parsing and use it as the `tz` option.

---

## Implementation Steps

### Phase 1 — DB model & migration

**1.1 Migration `db/sql/migrations/v2.20.1.sql`** (verify next free version at
implementation time; register in `db/Migration.go` after `2.20.0`):

```sql
alter table `project__schedule` add `timezone` varchar(64) not null default '';
```

- Backtick/dialect translation is handled by the migration runner — confirm the
  statement works on MySQL, PostgreSQL and SQLite (pattern-match existing
  `alter table ... add` migrations). Add a `.err.sql` companion only if a dialect
  needs one (see `v2.20.0.err.sql` for the convention).

**1.2 `db/Schedule.go`** — add field:

```go
Timezone string `db:"timezone" json:"timezone"`
```

- No `backup:"-"` tag ⇒ the field is included in project backup/restore
  automatically (same as `CronFormat`). Verify with a backup round-trip test.

**1.3 `db/sql/schedule.go`** — add `timezone` to the explicit column lists in
`CreateSchedule` (insert) and `UpdateSchedule` (update). Reads use `select *`,
no changes.

### Phase 2 — Scheduling engine

**2.1 Effective-spec helper** in `services/schedules` (pure function, easy to
test):

```go
// BuildCronSpec returns the spec to register in the cron pool,
// applying the timezone precedence rules.
func BuildCronSpec(cronFormat string, timezone string) string {
    if hasTZPrefix(cronFormat) || timezone == "" {
        return cronFormat // explicit CRON_TZ wins; empty tz falls back to pool location
    }
    return "CRON_TZ=" + timezone + " " + cronFormat
}
```

`hasTZPrefix` detects `CRON_TZ=`/`TZ=` at the start of the spec (same rule as
robfig/cron). Prepending `CRON_TZ=` lets robfig/cron do all location handling —
no changes to the pool's `cron.New(cron.WithLocation(...))` (which stays as the
tier-3 fallback for schedules without a timezone).

**2.2 `SchedulePool.Refresh()`** — for `ScheduleTypeCron`, call
`p.addRunner(runner, BuildCronSpec(schedule.CronFormat, schedule.Timezone))`.

`run_at` branch: no change (absolute time).

### Phase 3 — Validation & API

**3.1 `services/schedules`** — add:

```go
func ValidateTimezone(timezone string) error {
    if timezone == "" {
        return nil // = use server default
    }
    _, err := time.LoadLocation(timezone)
    return err
}
```

`ValidateCronFormat` stays as-is — `cron.ParseStandard` already accepts the
`CRON_TZ=` prefix (add a test proving it, including an invalid-zone prefix like
`CRON_TZ=Bad/Zone * * * * *` which must fail).

**3.2 `api/projects/schedules.go`**:

- `validateSchedulePayload` — for cron-type schedules also validate
  `schedule.Timezone` (400 with a clear `error` message on invalid zone).
- `ValidateScheduleCronFormat` (the `POST /schedules/validate` handler) — also
  validate the `timezone` field from the payload.
- **Recommended enhancement:** on success, return the computed next run time
  (`schedule.Next(time.Now())` from the parsed spec with precedence applied):

  ```json
  { "next_run": "2026-07-22T20:22:00+08:00" }
  ```

  This lets the UI show a backend-authoritative "Next run time" even for specs
  JS `cron-parser` can't fully handle, and removes tz-precedence duplication in
  the frontend. UI can keep the local computation as an instant preview and
  reconcile when the response arrives.

**3.3 `api-docs.yml`** — add `timezone` (string, e.g. `Asia/Hong_Kong`, empty =
server default) to `Schedule` and `ScheduleRequest`; document the `CRON_TZ=`
prefix in the `cron_format` description. Extend the validate endpoint's response
schema if 3.2's enhancement is implemented.

### Phase 4 — Frontend

**4.1 `ScheduleForm.vue`**:

- New timezone `v-autocomplete` (cron type only), items from
  `Intl.supportedValuesOf('timeZone')` plus a "Server default (<tz>)" empty
  option. New items default to `''` (server default). Bound to `item.timezone`.
- Effective-tz computed property implementing the same precedence:
  `CRON_TZ prefix → item.timezone → this.timezone prop`. Use it for:
  - `nextRunTime()` — strip the `CRON_TZ=`/`TZ=` prefix before
    `CronExpressionParser.parse(...)` and pass the effective tz via the `tz`
    option (JS cron-parser does not understand the prefix).
  - `refreshCheckboxes()` field extraction — parse the stripped expression; a
    prefixed expression stays fully supported in raw mode.
  - run_at parsing/formatting (`dayjs.tz(..., effectiveTz)`) in
    `setDefaultRunAt`, `setRunAtInputFromItem`, `nextRunTime`, `beforeSave`.
- When the cron expression carries a `CRON_TZ=` prefix, disable the timezone
  selector and show a hint ("timezone is set in the cron expression").
- The "Next run time" table's first row currently labels the global tz — label
  it with the effective tz instead. If 3.2's `next_run` response is implemented,
  prefer the backend value.
- The info alert about `SEMAPHORE_SCHEDULE_TIMEZONE` should mention it is only
  the *default*; per-schedule timezone overrides it.

**4.2 `Schedule.vue` (list view)**:

- Show the schedule's timezone: append to the cron format column (e.g.
  `22 20 * * 1-5 (Asia/Hong_Kong)`) when `item.timezone` is set or the format
  has a `CRON_TZ=` prefix.
- `formatRunAt` — unchanged (absolute instant shown in global tz), or display in
  the schedule's tz when set; pick one and be consistent.

**4.3 i18n** — add new strings (`Timezone`, `Server default`, hint text) to
`web/src/lang/*` (at minimum `en.js`, mirror keys elsewhere).

### Phase 5 — Tests

Per `.claude/CLAUDE.md` (testify; `_test.go` next to source; table-driven):

- `services/schedules/SchedulePool_test.go`:
  - `BuildCronSpec` — table-driven: no tz, tz set, existing `CRON_TZ=` prefix,
    existing `TZ=` prefix (prefix always wins).
  - `ValidateCronFormat` — plain spec, `CRON_TZ=Asia/Hong_Kong 22 20 * * 1-5`
    (valid), invalid zone in prefix, garbage spec.
  - `ValidateTimezone` — empty, valid IANA name, invalid name.
  - Behavioral check that a `CRON_TZ`-prefixed spec produces the expected next
    activation instant (parse via `cron.ParseStandard`, compare `Next()` against
    a fixed reference time in UTC).
- Store test: create/update a schedule with `Timezone` set, read back, assert
  round-trip (extend existing schedule store tests if present).
- `api/projects/schedules_test.go` — `httptest` on the validate endpoint:
  valid payloads, invalid timezone → 400, `CRON_TZ`-prefixed format → 200.
- Backup/restore round-trip includes `timezone`.

---

## Edge Cases & Notes

- **Existing schedules** — `timezone = ''` ⇒ global config tz; zero behavior
  change on upgrade.
- **Prefix + field conflict** — the prefix wins by design; validation should NOT
  reject the combination (the UI prevents it, the API tolerates it).
- **DST** — robfig/cron handles DST per its documented semantics once the
  location is right; nothing extra to do, but worth a sentence in user docs.
- **HA mode** — per-schedule tz comes from the DB, so all nodes agree even if
  their `schedule.timezone` configs differ (improvement over today). The
  dedup path (`TryLockExecution`) is unaffected.
- **`run_at`** — stored as absolute time; tz affects only UI interpretation.
- **Commit-checker schedules** (`repository_id` set) — same cron path, works via
  `BuildCronSpec` automatically.
- **Docs** — update the schedules page in `docs/` (timezone field, `CRON_TZ=`
  prefix, precedence rules) and the `SEMAPHORE_SCHEDULE_TIMEZONE` description
  ("default timezone for schedules without their own").

## Suggested PR Breakdown

1. PR 1 — migration + `db.Schedule.Timezone` + store column lists +
   `BuildCronSpec`/`ValidateTimezone` + `SchedulePool` wiring + tests.
2. PR 2 — API validation (+ optional `next_run` in validate response) +
   `api-docs.yml`.
3. PR 3 — UI (form selector, effective-tz preview, CRON_TZ prefix handling,
   list view, i18n) + docs.

## Open Questions

1. Should the validate endpoint return `next_run` (Phase 3.2 enhancement)?
   Recommended — it makes the UI preview authoritative and CRON_TZ-proof.
2. Timezone list source in UI: `Intl.supportedValuesOf('timeZone')` is
   browser-provided and may differ slightly from Go's tzdata. Acceptable, since
   the backend validates with `time.LoadLocation` anyway?
3. Schedule list display format for timezone (separate column vs suffix on the
   cron expression).
