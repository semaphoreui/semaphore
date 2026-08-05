package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
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

// TestGetInventorySSHCommonArgs verifies the ssh options ansible receives to
// reach inventory hosts through a jump host.
func TestGetInventorySSHCommonArgs(t *testing.T) {
	setupExecutorConfig(t)

	user := "ansible-proxy"
	port := 2222

	t.Run("no proxy means no ssh args", func(t *testing.T) {
		e := &LocalExecutor{Inventory: db.Inventory{}}
		assert.Empty(t, e.getInventorySSHCommonArgs())
	})

	t.Run("proxy adds ProxyJump", func(t *testing.T) {
		e := &LocalExecutor{Inventory: db.Inventory{
			Proxy: &db.Proxy{
				Type: db.ProxySSH,
				Host: "bastion.example.org",
				User: &user,
				Port: &port,
			},
		}}

		assert.Equal(t, "-o ProxyJump=ansible-proxy@bastion.example.org:2222", e.getInventorySSHCommonArgs())
	})
}

// TestGetPlaybookArgs_Proxy verifies the jump host reaches the ansible-playbook
// command line, and that it is absent when the inventory has no proxy.
func TestGetPlaybookArgs_Proxy(t *testing.T) {
	setupExecutorConfig(t)

	inventoryID := 1

	newExecutor := func(proxy *db.Proxy) *LocalExecutor {
		return &LocalExecutor{
			Template:  db.Template{App: db.AppAnsible, Playbook: "test.yml"},
			Inventory: db.Inventory{ID: inventoryID, Type: db.InventoryStatic, Proxy: proxy},
		}
	}

	t.Run("without proxy", func(t *testing.T) {
		args, _, err := newExecutor(nil).getPlaybookArgs("admin", nil)

		require.NoError(t, err)
		assert.NotContains(t, args, "--ssh-common-args")
	})

	t.Run("with proxy", func(t *testing.T) {
		args, _, err := newExecutor(&db.Proxy{Type: db.ProxySSH, Host: "bastion.example.org"}).
			getPlaybookArgs("admin", nil)

		require.NoError(t, err)
		require.Contains(t, args, "--ssh-common-args")

		i := indexOf(args, "--ssh-common-args")
		require.Less(t, i+1, len(args))
		assert.Equal(t, "-o ProxyJump=bastion.example.org", args[i+1])
	})
}

func indexOf(args []string, value string) int {
	for i, a := range args {
		if a == value {
			return i
		}
	}
	return -1
}
