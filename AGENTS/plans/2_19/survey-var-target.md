# Survey Variable Target (env vs CLI) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `target` field to Survey Variables that controls how the variable is delivered to the task process: the current app-specific CLI way (default) or as a process environment variable.

**Architecture:** `SurveyVar` gets a `Target` field (`""` = CLI, `"env"` = environment variable) stored in the existing `survey_vars` JSON column — no DB migration. At run time `LocalExecutor` (used by both server and remote runner) excludes env-target vars from the extra-vars map and appends them to the process environment instead. The env var name is the survey var name verbatim (users who need `TF_VAR_foo` name the variable `TF_VAR_foo`).

**Tech Stack:** Go (backend), Vue 2 + Vuetify (frontend), testify for tests.

## Global Constraints

- Version: 2.19 (branch `2-19-stable`).
- No global variables (project rule).
- Tests use `github.com/stretchr/testify` `assert`/`require` (project rule, see `.claude/CLAUDE.md`).
- No DB migration: survey vars are serialized JSON in the `survey_vars` column; a new optional JSON field is backward and forward compatible.
- Empty string `""` must remain the default target and preserve current behavior exactly (extra-vars / -var / CLI args).
- Do not touch `web/public/swagger/api-docs.yml` — only root `api-docs.yml` (matches how the concurrent `text` survey type change was done).

## Background for the implementer (read before Task 1)

Data flow of a survey variable value:

1. Template stores definitions in `db.Template.SurveyVars` (`db/Template.go:103-111`), serialized to the `survey_vars` column as JSON via `SurveyVarsJSON`.
2. At task start, values entered by the user live in `Task.Environment` (plain vars) and `Task.Secret` (secret-type vars).
3. `TaskRunner.populateTaskEnvironment` (`services/tasks/TaskRunner.go:395`) merges `Task.Environment` into `Environment.JSON`.
4. `TaskPool.AddTask` (`services/tasks/TaskPool.go:1035`) builds a `LocalExecutor` with `Template`, `Environment`, and `Secret`.
5. `LocalExecutor.getEnvironmentExtraVars` (`services/tasks/local_executor.go:144`) merges `Environment.JSON` + `Secret` into one map, which then becomes:
   - `--extra-vars <json>` for Ansible (`getPlaybookArgs`, line ~488),
   - `-var name=value` for Terraform/Tofu/Terragrunt (`getTerraformArgs`, line ~320),
   - `name=value` CLI arguments for shell apps (`getShellArgs`, line ~281).
6. Process env vars are assembled in `LocalExecutor.Prepare` (line ~732) starting from `getEnvironmentENV()`.

The remote runner receives the full `db.Template` as JSON (`services/runners/types.go:16` — `Template db.Template \`json:"template"\``) and runs the same `LocalExecutor`, so changing `LocalExecutor` covers both execution modes. `SurveyVars` has the json tag `survey_vars`, so it survives the transfer.

---

### Task 1: Backend model — `SurveyVar.Target` field + validation

**Files:**
- Modify: `db/Template.go` (SurveyVar struct at line ~103, constants near line ~63, `Validate()` at line ~215)
- Test: `db/Template_test.go` (create if missing; check first — `db/Store_test.go` exists but this test belongs next to `Template.go`)

**Interfaces:**
- Produces: `db.SurveyVarTarget` type; constants `db.SurveyVarTargetDefault` (`""`), `db.SurveyVarTargetEnv` (`"env"`); field `db.SurveyVar.Target` (`json:"target,omitempty"`). Task 2 consumes `v.Target == db.SurveyVarTargetEnv`.

- [ ] **Step 1: Write the failing test**

Add to `db/Template_test.go` (same package `db`). If the file does not exist, create it with this content; if it exists, append the test functions:

```go
package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSurveyVarTarget_JSONRoundTrip(t *testing.T) {
	v := SurveyVar{Name: "MY_VAR", Title: "My Var", Target: SurveyVarTargetEnv}

	data, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"target":"env"`)

	// Default target must be omitted for backward compatibility.
	v.Target = SurveyVarTargetDefault
	data, err = json.Marshal(v)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "target")

	var parsed SurveyVar
	require.NoError(t, json.Unmarshal([]byte(`{"name":"X","target":"env"}`), &parsed))
	assert.Equal(t, SurveyVarTargetEnv, parsed.Target)
}

