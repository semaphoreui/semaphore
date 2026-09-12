# Plan — Show the Runner Name on the Task Details Page

## Goal

Make it visible, on the task details page, **which runner has picked up the
task**. Today the UI shows the template, commit, trigger, timing and parameters,
but nothing about *where* the task is actually executing. In setups with more
than one runner this is a recurring blind spot.

## Scope

In scope:

- Display the runner's **name** on the task details page for tasks that are
  currently assigned to a runner (typically queued or running).
- Surface this through the existing task-list API response so the front-end
  doesn't need an extra request.

Out of scope:

- Showing runner information for **finished** tasks. The system today clears a
  task's runner assignment when the task completes, so historical tasks have no
  runner to display. Preserving that link would require separate work and is
  tracked as a follow-up below.
- Exposing the runner's internal ID or token in the API.
- Any change to how tasks are routed to runners.

## Why a Name (not an ID)

A numeric ID would be technically simpler but unhelpful to the operator looking
at the page. The runner *name* is what users see everywhere else (runner list,
tags, settings), so reusing it keeps the UI consistent and immediately
actionable — no second lookup required.

## Design Summary

1. **Backend** enriches the task record returned for the task list / task
   details with the runner's name. The internal runner identifier remains
   server-side only.
2. **Front-end** adds a single new row to the *Running info* card on the task
   details page, shown only when a runner name is present. When no runner is
   assigned, the page looks exactly as it does today.

## Steps

### 1. Backend — include runner name in task responses

- Extend the "task with template" record (used by the task list and task
  details endpoints) with an optional `runner_name` field.
- Populate it by looking up the runner that the task is currently assigned to.
  If the runner has been removed, or the task is not assigned to any runner,
  the field is simply omitted.
- Apply this consistently to both supported storage backends (SQL and BoltDB)
  so the behaviour is identical regardless of deployment.

### 2. Front-end — show the runner in *Running info*

- Add a **Runner** row to the *Running info* card on the task details page,
  underneath the existing duration row.
- Render the row only when the new `runner_name` field is present in the task
  payload, so existing tasks without a runner assignment are unchanged.
- No new request, no new component — this reuses the same task object the page
  already loads.

### 3. Verification

- Open a freshly started task and confirm the Runner row appears with the
  expected name.
- Open a finished task and confirm the Runner row is absent (current behaviour
  preserved).
- Run with both SQL and Bolt backends and confirm parity.
- Delete a runner that owns an in-flight task and confirm the page degrades
  cleanly (row hidden) rather than showing an empty value or error.

## Rollout

- No database migration is required.
- Backend and frontend ship together. An older frontend talking to the updated
  backend simply ignores the extra field; an updated frontend talking to an
  older backend never receives the field and hides the row.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| Runner name leaking sensitive info | Runner names are already visible to project members in the runner list; no new exposure. |
| Performance impact of the extra lookup on the task list | Done as a left join / cached lookup per page of tasks — negligible cost. |
| Front-end clutter | Row is hidden when there's no runner, so the existing layout is preserved for the common "already finished" case. |

## Follow-ups (not part of this plan)

- **Show runner for finished tasks.** Decide whether to preserve the runner
  assignment after a task ends — either by keeping `runner_id` populated, or by
  storing the runner name on the task row at assignment time. This is a
  separate decision because it affects data model and historical records.
- **App-aware task parameters block** (Terraform/Tofu/Terragrunt parameters,
  Ansible tags/skip tags) — orthogonal improvement to the same page.
- **Build / source task link** for deploy tasks that reference a build task.
