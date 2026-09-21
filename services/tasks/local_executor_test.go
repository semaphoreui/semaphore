package tasks

import (
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExecutorConfig(t *testing.T) {
	t.Helper()
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
}

func TestGetPlaybookArgs_UsesRepositoryRootedPaths(t *testing.T) {
	setupExecutorConfig(t)

	repoRoot := t.TempDir()
	wd := "ansible"
	executor := LocalExecutor{
		Template: db.Template{
			Playbook:         "playbooks/site.yml",
			WorkingDirectory: &wd,
		},
		Inventory: db.Inventory{
			Type:      db.InventoryFile,
			Inventory: "inventories/production.ini",
		},
		Repository: db.Repository{GitURL: repoRoot},
	}

	args, _, err := executor.getPlaybookArgs("", nil)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(repoRoot, "inventories", "production.ini"), args[1])
	assert.Equal(t, filepath.Join(repoRoot, "playbooks", "site.yml"), args[len(args)-1])
}

// TestGetShellArgs_PassesSurveySecretVar verifies that Survey variables of type
// "Secret" (delivered via LocalExecutor.Secret) are passed to Bash/Shell tasks,
// alongside plain survey vars that arrive in Environment.JSON.
func TestGetShellArgs_PassesSurveySecretVar(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "date.sh",
		},
		Environment: db.Environment{
			JSON: `{"PLAIN_VAR":"hello"}`,
		},
		Secret: `{"MY_VAR":"s3cr3t"}`,
	}

	args, err := exec.getShellArgs("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, args, "PLAIN_VAR=hello", "plain survey var must be passed")
	assert.Contains(t, args, "MY_VAR=s3cr3t", "secret survey var must be passed")
}

// TestGetEnvironmentExtraVars_MergesSecret checks the shared helper merges the
// Secret field into the extra vars map used by Shell and Terraform tasks.
func TestGetEnvironmentExtraVars_MergesSecret(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Environment: db.Environment{JSON: `{"PLAIN_VAR":"hello"}`},
		Secret:      `{"MY_VAR":"s3cr3t"}`,
	}

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.Equal(t, "hello", extraVars["PLAIN_VAR"])
	assert.Equal(t, "s3cr3t", extraVars["MY_VAR"])
}

// TestGetEnvironmentExtraVars_EmptySecret pins that "" and "{}" are equivalent
// "no survey secrets" values: "" comes from DB-loaded tasks and API clients
// omitting the field, "{}" from the UI and the AddTask sanitization.
func TestGetEnvironmentExtraVars_EmptySecret(t *testing.T) {
	setupExecutorConfig(t)

	for _, secret := range []string{"", "{}"} {
		t.Run("secret "+secret, func(t *testing.T) {
			exec := &LocalExecutor{
				Environment: db.Environment{JSON: `{"PLAIN_VAR":"hello"}`},
				Secret:      secret,
			}

			extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
			require.NoError(t, err)

			assert.Equal(t, "hello", extraVars["PLAIN_VAR"])
			assert.Len(t, extraVars, 2) // PLAIN_VAR + semaphore_vars only
		})
	}
}

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

// TestFormatVarValue verifies values from multi-select survey vars (arrays)
// are JSON-encoded instead of degrading to Go's fmt representation.
func TestFormatVarValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"array", []any{"1", "2"}, `["1","2"]`},
		{"empty array", []any{}, `[]`},
		{"object", map[string]any{"k": "v"}, `{"k":"v"}`},
		{"number", float64(42), "42"},
		{"bool", true, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatVarValue(tt.input))
		})
	}
}

// TestGetSurveyEnvVars_MultiSelect verifies env-target multi-select survey
// vars are delivered as JSON arrays, not Go-formatted slices like "[1 2]".
func TestGetSurveyEnvVars_MultiSelect(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "MULTI_VAR", Type: db.SurveyVarSelect, Target: db.SurveyVarTargetEnv},
			},
		},
		Environment: db.Environment{JSON: `{"MULTI_VAR":["1","2"]}`},
	}

	envVars, err := exec.getSurveyEnvVars()
	require.NoError(t, err)

	assert.Contains(t, envVars, `MULTI_VAR=["1","2"]`)
}

