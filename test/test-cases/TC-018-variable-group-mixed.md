# TC-018 — Variable group: JSON extra-vars + ENV var + secret

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Variable Groups (Environment)    |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A variable group exposes three categories to a task: JSON extra-vars (as
ansible `--extra-vars`), regular ENV vars, and secrets (type `var` or `env`).
All three reach the task process correctly; secrets remain unprintable in logs.

## Preconditions

* Project `Infra QA`.
* Inventory `qa-hosts` (TC-012).
* Template uses Ansible app.

## Test data

| Field         | Value                                  |
|---------------|----------------------------------------|
| Group name    | `dev-vars`                             |
| JSON          | `{"region": "eu-west-1"}`              |
| ENV           | `LOG_LEVEL=debug`                      |
| Secret (var)  | `db_password = s3cret!`                |
| Secret (env)  | `API_TOKEN = abcd1234`                 |

## Steps

1. Create variable group `dev-vars` with the four entries above; use the
   **Secrets** tab for the two secrets, selecting type `var` and `env`
   respectively.
2. Attach `dev-vars` to a template that runs a debug play:
   ```yaml
   - hosts: linux
     tasks:
       - debug: msg="region={{ region }} log_level={{ lookup('env','LOG_LEVEL') }} api_token={{ lookup('env','API_TOKEN') }} db_password={{ db_password }}"
   ```
3. Run the task.
4. Reopen the variable group and observe the secret values in the UI.

## Expected results

* Step 3: the debug message shows `region=eu-west-1`, `log_level=debug`,
  `api_token=abcd1234`, and the playbook treats `db_password` as a variable.
* In the **task log** values originating from `Secret` entries are masked
  (display as `***` / hidden) — at minimum the secret-tab values must not
  appear in clear text in the rendered HTML page.
* Step 4: secret values are not pre-filled in the form; the field shows
  *(unchanged)* placeholder.

## Postconditions

`dev-vars` remains attached.
