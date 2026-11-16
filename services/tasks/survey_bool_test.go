package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// A bool survey var reaches the task as a boolean: the task form stores a JSON
// boolean in the environment, and an answer that arrives as the string "true"
// or "false" (API call, integration, workflow binding) is turned into one.

func boolSurveyExecutor(environment string, target db.SurveyVarTarget) *LocalExecutor {
	return &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "run.sh",
			SurveyVars: []db.SurveyVar{
				{Name: "dry_run", Title: "Dry run", Type: db.SurveyVarBool, Target: target},
			},
		},
		Environment: db.Environment{JSON: environment},
	}
}

func TestGetEnvironmentExtraVars_SurveyBool(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        any
		wantPresent bool
	}{
		{"JSON true", `{"dry_run":true}`, true, true},
		{"JSON false", `{"dry_run":false}`, false, true},
		{"string true from an old client", `{"dry_run":"true"}`, true, true},
		{"string false from an old client", `{"dry_run":"false"}`, false, true},
		{"other string left as is", `{"dry_run":"yes"}`, "yes", true},
		{"capitalised string left as is", `{"dry_run":"True"}`, "True", true},
		{"number left as is", `{"dry_run":1}`, float64(1), true},
		{"absent stays absent", `{}`, nil, false},
		{"empty environment", ``, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupExecutorConfig(t)
			exec := boolSurveyExecutor(tt.environment, db.SurveyVarTargetDefault)

			extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
			require.NoError(t, err)

			value, ok := extraVars["dry_run"]
			assert.Equal(t, tt.wantPresent, ok)
			assert.Equal(t, tt.want, value)
		})
	}
}

// Only a var declared bool is converted: a string var answered "true" stays a
// string, as the extra vars file promises for webhook values.
func TestGetEnvironmentExtraVars_StringVarAnsweredTrueStaysString(t *testing.T) {
	setupExecutorConfig(t)
	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{{Name: "flag", Title: "Flag"}},
		},
		Environment: db.Environment{JSON: `{"flag":"true","undeclared":"false"}`},
	}

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.Equal(t, "true", extraVars["flag"])
	assert.Equal(t, "false", extraVars["undeclared"])
}

func TestGetShellArgs_SurveyBool(t *testing.T) {
	tests := []struct {
		environment string
		want        string
	}{
		{`{"dry_run":true}`, "dry_run=true"},
		{`{"dry_run":false}`, "dry_run=false"},
		{`{"dry_run":"true"}`, "dry_run=true"},
		{`{"dry_run":"false"}`, "dry_run=false"},
	}
	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			setupExecutorConfig(t)
			exec := boolSurveyExecutor(tt.environment, db.SurveyVarTargetDefault)

			args, err := exec.getShellArgs("admin", nil)
			require.NoError(t, err)

			assert.Contains(t, args, tt.want)
		})
	}
}

func TestGetTerraformArgs_SurveyBool(t *testing.T) {
	tests := []struct {
		environment string
		want        string
	}{
		{`{"dry_run":true}`, "dry_run=true"},
		{`{"dry_run":false}`, "dry_run=false"},
		{`{"dry_run":"true"}`, "dry_run=true"},
	}
	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			setupExecutorConfig(t)
			exec := boolSurveyExecutor(tt.environment, db.SurveyVarTargetDefault)
			exec.Template.App = db.AppTerraform

			argsMap, err := exec.getTerraformArgs("admin", nil)
			require.NoError(t, err)

			defaultArgs := argsMap["default"]
			i := indexOfArg(defaultArgs, tt.want)
			require.Positive(t, i, "expected -var %s in %v", tt.want, defaultArgs)
			assert.Equal(t, "-var", defaultArgs[i-1])
		})
	}
}

func TestGetSurveyEnvVars_SurveyBool(t *testing.T) {
	tests := []struct {
		environment string
		want        []string
	}{
		{`{"dry_run":true}`, []string{"dry_run=true"}},
		{`{"dry_run":false}`, []string{"dry_run=false"}},
		{`{"dry_run":"true"}`, []string{"dry_run=true"}},
		{`{}`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			setupExecutorConfig(t)
			exec := boolSurveyExecutor(tt.environment, db.SurveyVarTargetEnv)

			envVars, err := exec.getSurveyEnvVars()
			require.NoError(t, err)
			assert.Equal(t, tt.want, envVars)

			extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
			require.NoError(t, err)
			assert.NotContains(t, extraVars, "dry_run", "an env-target var is not an extra var")
		})
	}
}

// The Ansible extra vars file carries the answer as an untagged YAML boolean,
// which a playbook can use in `when: dry_run` without the bool filter.
func TestGetPlaybookArgs_SurveyBoolInExtraVarsFile(t *testing.T) {
	tests := []struct {
		environment string
		want        bool
	}{
		{`{"dry_run":true}`, true},
		{`{"dry_run":false}`, false},
		{`{"dry_run":"true"}`, true},
		{`{"dry_run":"false"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			executor := newExtraVarsExecutor(t)
			executor.Template.SurveyVars = []db.SurveyVar{
				{Name: "dry_run", Title: "Dry run", Type: db.SurveyVarBool},
			}
			executor.Environment.JSON = tt.environment

			_, _, err := executor.getPlaybookArgs("admin", nil)
			require.NoError(t, err)
			defer executor.Cleanup()

			content, err := os.ReadFile(executor.extraVarsFile)
			require.NoError(t, err)
			assert.Equal(t, "!!bool", tagsOf(t, content)[".dry_run"], "%s", content)

			var vars map[string]any
			require.NoError(t, yaml.Unmarshal(content, &vars))
			assert.Equal(t, tt.want, vars["dry_run"])
		})
	}
}

// The extra vars of a bool survey var, loaded by Ansible itself: a boolean,
// whether the task form sent one or the API sent its string form. Skipped
// without an Ansible installation (see ansiblePython).
func TestSurveyBool_AnsibleLoaderReadsBoolean(t *testing.T) {
	python := ansiblePython(t)
	setupExecutorConfig(t)

	executor := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "from_form", Title: "From form", Type: db.SurveyVarBool},
				{Name: "from_api", Title: "From API", Type: db.SurveyVarBool},
				{Name: "as_string", Title: "String var"},
			},
		},
		Environment: db.Environment{JSON: `{"from_form":false,"from_api":"true","as_string":"true"}`},
	}
	vars, err := executor.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)
	document, err := extraVarsYAML(vars)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, python, "-c", ansibleLoaderScript)
	cmd.Stdin = bytes.NewReader(document)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	require.NoError(t, err, "ansible loader failed: %s", stderr.String())

	var leaves map[string]ansibleLeaf
	require.NoError(t, json.Unmarshal(stdout, &leaves), "%s", stdout)

	assert.Equal(t, ansibleLeaf{Value: false, Type: "bool"}, leaves[".from_form"])
	assert.Equal(t, ansibleLeaf{Value: true, Type: "bool"}, leaves[".from_api"])
	assert.Equal(t, ansibleLeaf{Value: "true", Type: "str"}, leaves[".as_string"], "only a bool var is converted")
}