// TestGetShellArgs_MultiSelect verifies multi-select survey vars reach shell
// tasks as JSON arrays in KEY=value CLI args, not Go's "[1 2]" formatting.
func TestGetShellArgs_MultiSelect(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "run.sh",
		},
		Environment: db.Environment{JSON: `{"multi_var":["1","2"]}`},
	}

	args, err := exec.getShellArgs("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, args, `multi_var=["1","2"]`)
	assert.NotContains(t, args, "multi_var=[1 2]")
}

// TestGetTerraformArgs_MultiSelect verifies multi-select survey vars reach
// terraform as -var name=<JSON list>, which terraform parses as a list value.
func TestGetTerraformArgs_MultiSelect(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type: db.TemplateTask,
			App:  db.AppTerraform,
		},
		Environment: db.Environment{JSON: `{"multi_var":["1","2"]}`},
	}

	argsMap, err := exec.getTerraformArgs("admin", nil)
	require.NoError(t, err)

	defaultArgs := argsMap["default"]
	found := false
	for i, arg := range defaultArgs {
		if arg == "-var" && i+1 < len(defaultArgs) && defaultArgs[i+1] == `multi_var=["1","2"]` {
			found = true
		}
	}
	assert.True(t, found, "expected -var multi_var=[\"1\",\"2\"] in %v", defaultArgs)
}

// TestGetPlaybookArgs_HideDryRunAndDiff verifies that hide_dry_run / hide_diff
// are enforced when building the ansible-playbook command line.
func TestGetPlaybookArgs_HideDryRunAndDiff(t *testing.T) {
	tests := []struct {
		name           string
		templateParams db.MapStringAnyField
		expectDryRun   bool
		expectDiff     bool
	}{
		{
			name:           "not hidden",
			templateParams: db.MapStringAnyField{},
			expectDryRun:   true,
			expectDiff:     true,
		},
		{
			name:           "legacy template without the keys keeps both",
			templateParams: nil,
			expectDryRun:   true,
			expectDiff:     true,
		},
		{
			name:           "dry run hidden",
			templateParams: db.MapStringAnyField{"hide_dry_run": true},
			expectDryRun:   false,
			expectDiff:     true,
		},
		{
			name:           "diff hidden",
			templateParams: db.MapStringAnyField{"hide_diff": true},
			expectDryRun:   true,
			expectDiff:     false,
		},
		{
			name:           "both hidden",
			templateParams: db.MapStringAnyField{"hide_dry_run": true, "hide_diff": true},
			expectDryRun:   false,
			expectDiff:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupExecutorConfig(t)

			exec := &LocalExecutor{
				Template: db.Template{
					Type:       db.TemplateTask,
					App:        db.AppAnsible,
					Playbook:   "site.yml",
					TaskParams: tt.templateParams,
				},
				Inventory: db.Inventory{Type: db.InventoryStatic},
				Task: db.Task{
					Params: db.MapStringAnyField{"dry_run": true, "diff": true},
				},
			}

			args, _, err := exec.getPlaybookArgs("admin", nil)
			require.NoError(t, err)

			if tt.expectDryRun {
				assert.Contains(t, args, "--check")
			} else {
				assert.NotContains(t, args, "--check")
			}

			if tt.expectDiff {
				assert.Contains(t, args, "--diff")
			} else {
				assert.NotContains(t, args, "--diff")
			}
		})
	}
}

