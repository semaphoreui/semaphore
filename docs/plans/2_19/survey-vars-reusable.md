# Implementation Plan — Reusable Survey Variables (Issue #2212)

## Goal

Survey Variables are currently defined **inline** on each Task Template (stored as a
JSON blob in `project__template.survey_vars`). The issue asks that Survey Variables
become a **separate, reusable entity** that can be created once and **assigned to many
Task Templates** — exactly like an Environment ("Variable Group") is today.

> "Survey Variables should be set separate to Task Templates and then assigned to a
> task template like Environment is. Then when you want to change survey variables,
> you don't have to modify all your task templates one by one."

## Design Summary

We introduce a new project-scoped entity — a **Survey** — which is a named collection
of `SurveyVar` definitions. Templates reference Surveys through a **M:N junction
table**, mirroring how `project__template_environment` links templates to
environments.

Key decisions:

1. **Reuse the Environment pattern verbatim.** Environment already solved the exact
   "make it a separate, multi-assignable entity" problem (entity table + junction
   table + manager interface + API subrouter + list view + form + multi-select in
   `TemplateForm.vue`). Survey mirrors this 1:1 to minimize design risk.
2. **M:N relationship** (a template can have several Surveys, a Survey can be used by
   many templates), matching `EnvironmentIDs`.
3. **Backward compatibility is mandatory.** The existing inline
   `Template.SurveyVars` field is kept and still works. At task-run time the
   effective survey variable list is the **merge** of inline vars + all assigned
   Surveys' vars. No data migration is forced on users.
4. Naming: entity = **Survey** (`db.Survey`), table = `project__survey`, junction =
   `project__template_survey`. UI label: "Surveys".

## Affected Areas (reference paths)

Current relevant code (from research):

- `db/Template.go` — `SurveyVar`, `SurveyVarEnumValue`, `SurveyVarType`, and
  `Template` struct (`SurveyVarsJSON`, `SurveyVars`).
- `db/Environment.go` — entity to mirror.
- `db/Store.go` — `EnvironmentManager` interface (lines ~321-330),
  `EnvironmentProps`, `TemplateManager` (`GetTemplateEnvironments`,
  `UpdateTemplateEnvironments`).
- `db/sql/environment.go`, `db/sql/template.go` — SQL CRUD to mirror.
- `db/sql/migrations/` — latest is `v2.9.97.sql`; new migration goes after it.
- `api/router.go` — environment subrouter (lines ~306-307, ~426-433) to mirror.
- `api/projects/` — environment + template controllers.
- `web/src/views/project/Environment.vue`, `web/src/components/EnvironmentForm.vue`,
  `web/src/components/TemplateForm.vue`, `web/src/components/SurveyVars.vue`.
- Task-run code that consumes `template.SurveyVars` (validation of survey answers
  when a task is created/run).

---

## Implementation Steps

### Phase 1 — Backend data model

**1.1 New `db/Survey.go`**

```go
type Survey struct {
    ID        int         `db:"id" json:"id" backup:"-"`
    ProjectID int         `db:"project_id" json:"project_id" backup:"-"`
    Name      string      `db:"name" json:"name" binding:"required"`

    // VarsJSON is the persisted form; Vars is the working form.
    VarsJSON  *string     `db:"vars" json:"-" backup:"-"`
    Vars      []SurveyVar `db:"-" json:"vars" backup:"vars"`
}
```

- Reuse the existing `SurveyVar` / `SurveyVarEnumValue` / `SurveyVarType` types from
  `db/Template.go` — do **not** duplicate them.
