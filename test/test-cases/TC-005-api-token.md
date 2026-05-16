# TC-005 — Create and revoke a personal API token

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | API authentication               |
| Priority     | High                             |
| Type         | Functional / Security            |
| Automatable  | Yes (curl)                       |

## Objective

A user can issue a personal API token, use it to call the API on behalf of
their identity, and revoke it so further calls fail.

## Preconditions

* Logged in as `alice` (TC-004) or any active user.
* `curl` available on the workstation.

## Steps

1. Open **User Settings → API Tokens**.
2. Click **Create token**, name it `qa-token`, optional expiration in 1 day.
3. Copy the displayed token value (it is shown only once).
4. From a shell, call:
   ```
   curl -H "Authorization: Bearer <TOKEN>" https://<host>/api/user
   ```
5. Verify the JSON payload returns the same user.
6. Use the token to list projects:
   ```
   curl -H "Authorization: Bearer <TOKEN>" https://<host>/api/projects
   ```
7. Return to the UI and click **Revoke** next to `qa-token`.
8. Repeat the API call from step 4.
9. Wait 24h (or set a 1-minute expiration in a separate run) and check that an
   expired token also stops working without explicit revocation.

## Expected results

* Step 3: the token is shown exactly once; subsequent UI views display only
  the metadata and a *Revoke* action.
* Steps 4-6: HTTP 200 responses with the expected payload.
* Step 8: HTTP 401 with no user data.
* Step 9: HTTP 401 after the expiration timestamp; the token row is marked
  `expired`.
* All token creations/revocations recorded in the admin Activity log.

## Postconditions

No active tokens for `alice`.
