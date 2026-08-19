# TC-013 — Login/Password key as become credential

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Key Store / Inventory            |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A `Login With Password` key set as the inventory's *become key* is used by
Ansible for sudo password prompts; the password is not exposed in logs.

## Preconditions

* Project `Infra QA`.
* Host `db01` (10.0.0.21) configured for password-based sudo for user `qa`.
* Inventory `qa-hosts` from TC-012.

## Test data

| Field      | Value           |
|------------|-----------------|
| Key name   | `qa-become`     |
| Login      | (empty)         |
| Password   | `qa-sudo-pass`  |

## Steps

1. Create key `qa-become` of type `Login With Password` storing the sudo
   password.
2. Edit inventory `qa-hosts` and set **Become key** to `qa-become`.
3. Create or reuse an Ansible template that runs a privileged play, e.g.
   ```yaml
   - hosts: linux
     become: true
     tasks:
       - name: read shadow
         command: head -n1 /etc/shadow
   ```
4. Run the template.
5. Inspect the task output and the request payload sent to the runner.

## Expected results

* Step 4: task succeeds; the play has root output.
* Step 5: the password value is never shown in the task log; only a redacted
  marker appears (e.g. `BECOME-SUCCESS-…`).
* No `--ask-become-pass` prompt is emitted; the password is supplied to
  ansible via the runtime variable injection mechanism.

## Postconditions

Inventory still bound to `qa-become`.
