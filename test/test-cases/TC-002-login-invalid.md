# TC-002 — Login fails with wrong password and does not leak account existence

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Authentication / Security        |
| Priority     | Critical                         |
| Type         | Negative / Security              |
| Automatable  | Yes                              |

## Objective

Confirm that the login endpoint rejects invalid credentials, returns a generic
error message that does not disclose whether the user exists, and does not
issue a session cookie.

## Preconditions

* Admin user `admin` exists.
* No active session in the browser.

## Test data

| Case | Username        | Password   |
|------|-----------------|------------|
| A    | `admin`         | `wrong!`   |
| B    | `does-not-exist`| `whatever` |
| C    | `admin`         | `` (empty) |

## Steps

1. Open the login page.
2. For each case A, B, C: submit the form and observe the response and UI.
3. Inspect the network response body and status code for each case.
4. After 5 consecutive failures with valid username + wrong password, attempt
   case A again to check throttling/lockout behavior (if implemented).
5. Re-submit a correct password to confirm the account is not permanently
   locked.

## Expected results

* All three cases display the same generic error (e.g. *"Invalid credentials"*)
  — the UI must not distinguish "user not found" from "wrong password".
* HTTP status is `401` (or `400` for the empty password case), and no
  `Set-Cookie: semaphore=…` header is returned.
* No session cookie is created in the browser.
* Repeated failures either rate-limit or log a warning to the server log; if
  rate-limited, the response is `429` or includes a `Retry-After` header.
* A subsequent valid login (TC-001) succeeds — the account is not locked.

## Postconditions

No persistent session. Server log contains warn-level entries for each failed
attempt.
