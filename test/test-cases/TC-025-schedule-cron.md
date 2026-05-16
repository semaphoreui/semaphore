# TC-025 — Cron schedule triggers the template at the expected time

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Schedules                        |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A cron-formatted schedule on a template launches a task at the expected wall
clock minute in the configured timezone, and the **Active** toggle gates
execution.

## Preconditions

* Server timezone correctly set (default UTC) — or
  `SEMAPHORE_SCHEDULE_TIMEZONE` set to a known value.
* Template `health-check` (lightweight Bash echo).

## Test data

| Field          | Value                  |
|----------------|------------------------|
| Schedule name  | `every-2-min`          |
| Cron format    | `*/2 * * * *`          |
| Active         | true                   |

## Steps

1. Create the schedule from the template's **Schedules** tab.
2. Wait at least 4 minutes; observe `health-check` history.
3. Toggle the schedule **Active = false**.
4. Wait 4 minutes; observe history.
5. Re-enable the schedule.
6. Save an invalid cron expression (`*/foo`) and observe validation.

## Expected results

* Step 2: two tasks fired at minute :00 and :02 (or :02 and :04), each with
  `schedule_id` populated in the task payload and `user_id` null.
* Step 4: no new task fires while inactive.
* Step 5: tasks resume.
* Step 6: save rejected with a parse error from the cron library; previous
  good schedule preserved.
* Server time vs schedule next-run UI matches the configured timezone (verify
  by setting `SEMAPHORE_SCHEDULE_TIMEZONE=America/Los_Angeles` and confirming
  the next-run time shifts).

## Postconditions

Disable or delete the chatty schedule before suite continues.
