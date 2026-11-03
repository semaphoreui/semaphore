# Schedules that run a workflow

A schedule starts either a task template or a workflow: `project__schedule` carries a nullable
`template_id` and a nullable `workflow_id` (migration `2.20.17`, a table rebuild on SQLite with
foreign keys off because `task.schedule_id` refers to it). Exactly one is set — `Schedule.Validate`
and the store (`validateRunTarget`, shared with integrations) refuse anything else, and a workflow
schedule may not be a commit checker (`repository_id`) because the git working copy is keyed by
template id.

## What runs

- `ScheduleRunner.Run` takes the HA dedup lock first, then branches: `runTemplate` (as before)
  or `runWorkflow`, which loads the workflow through `db.WorkflowManager`, starts it through
  `pro_interfaces.WorkflowService.StartWorkflow` with the system actor
  (`audit.ComponentScheduler`) and `WorkflowRunParams{ScheduleID}`; the survey is filled from
  defaults only. Both are injected with `SchedulePool.SetWorkflowService` in `cli/cmd/root.go`.
- The open-source workflow store returns an empty template, which the runner reports as
  "Workflows are not available in this build" and skips.

## Gate and edges

- API gate: `projects.ScheduleController` refuses a new or changed workflow target without
  `Features.WorkflowSchedulesManagement` (403, no audit deny event); an update that keeps the same
  workflow passes so an expired license does not freeze the schedule. Backup restore refuses a
  backup with such a schedule the same way.
- Backup stores the target by name (`template` / `workflow`, one set); schedules are restored
  after workflows. The old-style exporter skips workflow schedules (it carries no workflows).
- `GetProjectSchedules` left-joins both targets and returns `tpl_name` / `wf_name` (empty, not
  null). `GetTemplateSchedules` keys on `template_id` only.
- UI: `ScheduleForm.vue` has the same target toggle as `IntegrationForm.vue` (Pro chip
  `workflow_schedules_management`); `Schedule.vue` shows the target column.

## ⚠️ Pro side (2026-11-03)

`pro_impl` is not in this checkout. Two things must land there for the feature to work in Pro:
the feature table must set `WorkflowSchedulesManagement` for the plans that get it, and the real
`WorkflowService.StartWorkflow` must stamp `WorkflowRunParams.ScheduleID` on the root task
(`task.schedule_id`), as it does for `IntegrationID`. Until then Pro answers 403 on a workflow
schedule, and a run started by a schedule has no `schedule_id` on its root task.
