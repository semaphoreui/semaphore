# TC-024 — Stop a running task and verify status `stopped`

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Tasks                            |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A user with appropriate permissions can stop a running task; the runner
terminates the underlying process and the task moves to terminal status
`stopped` (not `success` and not `error`).

## Preconditions

* Project `Infra QA`.
* Template `long-sleep` running a Bash script `sleep 300`.

## Steps

1. Launch `long-sleep`. Wait until the task is `running` and the log shows the
   process has started.
2. From the task detail page click **Stop**.
3. Confirm the prompt.
4. Watch the log and the status badge.
5. After the task is terminal, open **History** and verify the row.
6. Inspect runner logs / `ps` on the runner host to confirm the `sleep` PID
   no longer exists.
7. Repeat with **Stop** issued via API:
   `POST /api/project/{id}/tasks/{tid}/stop`.

## Expected results

* Step 4: status transitions `running → stopping → stopped` within ~5
  seconds; final exit/signal logged.
* Step 5: history shows status `stopped`.
* Step 6: process is gone on the runner; no orphan.
* Step 7: API returns 204 and behaves identically.
* If invoked on a task in `success`/`error`, API returns 4xx — stop is only
  valid for non-terminal tasks.

## Postconditions

None.