func TestTemplateValidate_SurveyVarTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  SurveyVarTarget
		wantErr bool
	}{
		{"default target is valid", SurveyVarTargetDefault, false},
		{"env target is valid", SurveyVarTargetEnv, false},
		{"unknown target is rejected", SurveyVarTarget("bogus"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := Template{
				Name:     "test",
				Playbook: "playbook.yml",
				// App left empty ("") so Validate skips the util.Config.Apps whitelist check.
				SurveyVars: []SurveyVar{{Name: "V", Title: "V", Target: tt.target}},
			}
			err := tpl.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, "invalid survey variable target")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./db/ -run 'TestSurveyVarTarget_JSONRoundTrip|TestTemplateValidate_SurveyVarTarget' -v -count=1`
Expected: compile error — `undefined: SurveyVarTargetEnv` (and `Target` field missing).

- [ ] **Step 3: Implement the model change**

In `db/Template.go`, after the `SurveyVarType` constants block (line ~70), add:

```go
type SurveyVarTarget string

const (
	// SurveyVarTargetDefault passes the variable the app-specific way:
	// --extra-vars for Ansible, -var for Terraform apps, CLI argument for shell apps.
	SurveyVarTargetDefault SurveyVarTarget = ""
	// SurveyVarTargetEnv passes the variable as a process environment variable.
	SurveyVarTargetEnv SurveyVarTarget = "env"
)
```

In the `SurveyVar` struct (line ~103), add the field after `Type`:

```go
type SurveyVar struct {
	Name         string               `json:"name" backup:"name"`
	Title        string               `json:"title" backup:"title"`
	Required     bool                 `json:"required,omitempty" backup:"required"`
	Type         SurveyVarType        `json:"type,omitempty" backup:"type"`
	Target       SurveyVarTarget      `json:"target,omitempty" backup:"target"`
	Description  string               `json:"description,omitempty" backup:"description"`
	Values       []SurveyVarEnumValue `json:"values,omitempty" backup:"values"`
	DefaultValue string               `json:"default_value,omitempty" backup:"default_value"`
}
```

In `Template.Validate()` (line ~215), before the final `return nil`, add:

```go
	for _, v := range tpl.SurveyVars {
		switch v.Target {
		case SurveyVarTargetDefault, SurveyVarTargetEnv:
		default:
			return &ValidationError{"invalid survey variable target: " + string(v.Target)}
		}
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./db/ -run 'TestSurveyVarTarget_JSONRoundTrip|TestTemplateValidate_SurveyVarTarget' -v -count=1`
Expected: PASS (all subtests).

Also run the whole package to catch regressions: `go test ./db/ -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add db/Template.go db/Template_test.go
git commit -m "feat(survey): add target field to survey vars (cli vs env)"
```

---

### Task 2: Executor — deliver env-target vars via process environment

**Files:**
- Modify: `services/tasks/local_executor.go` (`getEnvironmentExtraVars` at line ~144, `Prepare` at line ~732)
- Test: `services/tasks/local_executor_test.go` (append)

**Interfaces:**
- Consumes: `db.SurveyVarTargetEnv`, `db.SurveyVar.Target` from Task 1.
- Produces: `(t *LocalExecutor) getSurveyEnvVars() (res []string, err error)` — returns `NAME=value` pairs for survey vars with `Target == db.SurveyVarTargetEnv`; `getEnvironmentExtraVars` no longer returns env-target vars.

- [ ] **Step 1: Write the failing tests**

Append to `services/tasks/local_executor_test.go`:

```go
// TestGetEnvironmentExtraVars_SkipsEnvTargetVars verifies that survey vars with
// Target "env" are excluded from the extra-vars map (they must not reach
// --extra-vars / -var / shell CLI args).
func TestGetEnvironmentExtraVars_SkipsEnvTargetVars(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "ENV_VAR", Target: db.SurveyVarTargetEnv},
				{Name: "CLI_VAR"},
			},
		},
		Environment: db.Environment{JSON: `{"ENV_VAR":"via-env","CLI_VAR":"via-cli"}`},
		Secret:      `{"SECRET_ENV_VAR":"s3cr3t"}`,
	}
	exec.Template.SurveyVars = append(exec.Template.SurveyVars,
		db.SurveyVar{Name: "SECRET_ENV_VAR", Type: "secret", Target: db.SurveyVarTargetEnv})

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.NotContains(t, extraVars, "ENV_VAR")
	assert.NotContains(t, extraVars, "SECRET_ENV_VAR")
	assert.Equal(t, "via-cli", extraVars["CLI_VAR"])
}

