# TC-015 — Static inventory with two host groups

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Inventory                        |
| Priority     | Critical                         |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A static inventory with multiple groups is parsed correctly by Ansible launched
via Semaphore, and group-targeted plays only affect the selected group.

## Preconditions

* Project `Infra QA`.
* Key `qa-host-ssh` (TC-012) usable for both hosts.
* Two reachable test hosts:
  * `web01` 10.0.0.31
  * `db01` 10.0.0.21

## Test data — Inventory body

```
[web]
web01 ansible_host=10.0.0.31 ansible_user=qa

[db]
db01 ansible_host=10.0.0.21 ansible_user=qa
```

## Steps

1. Create static inventory `multi-group` with the body above and SSH key
   `qa-host-ssh`.
2. Create a template `ping-web` running `ansible web -m ping`.
3. Create a template `ping-db` running `ansible db -m ping`.
4. Run both templates and inspect the logs.
5. Intentionally introduce a syntax error in the inventory (missing closing
   bracket) and save.

## Expected results

* Step 4: `ping-web` only pings `web01`, `ping-db` only pings `db01`.
* Step 5: save is rejected client-side or server-side with a parse error
  ("inventory parsing failed"), and the previous good version is preserved.
* Inventory edit via API (`PUT /api/project/{id}/inventory/{iid}`) returns 4xx
  with the validation error.

## Postconditions

Inventory restored to the valid body.