// TestGetArgs_AnsibleForks verifies passing -f / --forks in template and task arguments,
// including invalid JSON error handling and AllowOverrideArgsInTask behavior.
func TestGetArgs_AnsibleForks(t *testing.T) {
	setupExecutorConfig(t)

	tests := []struct {
		name                  string
		templateArgs          *string
		allowOverride         bool
		taskArgs              *string
		expectedEffectiveFork string
		expectedForksSubArgs  []string
		mustNotContain        []string
		expectError           bool
		expectedErrorMsg      string
	}{
		{
			name:                  "Template arguments with --forks 10",
			templateArgs:          new(`["--forks", "10"]`),
			allowOverride:         false,
			taskArgs:              nil,
			expectedEffectiveFork: "10",
			expectedForksSubArgs:  []string{"--forks", "10"},
		},
		{
			name:                  "Template arguments with -f 10",
			templateArgs:          new(`["-f", "10"]`),
			allowOverride:         false,
			taskArgs:              nil,
			expectedEffectiveFork: "10",
			expectedForksSubArgs:  []string{"-f", "10"},
		},
		{
			name:                  "Task level overrides template forks when AllowOverrideArgsInTask is true",
			templateArgs:          new(`["--forks", "5"]`),
			allowOverride:         true,
			taskArgs:              new(`["--forks", "10"]`),
			expectedEffectiveFork: "10",
			expectedForksSubArgs:  []string{"--forks", "5", "--forks", "10"},
		},
		{
			name:                  "Task level ignored when AllowOverrideArgsInTask is false",
			templateArgs:          new(`["--forks", "5"]`),
			allowOverride:         false,
			taskArgs:              new(`["--forks", "10"]`),
			expectedEffectiveFork: "5",
			expectedForksSubArgs:  []string{"--forks", "5"},
			mustNotContain:        []string{"10"},
		},
		{
			name:             "Invalid JSON in template arguments returns descriptive error",
			templateArgs:     new(`--forks 10`),
			allowOverride:    false,
			taskArgs:         nil,
			expectError:      true,
			expectedErrorMsg: "invalid format of the template extra arguments, must be valid JSON",
		},
		{
			name:             "Invalid JSON in task arguments returns descriptive error when AllowOverrideArgsInTask is true",
			templateArgs:     new(`["--forks", "5"]`),
			allowOverride:    true,
			taskArgs:         new(`invalid-json`),
			expectError:      true,
			expectedErrorMsg: "invalid format of the TaskRunner extra arguments, must be valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &LocalExecutor{
				Logger: task_logger.NopLogger{},
				Template: db.Template{
					Type:                    db.TemplateTask,
					Playbook:                "site.yml",
					Arguments:               tt.templateArgs,
					AllowOverrideArgsInTask: tt.allowOverride,
				},
				Task: db.Task{
					Arguments: tt.taskArgs,
				},
				Inventory: db.Inventory{
					Type: db.InventoryStatic,
				},
			}

			args, _, err := exec.getPlaybookArgs("admin", nil)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				return
			}

			require.NoError(t, err)

			// Check that expected forks sub-arguments are present in sequence
			if len(tt.expectedForksSubArgs) > 0 {
				found := false
				for i := 0; i <= len(args)-len(tt.expectedForksSubArgs); i++ {
					match := true
					for j := 0; j < len(tt.expectedForksSubArgs); j++ {
						if args[i+j] != tt.expectedForksSubArgs[j] {
							match = false
							break
						}
					}
					if match {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected sequence %v in generated args: %v", tt.expectedForksSubArgs, args)
			}

			// Verify the effective (last) fork argument value matches the expected setting
			if tt.expectedEffectiveFork != "" {
				var lastForkVal string
				for i := 0; i < len(args)-1; i++ {
					if args[i] == "--forks" || args[i] == "-f" {
						lastForkVal = args[i+1]
					}
				}
				assert.Equal(t, tt.expectedEffectiveFork, lastForkVal, "Expected final effective fork value to be %s", tt.expectedEffectiveFork)
			}

			// Verify disqualified values are absent if specified
			for _, prohibited := range tt.mustNotContain {
				assert.NotContains(t, args, prohibited, "Generated args must not contain overridden task arg %q: %v", prohibited, args)
			}

			// Verify playbook is the trailing argument
			assert.Equal(t, exec.resolvePlaybookFile(), args[len(args)-1], "Playbook must be the last argument")
		})
	}
}

