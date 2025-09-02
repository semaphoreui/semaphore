package tasks

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

// MockLogger implements task_logger.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Log(msg string)                           {}
func (m *MockLogger) Logf(format string, a ...any)            {}
func (m *MockLogger) LogWithTime(now time.Time, msg string)   {}
func (m *MockLogger) LogfWithTime(now time.Time, format string, a ...any) {}
func (m *MockLogger) LogCmd(cmd *exec.Cmd)                    {}
func (m *MockLogger) SetStatus(status task_logger.TaskStatus) {}
func (m *MockLogger) AddStatusListener(l task_logger.StatusListener) {}
func (m *MockLogger) AddLogListener(l task_logger.LogListener) {}
func (m *MockLogger) SetCommit(hash, message string)          {}
func (m *MockLogger) WaitLog()                                {}

// TestSecretsNotInCommandLine verifies that secrets are not exposed in command line arguments
func TestSecretsNotInCommandLine(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	// Test Ansible secrets
	t.Run("Ansible secrets not in args", func(t *testing.T) {
		job := &LocalJob{
			Task: db.Task{
				ID: 1,
			},
			Template: db.Template{
				App:      db.AppAnsible,
				Playbook: "test.yml",
				ProjectID: 0,
			},
			Inventory: db.Inventory{
				Type: db.InventoryStatic,
			},
			Environment: db.Environment{
				JSON: `{"public_var": "public_value"}`,
				Secrets: []db.EnvironmentSecret{
					{
						Name:   "secret_var",
						Secret: "secret_value",
						Type:   db.EnvironmentSecretVar,
					},
				},
			},
			Secret: `{"survey_secret": "survey_value"}`,
			Logger: &MockLogger{}, // Add mock logger
		}

		args, _, err := job.getPlaybookArgs("testuser", nil)
		assert.NoError(t, err)

		// Convert args to single string for easier checking
		argsStr := strings.Join(args, " ")
		
		// Secrets should NOT be in command line arguments
		assert.NotContains(t, argsStr, "secret_value", "Secret value should not be in command line arguments")
		assert.NotContains(t, argsStr, "survey_value", "Survey secret should not be in command line arguments")
		
		// Public values should still be accessible (via --extra-vars JSON)
		assert.Contains(t, argsStr, "--extra-vars", "Should still use --extra-vars for public variables")
	})

	// Test Terraform secrets
	t.Run("Terraform secrets not in args", func(t *testing.T) {
		job := &LocalJob{
			Task: db.Task{
				ID: 1,
			},
			Template: db.Template{
				App: db.AppTerraform,
				ProjectID: 0,
			},
			Environment: db.Environment{
				JSON: `{"public_var": "public_value"}`,
				Secrets: []db.EnvironmentSecret{
					{
						Name:   "secret_var",
						Secret: "secret_value",
						Type:   db.EnvironmentSecretVar,
					},
				},
			},
			Secret: `{"survey_secret": "survey_value"}`,
			Logger: &MockLogger{}, // Add mock logger
		}

		args, err := job.getTerraformArgs("testuser", nil)
		assert.NoError(t, err)

		argsStr := strings.Join(args, " ")
		
		// Secrets should NOT be in command line arguments
		assert.NotContains(t, argsStr, "secret_value", "Secret value should not be in Terraform command line arguments")
		assert.NotContains(t, argsStr, "survey_value", "Survey secret should not be in Terraform command line arguments")
	})

	// Test Shell secrets
	t.Run("Shell secrets not in args", func(t *testing.T) {
		job := &LocalJob{
			Task: db.Task{
				ID: 1,
			},
			Template: db.Template{
				App: db.AppBash,
				ProjectID: 0,
			},
			Environment: db.Environment{
				JSON: `{"public_var": "public_value"}`,
				Secrets: []db.EnvironmentSecret{
					{
						Name:   "secret_var",
						Secret: "secret_value",
						Type:   db.EnvironmentSecretVar,
					},
				},
			},
			Secret: `{"survey_secret": "survey_value"}`,
			Logger: &MockLogger{}, // Add mock logger
		}

		args, err := job.getShellArgs("testuser", nil)
		assert.NoError(t, err)

		argsStr := strings.Join(args, " ")
		
		// Secrets should NOT be in command line arguments
		assert.NotContains(t, argsStr, "secret_value", "Secret value should not be in Shell command line arguments")
		assert.NotContains(t, argsStr, "survey_value", "Survey secret should not be in Shell command line arguments")
	})
}

// TestSecretsInEnvironmentVariables verifies that secrets are properly passed via environment variables
func TestSecretsInEnvironmentVariables(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	// Test that getEnvironmentExtraVars excludes secrets (they're handled separately now)
	t.Run("Environment extra vars exclude secrets", func(t *testing.T) {
		job := &LocalJob{
			Environment: db.Environment{
				JSON: `{"public_var": "public_value"}`,
				Secrets: []db.EnvironmentSecret{
					{
						Name:   "secret_var",
						Secret: "secret_value",
						Type:   db.EnvironmentSecretVar,
					},
				},
			},
			Secret: `{"survey_secret": "survey_value"}`,
		}

		extraVars, err := job.getEnvironmentExtraVars("testuser", nil)
		assert.NoError(t, err)

		// Public variables should be available in extraVars
		assert.Equal(t, "public_value", extraVars["public_var"])
		
		// Secrets should NOT be in extraVars (they're handled separately for security)
		assert.Nil(t, extraVars["secret_var"], "Environment secrets should not be in public extraVars")
		assert.Nil(t, extraVars["survey_secret"], "Survey secrets should not be in public extraVars")
		
		// Built-in vars should still be present
		assert.NotNil(t, extraVars["semaphore_vars"], "Built-in semaphore_vars should be present")
	})
}

// Helper function to check if a map contains secrets by value
func containsSecret(vars map[string]any, secretValue string) bool {
	for _, value := range vars {
		if str, ok := value.(string); ok && str == secretValue {
			return true
		}
		if mapVal, ok := value.(map[string]any); ok && containsSecret(mapVal, secretValue) {
			return true
		}
	}
	return false
}