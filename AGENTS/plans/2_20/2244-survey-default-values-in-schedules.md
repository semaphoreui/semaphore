# Fix Plan — Issue #2244: No Default Values for Survey Variables in Scheduled Tasks

- **Issue:** https://github.com/semaphoreui/semaphore/issues/2244
- **Discussion:** https://github.com/semaphoreui/semaphore/discussions/2226
- **Branch:** `develop`

## 1. Problem

Task templates can declare **Survey Variables** (`survey_vars`) that map to the
playbook's `vars_prompt` section. Each survey variable may have a `default_value`.

When a task is started **manually** from the web UI, the survey variables (and
their default values) are presented in a form and submitted as part of the
task's `environment` payload, so the playbook receives them as extra vars.

When the same template is started by a **schedule** (cron or run-at), the task
is created entirely on the backend — no form is involved — and **no survey
variable values, including default values, are applied**. The playbook's
`vars_prompt` then has no value and the task fails.

Reported behavior (issue comment by `lug-gh`):
- A scheduled task with survey variables that have default values **fails**.
- Restarting the same failed task from the web UI shows the variables already
  filled in (because the UI re-applies the defaults), and the task succeeds.

## 2. Root Cause

Survey variable default values are applied **only in the Vue frontend**, never
on the backend.

### Frontend applies defaults (the only place defaults are handled today)

`web/src/components/TaskForm.vue:380-390`

```js
const defaultVars = (this.template.survey_vars || [])
  .filter((s) => s.default_value)
  .reduce((res, curr) => ({ ...res, [curr.name]: curr.default_value }), {});

this.editedEnvironment = {
  ...defaultVars,
  ...this.editedEnvironment,
};
```

The identical logic exists in `web/src/components/TaskParamsForm.vue:304-312`.
The resulting `editedEnvironment` object is what becomes `task.environment`.

### Backend never touches survey defaults

- `db/Template.go:94-102` — `SurveyVar` struct has `DefaultValue`, but it is
  only stored/serialized; nothing on the backend reads it to build a task
  environment.
- A repo-wide search for `SurveyVar` in `api/`, `services/`, `db_lib/` returns
  **no** code that applies defaults — only persistence (`db/`) and backup
  (`services/project/backup.go`).

### Scheduled task creation path

`services/schedules/SchedulePool.go:159-176` — `ScheduleRunner.Run()`:

```go
var task db.Task
if schedule.TaskParams != nil {
    task = schedule.TaskParams.CreateTask(schedule.TemplateID)
} else {
    task = db.Task{ProjectID: schedule.ProjectID, TemplateID: schedule.TemplateID}
}
task.ScheduleID = &schedule.ID
_, err = r.pool.taskPool.AddTask(task, nil, "", schedule.ProjectID, tpl.App.NeedTaskAlias())
```

`db.TaskParams.CreateTask` (`db/TaskParams.go:23-37`) copies only
`TaskParams.Environment` verbatim. A schedule with no `TaskParams` produces a
task with an **empty** `Environment`. Either way, survey defaults are absent.

### Where the environment is assembled at run time

`services/tasks/TaskRunner.go:293-331` — `populateTaskEnvironment()` merges the
template's environment JSON with the task's environment JSON:

```go
func (t *TaskRunner) populateTaskEnvironment() (err error) {
    if t.Task.Environment == "" {
        return            // <-- early return: nothing merged for empty task env
    }
    ...
    for k, v := range taskEnvironment {  // task env overrides template env
        tplEnvironment[k] = v
    }
    t.Environment.JSON = string(ev)
}
```

Note the **early return** when `t.Task.Environment == ""`: a scheduled task
with no environment skips this function entirely, so even if we add default
handling here we must remove/adjust that guard.

By the time `populateTaskEnvironment()` runs, `t.Template` is already loaded by
`populateDetails()` (`TaskRunner.go:338`) via `store.GetTemplate`, which calls
`db.FillTemplate` (`db/Template.go:271-280`) and unmarshals
`SurveyVarsJSON → Template.SurveyVars`. **So `t.Template.SurveyVars` (with
`DefaultValue`) is available inside `populateTaskEnvironment`.**

