# TC-009 — Delete project only after dependents are removed

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Projects                         |
| Priority     | High                             |
| Type         | Negative + Positive              |
| Automatable  | Yes                              |

## Objective

A project can be deleted, but only after running tasks have finished and the
operator confirms data loss; deletion cascades to project-scoped entities.

## Preconditions

* Project `Infra QA Restored` (from TC-008) or a dedicated scratch project
  populated with at least one template and one inventory.
* No external user invites pending.

## Steps

1. Launch a long-running task in the project (e.g. `sleep 60` template).
2. While the task is `running`, open **Settings → Delete project**.
3. Stop the running task (TC-024 procedure).
4. Retry **Delete project**; type the confirmation string when prompted.
5. After deletion verify:
   * The project disappears from the project switcher.
   * `GET /api/projects` does not return it.
   * `GET /api/project/{id}/templates` returns 404/403.
6. Check the database (admin SQL access) and confirm related rows
   (`project__template`, `project__inventory`, `project__repository`,
   `project__environment`, `project__schedule`, `project__integration`,
   `project__user`) are gone for that project_id.

## Expected results

* Step 2: deletion is blocked or warns about active tasks.
* Step 4: confirmation prompt requires typing the project name (or similar
  acknowledgement) before the destructive action executes.
* Step 5: all cascading checks confirm the project and its dependents are
  removed.
* Audit / Activity log retains an entry indicating who deleted the project.

## Postconditions

Scratch project is gone.
