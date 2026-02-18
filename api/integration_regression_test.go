package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

// TestIntegrationExtractorRegression tests the specific issue reported in #2XXX
// where integration extractor variables from JSON body were not passed correctly
// to Ansible extra-vars in v2.17
func TestIntegrationExtractorRegression(t *testing.T) {
	// Simulate the exact scenario from the issue
	payload := []byte(`{"commonLabels": {"semaphore_integration_name": "etcd/move-leader", "etcd_api_url": "http://x.x.x.x:2379", "etcd_target_leader_name": "xxxx"}}`)

	extractValues := []db.IntegrationExtractValue{
		{
			Variable:     "etcd_api_url",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "commonLabels.etcd_api_url",
			VariableType: db.IntegrationVariableEnvironment,
		},
		{
			Variable:     "etcd_target_leader_name",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "commonLabels.etcd_target_leader_name",
			VariableType: db.IntegrationVariableEnvironment,
		},
	}

	// Test Extract function directly
	extracted := Extract(extractValues, http.Header{}, payload)

	// These should not be empty or null
	assert.Equal(t, "http://x.x.x.x:2379", extracted["etcd_api_url"], "etcd_api_url should be extracted correctly")
	assert.Equal(t, "xxxx", extracted["etcd_target_leader_name"], "etcd_target_leader_name should be extracted correctly")

	// Now test the full GetTaskDefinition flow
	integration := db.Integration{
		ID:         1,
		ProjectID:  1,
		TemplateID: 1,
	}

	taskDef, err := GetTaskDefinition(integration, payload, http.Header{}, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
		return extractValues, nil
	})

	assert.NoError(t, err)

	// Parse the environment to verify values are set correctly
	var env map[string]any
	err = json.Unmarshal([]byte(taskDef.Environment), &env)
	assert.NoError(t, err)

	// Values should be present and non-empty
	assert.Equal(t, "http://x.x.x.x:2379", env["etcd_api_url"], "etcd_api_url should be in environment")
	assert.Equal(t, "xxxx", env["etcd_target_leader_name"], "etcd_target_leader_name should be in environment")

	// Neither should be empty string or null
	assert.NotEqual(t, "", env["etcd_api_url"], "etcd_api_url should not be empty")
	assert.NotEqual(t, nil, env["etcd_api_url"], "etcd_api_url should not be null")
	assert.NotEqual(t, "", env["etcd_target_leader_name"], "etcd_target_leader_name should not be empty")
	assert.NotEqual(t, nil, env["etcd_target_leader_name"], "etcd_target_leader_name should not be null")
}
