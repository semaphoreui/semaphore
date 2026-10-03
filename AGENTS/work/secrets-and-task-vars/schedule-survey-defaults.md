# Survey default values for scheduled and API-created tasks (issue #2244) — open problem

Survey variable defaults are applied only by the Vue forms (`TaskForm.vue`,
`TaskParamsForm.vue`), never on the backend. A schedule without task params, a task created
through the API or an integration without survey values, or a schedule whose template defaults
changed after it was saved, runs with those variables absent — a playbook with `vars_prompt`
fails, and re-running it from the UI succeeds because the form fills the defaults in.

Status (2026-09-27): not implemented — `TaskRunner.populateTaskEnvironment`
(`services/tasks/TaskRunner.go:445`) still returns early when the task environment is empty, and
no `GetSurveyVarsDefaults` exists on any branch (`git log --all -S`); issue #2244 is open.
Not implemented; to be created as a workbench task when picked up.

## Agreed approach

- Apply defaults on the backend at run time, in `TaskRunner.populateTaskEnvironment`, not in
  `SchedulePool` — one change covers schedules, integrations and direct API calls and keeps one
  merge implementation.
- Precedence, lowest to highest: template environment JSON, survey defaults, task environment.
  This is exactly what the forms do (`{...defaults, ...edited}`), so an explicit value wins.
- Drop the early return on an empty task environment; `getEnvironmentExtraVars` already
  tolerates `"{}"`.
- A helper on `db.Template` returning name → default for vars that declare one, unit-tested in
  `db`; table-driven cases on `populateTaskEnvironment` (scheduled/no env, overridden, coexisting
  with template env, no vars, existing behaviour preserved).
- Optional follow-up: pre-fill `task.Environment` in `ScheduleRunner.Run` so task history shows
  the applied values; harmless together with the run-time merge.

## Constraints and what changed since the plan

- ⚠️ 2026-09-27: `SurveyVar.DefaultValue` is no longer a string — it is
  `*SurveyVarDefaultValue` (`Values []string`, array-vs-scalar aware, `db/Template.go`), and
  multi-select defaults are arrays. The helper must emit an array for `SurveyVarSelect` and a
  scalar otherwise, matching what the forms send; the plan's "strings, no coercion" rule needs
  re-deciding for `int`/`enum`.
- ⚠️ 2026-09-27: `services/tasks/LocalJob.go` named in the plan is now
  `services/tasks/local_executor.go`.
- Schedules created in the UI already snapshot defaults into `task_params`
  (`TaskParamsForm.vue`, since 9a0c08278, 2025-04-16); the backend merge must still let that
  explicit value win, and the snapshot explains why the bug shows only for schedules made
  without params or via the API.
- The task-API variable gate (default-deny, `decisions.md`) lives on branch
  `feat/survey_vars_validation`; defaults must be applied server-side after validation, never
  injected into the client payload.
- Required vars with no default remain unsatisfiable from a schedule — out of scope; a warning
  on the schedule form is a possible follow-up. Per-schedule survey editing (discussion #2226)
  already works through task params.

## Key files and references

- `services/tasks/TaskRunner.go` (`populateTaskEnvironment`, `populateDetails`),
  `services/schedules/SchedulePool.go` (`ScheduleRunner.Run`), `db/TaskParams.go`
  (`CreateTask`), `db/Template.go` (`SurveyVar`, `SurveyVarDefaultValue`),
  `web/src/components/TaskForm.vue`, `web/src/components/TaskParamsForm.vue`.
- Issue https://github.com/semaphoreui/semaphore/issues/2244; discussion
  https://github.com/semaphoreui/semaphore/discussions/2226.
- Source plan: `AGENTS/plans/2_20/2244-survey-default-values-in-schedules.md`.