// TestGetSurveyEnvVars verifies env-target survey vars are collected as
// NAME=value pairs from both Environment.JSON and the Secret field, and that
// CLI-target vars are ignored.
func TestGetSurveyEnvVars(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "ENV_VAR", Target: db.SurveyVarTargetEnv},
				{Name: "SECRET_ENV_VAR", Type: "secret", Target: db.SurveyVarTargetEnv},
				{Name: "CLI_VAR"},
				{Name: "MISSING_VAR", Target: db.SurveyVarTargetEnv}, // no value provided
			},
		},
		Environment: db.Environment{JSON: `{"ENV_VAR":"via-env","CLI_VAR":"via-cli"}`},
		Secret:      `{"SECRET_ENV_VAR":"s3cr3t"}`,
	}

	envVars, err := exec.getSurveyEnvVars()
	require.NoError(t, err)

	assert.Contains(t, envVars, "ENV_VAR=via-env")
	assert.Contains(t, envVars, "SECRET_ENV_VAR=s3cr3t")
	assert.Len(t, envVars, 2, "CLI-target and valueless vars must not be included")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./services/tasks/ -run 'TestGetEnvironmentExtraVars_SkipsEnvTargetVars|TestGetSurveyEnvVars' -v -count=1`
Expected: compile error — `exec.getSurveyEnvVars undefined`.

- [ ] **Step 3: Implement**

In `services/tasks/local_executor.go`:

**(a)** In `getEnvironmentExtraVars` (line ~144), after the `t.Secret` merge block and before the `semaphore_vars` block, add:

```go
	// Survey vars with the "env" target are delivered as process environment
	// variables (see getSurveyEnvVars), not as extra vars / CLI args.
	for _, v := range t.Template.SurveyVars {
		if v.Target == db.SurveyVarTargetEnv {
			delete(extraVars, v.Name)
		}
	}
```

**(b)** Add a new method right after `getEnvironmentENV` (line ~216):

```go
// getSurveyEnvVars returns NAME=value pairs for survey vars with Target "env".
// Values are read from the merged task environment (Environment.JSON) and the
// Secret field — the same sources getEnvironmentExtraVars reads, which excludes
// these vars so each one is delivered exactly once.
func (t *LocalExecutor) getSurveyEnvVars() (res []string, err error) {
	vars := make(map[string]any)

	if t.Environment.JSON != "" {
		if err = json.Unmarshal([]byte(t.Environment.JSON), &vars); err != nil {
			return
		}
	}

	if t.Secret != "" {
		secretVars := make(map[string]any)
		if err = json.Unmarshal([]byte(t.Secret), &secretVars); err != nil {
			return
		}
		maps.Copy(vars, secretVars)
	}

	for _, v := range t.Template.SurveyVars {
		if v.Target != db.SurveyVarTargetEnv {
			continue
		}
		if val, ok := vars[v.Name]; ok {
			res = append(res, fmt.Sprintf("%s=%v", v.Name, val))
		}
	}

	return
}
```

**(c)** In `Prepare` (line ~732), right after the `getEnvironmentENV()` call and its error check, add:

```go
	surveyEnvVars, err := t.getSurveyEnvVars()
	if err != nil {
		return
	}
	environmentVariables = append(environmentVariables, surveyEnvVars...)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./services/tasks/ -count=1`
Expected: PASS (including the pre-existing `TestGetShellArgs_PassesSurveySecretVar` and `TestGetEnvironmentExtraVars_MergesSecret` — default-target behavior is unchanged).

Also build everything: `go build ./...`
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add services/tasks/local_executor.go services/tasks/local_executor_test.go
git commit -m "feat(survey): pass env-target survey vars via process environment"
```

---

### Task 3: Frontend — target selector in survey var dialog