- Add a `FillSurvey` helper that unmarshals `VarsJSON` → `Vars` (mirror
  `FillTemplate`'s survey-vars handling in `db/Template.go`).

**1.2 `Template` struct (`db/Template.go`)**

Add, alongside `EnvironmentIDs`:

```go
// SurveyIDs is the list of reusable Surveys assigned to this template.
// Their variables are merged with the template's inline SurveyVars at run time.
SurveyIDs []int `db:"-" bolt:"include" json:"survey_ids" backup:"-"`
```

Keep `SurveyVarsJSON` / `SurveyVars` unchanged (inline vars, deprecated path but
still supported).

**1.3 `db/Store.go`**

- Add `SurveyProps` `ObjectProps` (table `project__survey`, sortable by `name`,
  referring suffix `survey_id`).
- Add a `SurveyManager` interface mirroring `EnvironmentManager`:
  ```go
  type SurveyManager interface {
      GetSurvey(projectID, surveyID int) (Survey, error)
      GetSurveyRefs(projectID, surveyID int) (ObjectReferrers, error)
      GetSurveys(projectID int, params RetrieveQueryParams) ([]Survey, error)
      CreateSurvey(survey Survey) (Survey, error)
      UpdateSurvey(survey Survey) error
      DeleteSurvey(projectID, surveyID int) error
  }
  ```
- Add to `TemplateManager`: `GetTemplateSurveys(projectID, templateID int) ([]int, error)`
  and `UpdateTemplateSurveys(projectID, templateID int, surveyIDs []int) error`.
- Wire `SurveyManager` into the aggregate `Store` interface.

### Phase 2 — Migration

New file `db/sql/migrations/v2.9.98.sql` (use the next available version; verify the
latest before assigning):

```sql
create table `project__survey` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `name` varchar(255) not null,
    `vars` longtext,
    foreign key (`project_id`) references `project`(`id`) on delete cascade
);

create table `project__template_survey` (
    `project_id` int not null,
    `template_id` int not null,
    `survey_id` int not null,
    primary key (`template_id`, `survey_id`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`template_id`) references `project__template`(`id`) on delete cascade,
    foreign key (`survey_id`) references `project__survey`(`id`) on delete cascade
);
```

- **No forced data migration.** Inline `project__template.survey_vars` stays.
- Confirm the migration runs on all three SQL dialects Semaphore supports
  (MySQL / PostgreSQL / SQLite) — check existing migrations for dialect-specific
  syntax handling.

### Phase 3 — Store implementations

**3.1 SQL — `db/sql/survey.go`** (new)

Mirror `db/sql/environment.go`: `GetSurvey`, `GetSurveys`, `CreateSurvey`,
`UpdateSurvey`, `DeleteSurvey` (via generic `deleteObject`), `GetSurveyRefs` (query
`project__template_survey`). Serialize `Vars` with `db.ObjectToJSON`.

**3.2 SQL — `db/sql/template.go`**

- Implement `GetTemplateSurveys` / `UpdateTemplateSurveys` against
  `project__template_survey` (copy the `*TemplateEnvironments` logic).
- In `CreateTemplate` / `UpdateTemplate` / `GetTemplate(s)`, also load & persist
  `SurveyIDs` (same place `EnvironmentIDs` is handled).

**3.3 Cascade on delete**

`DeleteSurvey` must remove junction rows (FK `on delete cascade` covers this).
`GetSurveyRefs` should let the UI warn / block deletion when a Survey is still
assigned.

### Phase 4 — REST API

**4.1 `api/router.go`** — add a survey subrouter mirroring environment:

```
GET    /api/project/{project_id}/surveys
POST   /api/project/{project_id}/surveys
GET    /api/project/{project_id}/surveys/{survey_id}
PUT    /api/project/{project_id}/surveys/{survey_id}
DELETE /api/project/{project_id}/surveys/{survey_id}
GET    /api/project/{project_id}/surveys/{survey_id}/refs
```

**4.2 `api/projects/surveys.go`** (new) — controller + `SurveyMiddleware` mirroring
`api/projects/environment.go`. Permission: gate create/update/delete on
`manageProjectResources` (same as Environment).

**4.3 Template controller** — accept and return `survey_ids`; call
`UpdateTemplateSurveys` from `AddTemplate` / `UpdateTemplate`.

### Phase 5 — Task run / survey-answer resolution

This is the behavioral core — find where `template.SurveyVars` is read when a task
is created/validated (validation of provided survey answers).

- Introduce a helper `EffectiveSurveyVars(template, surveys []Survey) []SurveyVar`
  that returns `template.SurveyVars` ++ each assigned Survey's `Vars`.
- On name collision, define precedence (recommend: inline template vars win, then
  Surveys in assignment order — document it; mirror environment merge semantics).
- Update task creation/validation to use the merged list instead of
  `template.SurveyVars` directly.
- The runner/task payload still receives a flat resolved list, so runner code is
  unaffected.

### Phase 6 — Frontend

**6.1 List view `web/src/views/project/Survey.vue`** (new) — mirror
`Environment.vue`: list, New/Edit/Delete, `manageProjectResources` gate. Add the
route + nav item next to "Environment".

**6.2 `web/src/components/SurveyForm.vue`** (new) — name field + an embedded survey
editor. **Reuse `SurveyVars.vue`'s editor logic** by extracting the variable-editing
portion into a shared child component so both the inline template editor and the new
Survey entity use identical UI.

**6.3 `web/src/components/TemplateForm.vue`** — add a multi-select for Surveys next
to the Environment multi-select (`v-model="item.survey_ids"`, `:items="surveys"`,
`multiple`). Keep the existing inline `SurveyVars.vue` editor (still supported).

**6.4 API client** — add survey CRUD calls; load the project's surveys for the
TemplateForm dropdown.

**6.5 Task launch dialog** — the survey-answer prompt must render the **merged**
effective survey vars. Confirm whether the dialog reads vars from the template
payload or fetches them; adjust so assigned Surveys' vars appear.

### Phase 7 — Backup / restore & i18n

- `Survey` already has `backup` struct tags. Update the project backup/restore code
  (search `backup:` usage for environments) to include Surveys and the
  template→survey assignments, so import/export round-trips.
- Add i18n strings for new UI labels in `web/src/lang/`.

### Phase 8 — Tests

Per `.claude/CLAUDE.md` (testify, `_test.go` next to source):

- `db/sql/survey_test.go` — CRUD + `GetSurveyRefs` + delete-cascade.
- Survey-vars merge helper test (precedence, collisions, empty cases) —
  table-driven.
- `api/projects/surveys_test.go` — handler tests via `httptest`.
- Template create/update tests asserting `survey_ids` round-trip.
- Backup/restore test including a Survey.

---

## Backward Compatibility & Rollout

- Existing templates keep working — inline `survey_vars` untouched.
- `survey_ids` is additive and optional everywhere (JSON `omitempty`-friendly).
- Optional follow-up (separate PR, not this issue): a one-click "Extract to reusable
  Survey" action on the template form to migrate inline vars into a new Survey.
- API stays additive — no breaking changes to existing endpoints.

## Open Questions

1. **Entity name** — "Survey" vs "Survey Group" / "Variable Survey". `Environment`
   is shown as "Variable Group" in newer UI; pick a consistent user-facing label.
2. **Merge precedence** on name collision between inline vars and multiple Surveys —
   confirm desired rule with maintainers.
3. Should inline template survey vars be **deprecated in the UI** (hidden behind an
   "advanced" toggle) to steer users toward reusable Surveys? Recommend keeping
   visible for now.

## Suggested PR Breakdown

1. PR 1 — DB model, migration, store interfaces + SQL implementation + tests.
2. PR 2 — REST API (survey controller, routes, template controller wiring) + tests.
3. PR 3 — Task-run merge logic + tests.
4. PR 4 — Frontend (list view, form, TemplateForm multi-select, task launch dialog).
5. PR 5 — Backup/restore + i18n.