// TestGetEnvironmentExtraVars_MergesEnvironmentSecretVars verifies Variable
// Group secrets of type "var" are merged into the shared extra-vars map
// (JSON for Ansible) so values with spaces/newlines survive, override
// same-named plain vars, and type "env" secrets stay out of extra-vars.
func TestGetEnvironmentExtraVars_MergesEnvironmentSecretVars(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Environment: db.Environment{
			JSON: `{"PLAIN_VAR":"from-json","SHARED_KEY":"plain"}`,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "SPACED", Secret: "value with spaces"},
				{Type: db.EnvironmentSecretVar, Name: "MULTILINE", Secret: "line1\nline2"},
				{Type: db.EnvironmentSecretVar, Name: "SHARED_KEY", Secret: "from-secret"},
				{Type: db.EnvironmentSecretEnv, Name: "ENV_ONLY", Secret: "should-not-appear"},
			},
		},
	}

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.Equal(t, "from-json", extraVars["PLAIN_VAR"])
	assert.Equal(t, "value with spaces", extraVars["SPACED"])
	assert.Equal(t, "line1\nline2", extraVars["MULTILINE"])
	assert.Equal(t, "from-secret", extraVars["SHARED_KEY"], "secret var must override plain JSON key")
	assert.NotContains(t, extraVars, "ENV_ONLY")

	jsonStr, err := exec.getEnvironmentExtraVarsJSON("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, jsonStr, `"SPACED":"value with spaces"`)
	assert.Contains(t, jsonStr, `"MULTILINE":"line1\nline2"`)
	assert.Contains(t, jsonStr, `"SHARED_KEY":"from-secret"`)
	assert.NotContains(t, jsonStr, "ENV_ONLY")
	assert.NotContains(t, jsonStr, "should-not-appear")
}

// TestGetPlaybookArgs_SecretVarsInJSON verifies Ansible receives environment
// secret vars inside the JSON --extra-vars payload, not as name=value.
func TestGetPlaybookArgs_SecretVarsInJSON(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Playbook: "site.yml",
		},
		Inventory: db.Inventory{
			Type:      db.InventoryStatic,
			Inventory: "localhost",
		},
		Environment: db.Environment{
			JSON: `{"PLAIN":"ok"}`,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "TOKEN", Secret: "value with spaces"},
				{Type: db.EnvironmentSecretEnv, Name: "ENV_SECRET", Secret: "env-only"},
			},
		},
		Repository: db.Repository{GitURL: t.TempDir()},
	}

	args, _, err := exec.getPlaybookArgs("admin", nil)
	require.NoError(t, err)

	foundJSON := false
	for i, arg := range args {
		if arg == "--extra-vars" && i+1 < len(args) {
			payload := args[i+1]
			// Must not be the old name=value transport.
			assert.NotEqual(t, "TOKEN=value with spaces", payload)
			if assert.Contains(t, payload, `"TOKEN":"value with spaces"`) {
				foundJSON = true
			}
			assert.Contains(t, payload, `"PLAIN":"ok"`)
			assert.NotContains(t, payload, "ENV_SECRET")
		}
	}
	assert.True(t, foundJSON, "expected TOKEN inside JSON --extra-vars, got %v", args)
	assert.NotContains(t, args, "TOKEN=value with spaces")
}

// TestGetShellArgs_EnvironmentSecretVar verifies shell tasks get environment
// secret vars via the merged extra-vars path (single KEY=value argv).
func TestGetShellArgs_EnvironmentSecretVar(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "run.sh",
		},
		Environment: db.Environment{
			JSON: `{"PLAIN":"ok"}`,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "TOKEN", Secret: "value with spaces"},
				{Type: db.EnvironmentSecretEnv, Name: "ENV_SECRET", Secret: "env-only"},
			},
		},
	}

	args, err := exec.getShellArgs("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, args, "PLAIN=ok")
	assert.Contains(t, args, "TOKEN=value with spaces")
	assert.NotContains(t, args, "ENV_SECRET=env-only")
}

// TestGetTerraformArgs_EnvironmentSecretVar verifies terraform -var values
// come from the merged extra-vars map, including secrets with spaces.
func TestGetTerraformArgs_EnvironmentSecretVar(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type: db.TemplateTask,
			App:  db.AppTerraform,
		},
		Environment: db.Environment{
			JSON: `{"PLAIN":"ok"}`,
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretVar, Name: "TOKEN", Secret: "value with spaces"},
				{Type: db.EnvironmentSecretEnv, Name: "ENV_SECRET", Secret: "env-only"},
			},
		},
	}

	argsMap, err := exec.getTerraformArgs("admin", nil)
	require.NoError(t, err)

	defaultArgs := argsMap["default"]
	assert.Contains(t, defaultArgs, "-var")
	assert.Contains(t, defaultArgs, "PLAIN=ok")
	assert.Contains(t, defaultArgs, "TOKEN=value with spaces")
	assert.NotContains(t, defaultArgs, "ENV_SECRET=env-only")
}
