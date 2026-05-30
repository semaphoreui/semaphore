# TC-028 — Tag-scoped task routing to matching runner

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Runners                          |
| Priority     | Medium                           |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A task whose template/inventory specifies a runner tag executes on a runner
with the matching tag and never on a runner without it.

## Preconditions

* Two runners online:
  * `runner-a` with tag `linux`.
  * `runner-b` with tag `windows`.
* Templates `linux-hello` (tag `linux`) and `windows-hello` (tag `windows`).

## Steps

1. Configure the tags via **Admin → Runners → Edit** for each runner.
2. On `linux-hello` set inventory's `runner_tag` to `linux`. Same for
   `windows-hello` with `windows`.
3. Run `linux-hello` 3 times back-to-back.
4. Run `windows-hello` 3 times back-to-back.
5. Take `runner-a` offline (`Active=false` in UI), run `linux-hello` again.
6. Set a template's runner tag filter to a non-existent tag and try to run.

## Expected results

* Step 3: all three tasks are assigned to `runner-a` only — verify by
  `runner_id` on each task.
* Step 4: all three assigned to `runner-b`.
* Step 5: task stays `waiting`; bringing `runner-a` back online unblocks the
  queue.
* Step 6: task stays `waiting` indefinitely with no eligible runner;
  the UI surfaces a warning.

## Postconditions

Re-enable both runners.
