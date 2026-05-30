# TC-023 — Override branch, limit and tags at task launch

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Task Templates / Tasks           |
| Priority     | Medium                           |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

When the Ansible-specific overrides are enabled
(`allow_override_inventory`, `allow_override_limit`, `allow_override_tags`),
the launch dialog exposes inputs for them and the values flow to the
ansible-playbook invocation.

## Preconditions

* Template `Deploy site (dev)` from TC-020.
* Repository `playbooks` has both `main` and `feature/new-task` branches; the
  feature branch adds a tagged task.

## Steps

1. Edit template, in **Ansible parameters** toggle on:
   * `allow_override_inventory`
   * `allow_override_limit` and add default `[]`
   * `allow_override_tags` and add default `[]`
2. Save the template.
3. Click **Run** and:
   * Pick `feature/new-task` from the **Git branch** input.
   * Choose inventory `multi-group` (TC-015) for this run.
   * Enter `web` in the **Limit** field.
   * Enter `feature-only` in **Tags**.
4. Submit. Observe the task log for the resolved command line.
5. Re-run via **Re-run** and verify the overrides are remembered for that task
   but the template defaults are unchanged.

## Expected results

* Step 4: the runner-side log shows the invocation containing
  `--limit web --tags feature-only`, the repo cloned at the override branch,
  and the inventory file matching `multi-group`.
* The task detail page shows the override values under **Arguments**.
* Step 5: re-run preserves the per-task overrides; opening the template form
  still shows the defaults from step 1.
* Toggling the override flags back off hides the corresponding launch dialog
  fields.

## Postconditions

Template overrides remain on for downstream tests; revert in cleanup if
needed.
