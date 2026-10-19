# Reusable survey variables (issue #2212) — open problem

Survey variables are defined inline per template, as JSON in `project__template.survey_vars`.
Issue #2212 asks for a separate, project-scoped entity — a named collection of survey variables —
that can be assigned to many templates the way an Environment (Variable Group) is, so one change
replaces editing every template by hand.

Status (2026-09-27): not implemented — no `project__survey` table, `db/Survey.go`,
`SurveyManager` or `survey_ids` on any branch (`git log --all -S project__survey` is empty);
issue #2212 is open. Not implemented; to be created as a workbench task when picked up.

## Agreed approach

- Mirror the Environment pattern 1:1 — it already solved "separate, multi-assignable entity":
  entity table `project__survey` (id, project_id, name, vars JSON) plus M:N junction
  `project__template_survey`, a `SurveyManager` interface shaped like `EnvironmentManager`,
  `GetTemplateSurveys`/`UpdateTemplateSurveys` on `TemplateManager`, a
  `/api/project/{project_id}/surveys` subrouter gated by `manageProjectResources`, a list view
  and form, a multi-select next to the environment picker in `TemplateForm.vue`, backup/restore
  via `backup:` tags, i18n.
- Reuse the existing `SurveyVar` / `SurveyVarEnumValue` / `SurveyVarType` types — no duplicates.
- Inline `Template.SurveyVars` stays supported; the effective list at task time is the merge of
  inline vars and every assigned survey's vars. No forced data migration; `survey_ids` is
  additive and optional everywhere. The runner payload keeps receiving a flat resolved list.
- Extract the variable editor out of `SurveyVars.vue` into a shared child so the inline editor
  and the new entity form are the same UI.
- Delivery in five PRs: model + migration + store; REST API; merge logic; frontend; backup + i18n.

## Constraints

- Migration must run on MySQL, PostgreSQL and SQLite; number it after the latest in
  `db/sql/migrations` (v2.20.7 today — the plan's `v2.9.98` predates the 2.20 series).
- FK cascade on the junction; `GetSurveyRefs` so the UI can warn before deleting a used survey.
- HA: nothing held in memory; security: same permission gate as Environment.
- Since the plan was written `SurveyVar.DefaultValue` became the `SurveyVarDefaultValue`
  struct and `ValidateSurveyVar` enforces type/default compatibility (`db/Template.go`); the
  merged list must pass the same validation, and the task-API variable gate (default-deny, see
  `decisions.md`) must validate against the merged list, not the inline one.

## Open questions

- User-facing name: "Survey" vs "Survey Group" — Environment is shown as "Variable Group".
- Merge precedence on name collision: recommended inline wins, then surveys in assignment order;
  confirm with the maintainer.
- Hide inline survey vars behind an advanced toggle to steer users to reusable surveys, or keep
  visible (recommended for now). A one-click "extract to reusable survey" is a possible follow-up.
- Where the task launch dialog gets its vars (template payload vs fetch) — it must render the
  merged list.

## Key files and references

- Pattern to mirror: `db/Environment.go`, `db/Store.go` (`EnvironmentManager`,
  `TemplateManager`), `db/sql/environment.go`, `db/sql/template.go`,
  `api/projects/environment.go`, `api/router.go`, `web/src/views/project/Environment.vue`,
  `web/src/components/EnvironmentForm.vue`, `web/src/components/TemplateForm.vue`,
  `web/src/components/SurveyVars.vue`, `services/project/backup.go`.
- Issue: https://github.com/semaphoreui/semaphore/issues/2212.
- Source plan: `AGENTS/plans/2_20/survey-vars-reusable.md`.