## 3. Proposed Fix

Apply survey variable default values **on the backend, at task run time**, in
`TaskRunner.populateTaskEnvironment()`.

Doing it in `TaskRunner` (rather than only in `SchedulePool`) is deliberate: it
fixes **every** non-UI trigger path with one change — schedules, integrations
(`api/integration.go` task creation), and direct API calls that omit survey
values — and keeps a single source of truth for the merge logic.

### Merge precedence

To replicate the manual-UI behavior exactly, the precedence (lowest → highest)
must be:

1. Template environment JSON (`t.Environment.JSON`) — base.
2. **Survey variable default values** — new.
3. Task environment JSON (`t.Task.Environment`) — explicit per-task values win.

This matches the frontend, where `editedEnvironment = { ...defaultVars,
...editedEnvironment }` lets an explicitly supplied value override a default.

### Type handling

The frontend inserts `default_value` as-is — a **string** — with no type
conversion, even for `int`/`enum` survey vars. The backend fix will do the
same (apply `DefaultValue` as a string) to stay byte-for-byte consistent with
manually triggered tasks. Changing type coercion is explicitly out of scope.

## 4. Implementation Steps

### Step 1 — Helper on `db.Template` to expose survey defaults

File: `db/Template.go`

Add a method that returns the survey-variable defaults as a map:

```go
// GetSurveyVarsDefaults returns the default values of survey variables that
// declare a non-empty default_value, keyed by variable name.
func (tpl *Template) GetSurveyVarsDefaults() map[string]any {
    defaults := make(map[string]any)
    for _, sv := range tpl.SurveyVars {
        if sv.DefaultValue != "" {
            defaults[sv.Name] = sv.DefaultValue
        }
    }
    return defaults
}
```

Rationale: keeps survey knowledge in the `db` package alongside the
`SurveyVar` definition and makes the behavior unit-testable in isolation.

### Step 2 — Apply defaults in `populateTaskEnvironment`

File: `services/tasks/TaskRunner.go` (function `populateTaskEnvironment`,
lines 293-331)

Changes:

1. **Remove the early return** on `t.Task.Environment == ""` so defaults are
   still merged when the task carries no environment (the scheduled-task case).
2. After unmarshaling the template environment into `tplEnvironment` and
   *before* applying `taskEnvironment`, layer in the survey defaults:

```go
func (t *TaskRunner) populateTaskEnvironment() (err error) {
    tplEnvironment := make(map[string]any)
    if t.Environment.JSON != "" {
        if err = json.Unmarshal([]byte(t.Environment.JSON), &tplEnvironment); err != nil {
            return
        }
    }

    // Apply survey variable default values. These sit above the template
    // environment but below explicit per-task values, mirroring the web UI.
    for k, v := range t.Template.GetSurveyVarsDefaults() {
        tplEnvironment[k] = v
    }

    taskEnvironment := make(map[string]any)
    if t.Task.Environment != "" {
        if err = json.Unmarshal([]byte(t.Task.Environment), &taskEnvironment); err != nil {
            return
        }
    }

    for k, v := range taskEnvironment {
        tplEnvironment[k] = v
    }

    ev, err := json.Marshal(tplEnvironment)
    if err != nil {
        return err
    }
    t.Environment.JSON = string(ev)
    return
}
```

Note: with the early return removed, the function may now set
`t.Environment.JSON` to `"{}"` when there is no template environment, no
survey defaults, and no task environment. Verify downstream consumers
(`LocalJob.getEnvironmentExtraVars`, `getEnvironmentExtraVarsJSON` in
`services/tasks/LocalJob.go`) tolerate `"{}"` — they unmarshal JSON, so an
empty object is fine. If a stricter guard is preferred, keep the function from
overwriting `t.Environment.JSON` when the merged map is empty and the original
template JSON was empty.

### Step 3 — (Optional) Persist applied defaults onto the task record

