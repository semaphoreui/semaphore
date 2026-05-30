# TC-026 — One-shot `run_at` schedule with delete-after-run

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Schedules                        |
| Priority     | Medium                           |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A schedule of type `run_at` fires exactly once at the configured timestamp and
is removed when **Delete after run** is on.

## Preconditions

* Template `health-check` (TC-025).
* Server clock NTP-synced.

## Test data

| Field            | Value                                  |
|------------------|----------------------------------------|
| Type             | `run_at`                               |
| Name             | `one-shot`                             |
| Run at           | now + 2 minutes (ISO 8601 UTC)         |
| Delete after run | true                                   |

## Steps

1. Create the schedule with the test data.
2. Wait ~3 minutes.
3. Inspect history for `health-check`.
4. Reload the **Schedules** tab.
5. Repeat the case with `Delete after run = false` and a future timestamp;
   verify it stays but is marked inactive after firing.

## Expected results

* Step 3: exactly one task fired within 60 seconds of the configured time.
* Step 4: the `one-shot` schedule is gone from the list.
* `GET /api/project/{id}/schedules` returns 200 and the schedule is not in
  the body.
* Step 5: the schedule remains, but `Active` flips to false after firing — it
  does not fire again.
* Saving a `run_at` timestamp in the past returns a validation error.

## Postconditions

No leftover scheduled tasks.
