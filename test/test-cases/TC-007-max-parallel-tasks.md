# TC-007 — Update `max_parallel_tasks` and observe queueing

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Projects / Scheduling            |
| Priority     | Medium                           |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

Setting `max_parallel_tasks = N` on a project caps the number of concurrently
running tasks at N; further tasks remain in `waiting` until a slot frees.

## Preconditions

* Project `Infra QA` exists (TC-006).
* A template `sleep-30` exists that runs a Bash script `sleep 30` and is
  configured with **Allow parallel tasks** enabled.
* At least one runner is online and not constrained by tags.

## Test data

* Project `max_parallel_tasks` initial value: `2`.

## Steps

1. Open **Settings → Edit** on the project and set `max_parallel_tasks = 2`.
2. From the `sleep-30` template launch 4 tasks back-to-back from the UI (or
   loop `POST /api/project/{id}/templates/{tid}/tasks` 4 times via curl).
3. Open the **Tasks** dashboard and observe the status column.
4. Wait for the first two to finish and observe the third/fourth.
5. Change `max_parallel_tasks` to `4`, repeat step 2 with 4 fresh launches.

## Expected results

* Step 3: exactly 2 tasks are in `running`, the other 2 are in `waiting`.
* Step 4: as each running task completes, a waiting task transitions to
  `running`; final state of all four is `success`.
* Step 5: all 4 tasks transition to `running` immediately (no `waiting`).
* No task is silently dropped; the total count after each batch matches the
  number of launches.

## Postconditions

`max_parallel_tasks` left at the value required by downstream tests.