The fix above merges defaults at run time but does **not** write them back to
`task.Environment` in the database, so the Task History UI will still show an
empty environment for scheduled runs.

Optional improvement: in `services/schedules/SchedulePool.go` (`Run()`), after
fetching `tpl`, pre-fill the task environment from
`tpl.GetSurveyVarsDefaults()` merged under any `schedule.TaskParams.Environment`,
and assign the JSON to `task.Environment` before `AddTask`. This makes the
values visible/auditable in task history.

Recommendation: ship Step 2 as the core fix; treat Step 3 as a follow-up if
task-history visibility is desired. Steps 2 and 3 are not mutually exclusive —
Step 2's run-time merge stays correct even if Step 3 already populated the
task environment (the values are simply identical).

## 5. Testing

### Unit tests

**`db/Template_test.go`** — new `TestTemplate_GetSurveyVarsDefaults`:
- survey vars with defaults → only those with non-empty `DefaultValue` returned;
- empty `SurveyVars` → empty map;
- mix of `str`/`int`/`enum` types → default returned as string for all.

**`services/tasks/TaskRunner_test.go`** — extend
`TestTaskRunner_populateTaskEnvironment` (currently line 733) with table-driven
cases:

| Case | Template JSON | Survey defaults | Task env | Expected merged |
|------|---------------|-----------------|----------|-----------------|
| scheduled task, no task env | `{}` | `host=web1` | `""` | `{"host":"web1"}` |
| default overridden by task env | `{}` | `host=web1` | `{"host":"db1"}` | `{"host":"db1"}` |
| default + template env coexist | `{"a":1}` | `host=web1` | `""` | `{"a":1,"host":"web1"}` |
| no survey vars, no task env | `{}` | — | `""` | `{}` |
| existing behavior preserved | `{"a":1,"d":4}` | — | `{"a":11,"b":22,"c":33}` | `{"a":11,"b":22,"c":33,"d":4}` |

Each `TaskRunner` is built with a `Template{SurveyVars: [...]}` so
`GetSurveyVarsDefaults()` is exercised through `populateTaskEnvironment`.

Run:

```bash
go test ./db/ -run TestTemplate_GetSurveyVarsDefaults -v -count=1
go test ./services/tasks/ -run TestTaskRunner_populateTaskEnvironment -v -count=1
```

### Manual / integration verification

1. Create a template with a survey variable that has a `default_value` and a
   playbook whose `vars_prompt` uses it.
2. Create a cron (or run-at) schedule for that template with **no** task
   params.
3. Trigger the schedule (or wait for it) and confirm the task **succeeds** and
   the playbook receives the default value via `--extra-vars`.
4. Confirm a manually triggered task still behaves identically (regression).
5. Confirm a schedule whose `TaskParams.Environment` sets the same variable
   uses the explicit value, not the default.

## 6. Edge Cases & Out of Scope

- **Required survey variables without a default** — still cannot be satisfied
  by a schedule; the task will fail as before. Filling those requires the user
  to set values on the schedule's `TaskParams` (already supported). Out of
  scope for this fix; could be a follow-up validation/warning.
- **Type coercion** (`int`/`enum` defaults) — kept as strings to match the
  current frontend behavior; not changed here.
- **Per-schedule survey value editing UI** (the broader ask in discussion
  #2226) — already possible today via the schedule's task params; no UI work
  included in this plan.
- **Integrations / direct API task creation** — automatically benefit from the
  `TaskRunner` fix; no extra changes needed.

## 7. Files Touched

| File | Change |
|------|--------|
| `db/Template.go` | Add `GetSurveyVarsDefaults()` method |
| `services/tasks/TaskRunner.go` | Apply survey defaults in `populateTaskEnvironment`; remove empty-env early return |
| `db/Template_test.go` | New unit test for `GetSurveyVarsDefaults` |
| `services/tasks/TaskRunner_test.go` | Extend `populateTaskEnvironment` test (table-driven) |
| `services/schedules/SchedulePool.go` | *(Optional, Step 3)* pre-fill task environment with defaults |
