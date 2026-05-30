# TC-016 — File-based Ansible inventory from the playbook repo

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Inventory                        |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

When inventory type is `file`, Ansible reads the inventory from the path inside
the repository (or a linked repository) at task launch time.

## Preconditions

* Project `Infra QA`.
* Repository `playbooks` (TC-010) contains `inventories/dev/hosts.ini`.
* Key `qa-host-ssh` available.

## Test data

| Field        | Value                          |
|--------------|--------------------------------|
| Inventory    | `dev-from-repo`                |
| Type         | `file`                         |
| Inventory    | `inventories/dev/hosts.ini`    |
| SSH key      | `qa-host-ssh`                  |
| Repository   | (default — template's repo)    |

## Steps

1. Create inventory `dev-from-repo` per the test data.
2. Run an existing Ansible template using `dev-from-repo`.
3. Push a commit to the repo that removes `inventories/dev/hosts.ini`.
4. Re-run the template.
5. Restore the file in the repo and re-run.

## Expected results

* Step 2: task succeeds; the inventory file is read from the cloned repo and
  hosts are processed.
* Step 4: task fails with a clear "inventory file not found" error; no fallback
  to the prior version.
* Step 5: task succeeds again.
* The inventory `Inventory` field stored in the DB is the relative path string
  (no leading slash).

## Postconditions

Repo restored.
