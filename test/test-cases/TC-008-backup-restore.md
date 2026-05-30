# TC-008 — Backup and restore a project

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Projects / Data portability      |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Partial                          |

## Objective

Exporting a project produces a backup JSON that can be re-imported into the
same or another Semaphore instance, recreating templates, inventories,
repositories, variable groups, schedules and integrations (secret values are
not included).

## Preconditions

* Project `Infra QA` populated with at least:
  * 1 repository
  * 1 inventory
  * 1 variable group with one ENV var and one secret
  * 1 task template (Ansible)
  * 1 cron schedule
  * 1 integration with one matcher and one extractor

## Steps

1. Open **Settings → Backup** and click **Download project backup**.
2. Inspect the downloaded JSON for the expected entities.
3. Create a second empty project `Infra QA Restored`.
4. In `Infra QA Restored` use **Settings → Restore** to upload the JSON.
5. Walk each section of the restored project and compare with the source.
6. Run the previously-existing template inside the restored project.
7. Open a secret in the variable group on the restored side.

## Expected results

* Step 2: file is valid JSON; counts of templates/inventories/etc. match the UI.
* Step 5: all entities recreated with the same names and configuration.
* Step 6: the template runs (after providing missing secrets and SSH keys).
* Step 7: secret value placeholders are visible (e.g. blank/`***`) — secrets
  are intentionally not exported and must be re-entered.
* No 500-level errors during export or import.

## Postconditions

A working `Infra QA Restored` project exists. Delete after the test if not
needed.
