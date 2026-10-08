# schedules — zone index

Cron and one-time schedules: their syntax and day offsets, how they are evaluated, in which
timezone, and what a scheduled task receives.

| When you need it | Read |
| --- | --- |
| Picking up per-schedule timezones or the `CRON_TZ=` prefix, or touching the global `schedule.timezone` | `per-schedule-timezone.md` |
| Changing the schedule form's next-run preview or `run_at` handling | `per-schedule-timezone.md` |
| Adding cron syntax, or wondering why both robfig/cron and cronexpr parse cron formats | `quartz-days-and-offsets.md` |
| Changing day offsets, the "Monthly by weekday" timing or the server-side next-run preview | `quartz-days-and-offsets.md` |
| Survey defaults in scheduled tasks | `../secrets-and-task-vars/INDEX.md` |
