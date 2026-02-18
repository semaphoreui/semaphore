package tasks

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

// TestTaskRunner_populateTaskEnvironment_EmptyTemplateEnv tests the bug fix for
// integration extractor variables not being passed when template environment is empty
func TestTaskRunner_populateTaskEnvironment_EmptyTemplateEnv(t *testing.T) {
	// This scenario replicates what happens with integration extractors:
	// - Task.Environment has extracted variables from webhook
	// - Template environment is empty or doesn't exist
	tsk := TaskRunner{
		Task: db.Task{
			// Integration-extracted variables in Task.Environment
			Environment: `{"etcd_api_url":"http://x.x.x.x:2379","etcd_target_leader_name":"xxxx"}`,
		},
		Environment: db.Environment{
			// Empty template environment (typical case when no environment is configured)
			JSON: "",
		},
	}

	err := tsk.populateTaskEnvironment()
	assert.NoError(t, err)

	// The extracted variables should now be in Environment.JSON
	var env map[string]any
	err = json.Unmarshal([]byte(tsk.Environment.JSON), &env)
	assert.NoError(t, err)

	// Verify extracted values are present and not empty
	assert.Equal(t, "http://x.x.x.x:2379", env["etcd_api_url"])
	assert.Equal(t, "xxxx", env["etcd_target_leader_name"])
}

// TestTaskRunner_populateTaskEnvironment_MergeWithTemplateEnv tests that
// task environment correctly merges with non-empty template environment
func TestTaskRunner_populateTaskEnvironment_MergeWithTemplateEnv(t *testing.T) {
	tsk := TaskRunner{
		Task: db.Task{
			// Integration-extracted variables
			Environment: `{"webhook_var":"from_webhook"}`,
		},
		Environment: db.Environment{
			// Template has its own environment
			JSON: `{"template_var":"from_template","shared_var":"template_value"}`,
		},
	}

	err := tsk.populateTaskEnvironment()
	assert.NoError(t, err)

	var env map[string]any
	err = json.Unmarshal([]byte(tsk.Environment.JSON), &env)
	assert.NoError(t, err)

	// Both template and task variables should be present
	assert.Equal(t, "from_template", env["template_var"])
	assert.Equal(t, "from_webhook", env["webhook_var"])
}

// TestTaskRunner_populateTaskEnvironment_TaskOverridesTemplate tests that
// task environment takes precedence over template environment for same keys
func TestTaskRunner_populateTaskEnvironment_TaskOverridesTemplate(t *testing.T) {
	tsk := TaskRunner{
		Task: db.Task{
			Environment: `{"shared_var":"task_value"}`,
		},
		Environment: db.Environment{
			JSON: `{"shared_var":"template_value","other_var":"keep"}`,
		},
	}

	err := tsk.populateTaskEnvironment()
	assert.NoError(t, err)

	var env map[string]any
	err = json.Unmarshal([]byte(tsk.Environment.JSON), &env)
	assert.NoError(t, err)

	// Task value should override template value
	assert.Equal(t, "task_value", env["shared_var"])
	assert.Equal(t, "keep", env["other_var"])
}
