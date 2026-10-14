# tasks — registry

All tasks and plans live in the `workbench` MCP: brief, plan, stages, journal and status are
stored there and nowhere else. This folder holds only this file — the project's workspace binding.
There are no per-task or per-plan files or folders here.

## Workspace

The workspace that holds this project's tasks — the only one. Its code is given by the user during
setup and is also recorded in `agent-primary.md`; keep the two in sync.

| Code | Title |
| --- | --- |
| `WORKSPACE@16484a7f90` | Semaphore UI |

## Rules

- At the start of every session, before any other task call, bind the workspace:

  ```
  mcp__workbench__workspace_use(workspace_code="WORKSPACE@16484a7f90")
  ```

  then `mcp__workbench__tasks_list` for open work.
- Before starting work: check whether it was done or started before — `tasks_list` with
  `status="any"` and a `query`, then `task_get` / `notes_list` on a match.
- Get a skill with `skills_list` / `skill_get` when its guidance is needed.
- Run the work in the MCP: `task_create` → `task_status` → plan via `body_set` after reading the
  code → `stage_add` / `stage_close` with evidence → `note_add` for decisions, findings and facts.
- A plan is part of its task, never a separate document: the approach and the files go in the
  task body, the steps go in stages. Work that needs steps with their own evidence is an
  `extended` task.
- Check tasks in the MCP as well: `task_get` shows what is still open; clear open decisions before
  hand-over.
- Stop at `in_review`. `done` and `canceled` are the user's to set.
- Never create a task or plan file in the repo, and never copy a brief, plan or journal into one —
  a copy rots and diverges from the original silently. Refer to a task by its `TASK@` code.
