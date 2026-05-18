# TC-027 — Register a remote runner via registration token

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Runners                          |
| Priority     | High                             |
| Type         | Functional / Integration         |
| Automatable  | Partial                          |

## Objective

A remote `semaphore runner` process registers itself against the server with a
valid registration token, becomes visible in **Runners**, and picks up tasks.

## Preconditions

* Semaphore server reachable over network from a second host.
* `runner_registration_token` configured on the server (or generated via the
  admin UI).
* Semaphore binary present on the runner host.

## Steps

1. On the server side, copy the registration token from
   **Admin → Runners → Registration token**.
2. On the runner host run:
   ```
   semaphore runner --config /etc/semaphore/runner.json
   ```
   with `registration_token` and `server_url` set in the config.
3. Refresh **Admin → Runners** in the UI.
4. Stop a project's local execution (set the project to use the new runner) or
   tag the runner and route a tagged template to it (see TC-028).
5. Run any template and watch which runner picks it up (visible on the task
   page).
6. Kill the runner process.
7. Re-run the template.

## Expected results

* Step 3: the new runner appears `Active=true`, `Touched` updates within
  10 seconds; a public key is recorded server-side and the one-time
  registration credential is consumed. The long-lived runner token is stored
  where the runner is configured to persist it: e.g. `runner.token` in
  `config.json` when no `token_file` is set, or the path in `token_file` /
  `SEMAPHORE_RUNNER_TOKEN_FILE` (the official Docker image defaults this to
  `runner_token.txt` under `SEMAPHORE_DATA_PATH` when using `--no-config`).
* Step 5: task is assigned to the new runner; `runner_id` on the task matches.
* Step 6-7: task stays in `waiting` because no eligible runner is online; once
  the runner restarts, the task is picked up.
* If an invalid registration token is used the registration endpoint returns
  401/403 with no runner row created.

## Postconditions

Runner left online for TC-028.
