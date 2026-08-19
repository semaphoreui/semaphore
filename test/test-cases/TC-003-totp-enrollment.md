# TC-003 — Enable TOTP and log in with one-time code

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Authentication / 2FA             |
| Priority     | High                             |
| Type         | Functional / Positive            |
| Automatable  | Partial (TOTP code generation)   |

## Objective

A user can enroll a TOTP authenticator, store the recovery code, and
authenticate using the second factor on subsequent logins.

## Preconditions

* Standard non-admin user `qa-user` exists with a known password.
* TOTP feature is enabled on the server (default).
* TOTP application available on a test device (e.g. `oathtool` or Google
  Authenticator).

## Test data

* Username: `qa-user`
* Password: `Passw0rd!`

## Steps

1. Log in as `qa-user`.
2. Navigate to **User Settings → Security → Enable Two-Factor**.
3. Scan the QR code (or copy the secret) into the authenticator.
4. Enter the current 6-digit TOTP code and click **Activate**.
5. Note the displayed recovery code; store it.
6. Log out.
7. Log in again with `qa-user` and password.
8. When prompted, enter the current TOTP code.
9. Attempt a final fresh login but use the recovery code instead of TOTP.
10. After step 9, generate a new recovery code and verify the previous one is
    rejected.

## Expected results

* Step 4: activation succeeds, the UI confirms 2FA is enabled, and a recovery
  code is shown exactly once.
* Step 7-8: a TOTP challenge page appears after the password prompt; a valid
  6-digit code completes the login, an invalid one returns an error and keeps
  the user on the challenge page.
* Step 9: a recovery code is accepted exactly once.
* Step 10: the previously consumed recovery code is rejected.
* `GET /api/user` shows `totp` populated.

## Postconditions

The user has 2FA enabled. Disable it via Security settings before subsequent
non-2FA test cases.
