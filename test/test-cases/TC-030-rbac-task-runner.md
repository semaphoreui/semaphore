# TC-030 — Invited user with `Task Runner` role cannot edit project resources

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | RBAC                             |
| Priority     | Critical                         |
| Type         | Functional / Negative / Security |
| Automatable  | Yes                              |

## Objective

A user invited to a project with the `Task Runner` role can run existing
templates but cannot create, edit, or delete project resources; the API also
enforces these permissions.

## Preconditions

* Admin session with project `Infra QA`.
* Existing template `Deploy site (dev)` and inventory `qa-hosts`.
* User `alice` exists (TC-004).

## Steps

1. As admin go to **Team → Invite user**, choose `alice`, role `Task Runner`.
2. Log out and log in as `alice`.
3. Open project `Infra QA`. Verify which left-nav entries are visible.
4. Open `Deploy site (dev)` and click **Run**.
5. Try to click **Edit** on the template, **New Template**, **Edit** on an
   inventory, **Delete** on a variable group.
6. Using `alice`'s API token, call:
   * `POST /api/project/{id}/templates` with a valid body.
   * `PUT /api/project/{id}/inventory/{iid}`.
   * `DELETE /api/project/{id}/environment/{eid}`.
7. Log back in as admin, change `alice`'s role to `Manager` and repeat step 6.

## Expected results

* Step 3: only **Templates**, **Tasks**, **History** are visible (no Edit
  buttons); **Settings/Team/Keys** etc. hidden or read-only.
* Step 4: the task launches and `alice` is recorded as `user_id` on the task.
* Step 5: edit/new/delete actions are disabled (button hidden or grayed-out).
* Step 6: API returns 403 for each call; no DB mutation.
* Step 7: with `Manager` role the same calls succeed (201/200/204).
* Activity log captures the role change and the 403 attempts (or at least the
  successful actions).

## Postconditions

Either remove `alice` from project or revert role per suite cleanup.
