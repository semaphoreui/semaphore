# TC-004 — Admin creates and deactivates a user

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | User management                  |
| Priority     | High                             |
| Type         | Functional / Positive            |
| Automatable  | Yes                              |

## Objective

An administrator can create a new local user, the user can sign in, and the
administrator can deactivate the account so the user can no longer authenticate.

## Preconditions

* Logged in as admin (TC-001).
* SMTP not required — local auth path is being verified.

## Test data

| Field    | Value                |
|----------|----------------------|
| Username | `alice`              |
| Email    | `alice@example.com`  |
| Name     | `Alice QA`           |
| Password | `S3curePass!`        |
| Admin    | No                   |

## Steps

1. Navigate to **Settings → Users → New User**.
2. Fill in the form with the test data and submit.
3. Verify the user appears in the list with `Admin = false`.
4. Log out and log in as `alice` / `S3curePass!`.
5. Log back in as admin and edit `alice` → toggle **Deactivate** (or remove).
6. Log out and attempt to log in as `alice` again.
7. Restore by recreating or reactivating `alice` for downstream tests if needed.

## Expected results

* Step 2: response is 201/200; user list refreshes and `alice` is shown.
* Step 4: `alice` lands on the empty projects screen and the admin-only
  navigation entries are hidden.
* Step 6: login is rejected with the standard *Invalid credentials* error.
* All actions are recorded in **Activity** (admin-only event log).

## Postconditions

`alice` either deleted or reactivated depending on suite needs.
