# TC-020 — Create an Ansible task template and run it

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Task Templates                   |
| Priority     | Critical                         |
| Type         | Functional / Smoke               |
| Automatable  | Yes                              |

## Objective

An admin can create a `task` template for the Ansible app, configure a
playbook, inventory and variable group, run it from the UI, and view a
successful task with full log streaming.

## Preconditions

* Project `Infra QA`.
* Repository `playbooks` containing `site.yml`.
* Inventory `qa-hosts`.
* Variable group `dev-vars`.

## Test data

| Field          | Value                  |
|----------------|------------------------|
| Template name  | `Deploy site (dev)`    |
| Type           | `task`                 |
| App            | `ansible`              |
| Playbook       | `site.yml`             |
| Inventory      | `qa-hosts`             |
| Environment    | `dev-vars`             |
| Repository     | `playbooks`            |
| Allow override | none                   |

## Steps

1. **Task Templates → New Template** with the data above.
2. Click **Run** on the new template.
3. Observe the task page: live log, status transitions
   (`waiting` → `running` → `success`), commit hash and commit message
   populated.
4. After completion, click **Re-run** from the task detail page.
5. Open the **History** view and verify both tasks are listed with author and
   duration.

## Expected results

* Step 2: a task is created and a runner picks it up within a few seconds.
* Step 3: the log streams progressively (websocket / SSE); on success the
  status badge is green, exit code = 0, and the commit hash matches `git
  rev-parse HEAD` in the repo.
* Step 4: the new task uses the latest HEAD of the same branch.
* Step 5: both runs appear, ordered newest first, with correct durations.

## Postconditions

The template is ready for downstream tests (overrides, schedules,
integrations).