**Files:**
- Modify: `web/src/components/SurveyVars.vue` (dialog form at line ~36, `data()` at line ~235)
- Modify: `web/src/lang/en.js` (near line ~157, next to `default_value`)

**Interfaces:**
- Consumes: `target` JSON field from Task 1 (`""` or `"env"`). No other component changes: `TaskForm.vue` / `TaskParamsForm.vue` render inputs by `type` only — target does not affect how the value is entered, only how the backend delivers it.

- [ ] **Step 1: Add the select to the dialog**

In `web/src/components/SurveyVars.vue`, immediately after the existing `type` `v-select` (ends line ~44), add:

```vue
            <v-select
              v-model="editedVar.target"
              :label="$t('survey_var_target')"
              :items="varTargets"
              item-value="id"
              item-text="name"
              dense
              outlined
            ></v-select>
```

- [ ] **Step 2: Add the items list**

In `data()` (line ~235), after the `varTypes` array, add:

```js
      varTargets: [
        {
          id: '',
          name: 'CLI argument (extra-vars / -var / argument)',
        },
        {
          id: 'env',
          name: 'Environment variable',
        },
      ],
```

(Hardcoded English names match the existing `varTypes` convention in this file.)

- [ ] **Step 3: Add the translation key**

In `web/src/lang/en.js`, next to `default_value` (line ~157), add:

```js
  survey_var_target: 'Pass variable as',
```

(Other language files fall back to English for missing keys — only `en.js` is required, matching how recent keys were added.)

- [ ] **Step 4: Lint the frontend**

Run: `cd web && npm run lint`
Expected: no errors (warnings pre-existing at worst).

- [ ] **Step 5: Manual smoke check (optional but cheap)**

Run `cd web && npm run serve`, open a template, add a survey variable, verify the "Pass variable as" select shows both options, save, reopen — the value persists (it round-trips through `survey_vars` JSON).

- [ ] **Step 6: Commit**

```bash
git add web/src/components/SurveyVars.vue web/src/lang/en.js
git commit -m "feat(survey): UI for survey var target selection"
```

---

### Task 4: API docs

**Files:**
- Modify: `api-docs.yml` (`TemplateSurveyVar` definition at line ~1043)

- [ ] **Step 1: Add the `target` property**

In `api-docs.yml`, in `TemplateSurveyVar` properties (after `type`, line ~1055), add:

```yaml
      target:
        type: string
        enum: ["", env] # CLI (extra-vars / -var / argument) => "", Environment variable => "env"
        example: env
```

- [ ] **Step 2: Validate the YAML**

Run: `python3 -c "import yaml; yaml.safe_load(open('api-docs.yml'))" && echo OK`
Expected: `OK`

- [ ] **Step 3: Commit**

```bash
git add api-docs.yml
git commit -m "docs(api): survey var target field"
```

---

## Design decisions (recorded for reviewers)

1. **Field name `target`, values `""` / `"env"`** — mirrors the existing `EnvironmentSecretType` precedent (`db/Environment.go`: `"var"` / `"env"`). Empty default means zero migration and unchanged behavior for every existing template.
2. **Env var name is the survey var name verbatim.** No auto-prefixing (`TF_VAR_`, etc.) — Terraform users name the variable `TF_VAR_foo` themselves. Keeps behavior app-agnostic and predictable.
3. **Single implementation point in `LocalExecutor`** covers server-local execution and remote runners (runner deserializes the full `db.Template` including `survey_vars`).
4. **Secret survey vars respect `target` too**: an env-target secret goes to the process environment instead of CLI/extra-vars (arguably the main use case — env vars don't appear in process listings, unlike CLI args).
5. **Not done (YAGNI):** per-app target overrides, name prefix option, validation that env-target var names are valid shell identifiers, translations beyond `en.js`.

## Out of scope / do not touch

- The uncommitted `text` survey var type work (`db/Template.go`, `SurveyVars.vue`, `TaskForm.vue`, `TaskParamsForm.vue`, `api-docs.yml`) is a separate in-flight change on this branch. This plan's edits are adjacent but independent; do not revert or absorb them.
- `web/public/swagger/api-docs.yml` — not updated by hand in recent survey changes.
- Schedules (`TaskParamsForm.vue`) — no change needed; target does not affect value entry.
