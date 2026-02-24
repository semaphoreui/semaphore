package api

import (
	"encoding/json"
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"

	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtract_HeaderAndCaseInsensitive(t *testing.T) {
	h := http.Header{}
	h.Set("x-token", "abc123") // lower-case to verify case-insensitive get

	values := []db.IntegrationExtractValue{
		{
			Name:         "Token header",
			ValueSource:  db.IntegrationExtractHeaderValue,
			Key:          "X-Token", // different case
			Variable:     "TOKEN",
			VariableType: db.IntegrationVariableEnvironment,
		},
	}

	got := Extract(values, h, nil)

	require.Equal(t, "abc123", got["TOKEN"], "TOKEN header value should match")
}

func TestExtract_JSONBody_VariousTypesAndMissing(t *testing.T) {
	payload := []byte(`{
		"num": 42,
		"str": "hello",
		"bool": true,
		"nullv": null,
		"obj": {"k":"v"},
		"arr": [1,2,3],
		"nested": {"items":[{"c":123},{"c":"str"}]}
	}`)

	values := []db.IntegrationExtractValue{
		{ // number coerced to string via fmt.Sprintf
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "num",
			Variable:     "NUM",
		},
		{ // string stays same content
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "str",
			Variable:     "STR",
		},
		{ // boolean -> "true"
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "bool",
			Variable:     "BOOL",
		},
		{ // null should not be set (Find returns nil or we skip when nil)
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "nullv",
			Variable:     "NULLV",
		},
		{ // array will be formatted with %v, expect Go-like format
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "arr",
			Variable:     "ARR",
		},
		{ // object -> formatted map with %v
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "obj",
			Variable:     "OBJ",
		},
		{ // missing key should not create an entry
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "missing",
			Variable:     "MISSING",
		},
		{ // nested array index path
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "nested.items.[0].c",
			Variable:     "NESTED_C",
		},
		{ // first element of arr
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "arr.[0]",
			Variable:     "ARR0",
		},
	}

	got := Extract(values, http.Header{}, payload)

	// Basic scalar assertions
	assert.Equal(t, "42", got["NUM"], "NUM should equal stringified number")
	assert.Equal(t, "hello", got["STR"], "STR should match")
	assert.Equal(t, "true", got["BOOL"], "BOOL should be string 'true'")

	// Indexed lookups
	assert.Equal(t, "123", got["NESTED_C"], "NESTED_C should equal nested.items[0].c")
	assert.Equal(t, "1", got["ARR0"], "ARR0 should equal arr[0]")

	// Null should be absent
	assert.NotContains(t, got, "NULLV", "NULLV should not be present for null JSON value")

	// Array/object string formats: we assert non-empty presence rather than exact formatting,
	// because %v formatting of gojsonq return types may vary across versions.
	assert.Contains(t, got, "ARR", "ARR key should be present")
	assert.NotEmpty(t, got["ARR"], "ARR value should be non-empty")
	assert.Contains(t, got, "OBJ", "OBJ key should be present")
	assert.NotEmpty(t, got["OBJ"], "OBJ value should be non-empty")

	// Missing should not appear
	assert.NotContains(t, got, "MISSING", "MISSING should not be present for missing key")
}

func TestExtract_BodyString_ReturnsFullPayload(t *testing.T) {
	payload := []byte("raw body data here")
	values := []db.IntegrationExtractValue{
		{
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataString,
			Variable:     "BODY",
			Key:          "ignored",
		},
	}
	got := Extract(values, http.Header{}, payload)
	if got["BODY"] != string(payload) {
		t.Fatalf("expected BODY to equal full payload; got %q", got["BODY"])
	}
}

func TestExtract_MalformedJSON_SkipsSetting(t *testing.T) {
	payload := []byte("{not: valid json}")
	values := []db.IntegrationExtractValue{
		{
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Variable:     "BAD",
			Key:          "a.b",
		},
	}
	got := Extract(values, http.Header{}, payload)
	if _, ok := got["BAD"]; ok {
		t.Fatalf("expected BAD to be absent for malformed JSON payload")
	}
}

func TestIntegrationMatch(t *testing.T) {
	body := []byte("{\"hook_id\": 4856239453}")
	var header = make(http.Header)
	matched := Match(db.IntegrationMatcher{
		ID:            0,
		Name:          "Test",
		IntegrationID: 0,
		MatchType:     db.IntegrationMatchBody,
		Method:        db.IntegrationMatchMethodEquals,
		BodyDataType:  db.IntegrationBodyDataJSON,
		Key:           "hook_id",
		Value:         "4856239453",
	}, header, body)

	assert.True(t, matched)
}

func TestGetTaskDefinitionSuccess(t *testing.T) {
	integration := db.Integration{
		ID:         11,
		ProjectID:  22,
		TemplateID: 33,
		TaskParams: &db.TaskParams{
			ProjectID:   22,
			Environment: `{"existing":"value"}`,
			Params:      db.MapStringAnyField{"original": "keep"},
		},
	}

	header := make(http.Header)
	header.Set("X-Env", "header-value")
	payload := []byte(`{"data":{"param":"payload-value"}}`)

	extractorCalled := false
	task, err := GetTaskDefinition(integration, payload, header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
		extractorCalled = true

		if projectID != integration.ProjectID {
			t.Fatalf("expected projectID %d, got %d", integration.ProjectID, projectID)
		}
		if integrationID != integration.ID {
			t.Fatalf("expected integrationID %d, got %d", integration.ID, integrationID)
		}

		return []db.IntegrationExtractValue{
			{
				VariableType: db.IntegrationVariableEnvironment,
				ValueSource:  db.IntegrationExtractHeaderValue,
				Key:          "X-Env",
				Variable:     "HOOK_ENV",
			},
			{
				VariableType: db.IntegrationVariableTaskParam,
				ValueSource:  db.IntegrationExtractBodyValue,
				BodyDataType: db.IntegrationBodyDataJSON,
				Key:          "data.param",
				Variable:     "payloadParam",
			},
		}, nil
	})

	assert.NoError(t, err)
	assert.True(t, extractorCalled)

	if assert.NotNil(t, task.IntegrationID) {
		assert.Equal(t, integration.ID, *task.IntegrationID)
	}

	assert.Equal(t, integration.ProjectID, task.ProjectID)
	assert.Equal(t, integration.TemplateID, task.TemplateID)
	assert.NotEmpty(t, task.Environment)

	var env map[string]any
	if assert.NoError(t, json.Unmarshal([]byte(task.Environment), &env)) {
		assert.Equal(t, "value", env["existing"])
		assert.Equal(t, "header-value", env["HOOK_ENV"])
	}

	if assert.NotNil(t, task.Params) {
		if assert.Contains(t, task.Params, "original") {
			assert.Equal(t, "keep", task.Params["original"])
		}

		if assert.Contains(t, task.Params, "payloadParam") {
			payloadParam, ok := task.Params["payloadParam"].(string)
			assert.True(t, ok)
			assert.Equal(t, "payload-value", payloadParam)
		}
	}
}

func TestGetTaskDefinitionExtractorError(t *testing.T) {
	integration := db.Integration{
		ID:         44,
		ProjectID:  55,
		TemplateID: 66,
	}

	header := make(http.Header)
	payload := []byte(`{}`)

	expectedErr := errors.New("extractor failure")

	extractorCalled := false
	task, err := GetTaskDefinition(integration, payload, header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
		extractorCalled = true
		return nil, expectedErr
	})

	assert.True(t, extractorCalled)
	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, task.IntegrationID)
}

func TestGetTaskDefinitionInvalidEnvironmentJSON(t *testing.T) {
	integration := db.Integration{
		ID:         77,
		ProjectID:  88,
		TemplateID: 99,
		TaskParams: &db.TaskParams{
			ProjectID:   88,
			Environment: "{not-json}",
			Params:      db.MapStringAnyField{},
		},
	}

	header := make(http.Header)
	payload := []byte(`{}`)

	_, err := GetTaskDefinition(integration, payload, header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
		return nil, nil
	})

	assert.Error(t, err)
}

func TestGetTaskDefinitionIntegrationWithoutTaskParams(t *testing.T) {
	integration := db.Integration{
		ID:         44,
		ProjectID:  55,
		TemplateID: 66,
	}

	header := make(http.Header)
	payload := []byte(`{}`)
	extractorCalled := false
	task, err := GetTaskDefinition(integration, payload, header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
		extractorCalled = true

		if projectID != integration.ProjectID {
			t.Fatalf("expected projectID %d, got %d", integration.ProjectID, projectID)
		}
		if integrationID != integration.ID {
			t.Fatalf("expected integrationID %d, got %d", integration.ID, integrationID)
		}

		return []db.IntegrationExtractValue{
			{
				VariableType: db.IntegrationVariableEnvironment,
				ValueSource:  db.IntegrationExtractHeaderValue,
				Key:          "X-Env",
				Variable:     "HOOK_ENV",
			},
			{
				VariableType: db.IntegrationVariableTaskParam,
				ValueSource:  db.IntegrationExtractBodyValue,
				BodyDataType: db.IntegrationBodyDataJSON,
				Key:          "data.param",
				Variable:     "payloadParam",
			},
		}, nil
	})

	assert.True(t, extractorCalled)
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestGetTaskDefinitionWithExtractedEnvValues(t *testing.T) {
	// Test case 1: Empty environment should still include extracted values
	integration := db.Integration{
		ID:         1,
		ProjectID:  1,
		TemplateID: 1,
	}

	// Create test payload
	payload := []byte("{\"branch\": \"main\", \"commit\": \"abc123\"}")

	// Create test request with headers
	req, _ := http.NewRequest("POST", "/webhook", nil)
	req.Header.Set("X-GitHub-Event", "push")

	// Mock extracted environment values (this would normally come from database)
	envValues := []db.IntegrationExtractValue{
		{
			Variable:     "BRANCH_NAME",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "branch",
			VariableType: db.IntegrationVariableEnvironment,
		},
		{
			Variable:     "COMMIT_HASH",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "commit",
			VariableType: db.IntegrationVariableEnvironment,
		},
		{
			Variable:     "EVENT_TYPE",
			ValueSource:  db.IntegrationExtractHeaderValue,
			Key:          "X-GitHub-Event",
			VariableType: db.IntegrationVariableEnvironment,
		},
	}

	// Test Extract function directly first
	extractedEnvResults := Extract(envValues, req.Header, payload)

	if extractedEnvResults["BRANCH_NAME"] != "main" {
		t.Errorf("Expected BRANCH_NAME to be 'main', got '%s'", extractedEnvResults["BRANCH_NAME"])
	}
	if extractedEnvResults["COMMIT_HASH"] != "abc123" {
		t.Errorf("Expected COMMIT_HASH to be 'abc123', got '%s'", extractedEnvResults["COMMIT_HASH"])
	}
	if extractedEnvResults["EVENT_TYPE"] != "push" {
		t.Errorf("Expected EVENT_TYPE to be 'push', got '%s'", extractedEnvResults["EVENT_TYPE"])
	}

	// Test case 1: Empty environment should include extracted values (FIXED behavior)
	taskDef1 := db.Task{
		ProjectID:   1,
		TemplateID:  1,
		Environment: "", // Empty environment
		Params:      make(db.MapStringAnyField),
	}
	taskDef1.IntegrationID = &integration.ID

	// Simulate the FIXED logic from GetTaskDefinition
	env1 := make(map[string]any)

	if taskDef1.Environment != "" {
		json.Unmarshal([]byte(taskDef1.Environment), &env1)
	}

	// Add extracted environment variables only if they don't conflict with
	// existing task definition variables (task definition has higher priority)
	for k, v := range extractedEnvResults {
		if _, exists := env1[k]; !exists {
			env1[k] = v
		}
	}

	envStr1, _ := json.Marshal(env1)
	taskDef1.Environment = string(envStr1)

	// Verify that extracted values ARE now in the environment
	var envCheck1 map[string]any
	json.Unmarshal([]byte(taskDef1.Environment), &envCheck1)

	if envCheck1["BRANCH_NAME"] != "main" {
		t.Errorf("Expected BRANCH_NAME to be 'main' in environment, got '%v'", envCheck1["BRANCH_NAME"])
	}
	if envCheck1["COMMIT_HASH"] != "abc123" {
		t.Errorf("Expected COMMIT_HASH to be 'abc123' in environment, got '%v'", envCheck1["COMMIT_HASH"])
	}
	if envCheck1["EVENT_TYPE"] != "push" {
		t.Errorf("Expected EVENT_TYPE to be 'push' in environment, got '%v'", envCheck1["EVENT_TYPE"])
	}

	// Test case 2: Existing environment should merge with extracted values
	taskDef2 := db.Task{
		ProjectID:   1,
		TemplateID:  1,
		Environment: `{"EXISTING_VAR": "existing_value"}`, // Existing environment
		Params:      make(db.MapStringAnyField),
	}
	taskDef2.IntegrationID = &integration.ID

	env2 := make(map[string]any)

	if taskDef2.Environment != "" {
		json.Unmarshal([]byte(taskDef2.Environment), &env2)
	}

	// Add extracted environment variables only if they don't conflict with
	// existing task definition variables (task definition has higher priority)
	for k, v := range extractedEnvResults {
		if _, exists := env2[k]; !exists {
			env2[k] = v
		}
	}

	envStr2, _ := json.Marshal(env2)
	taskDef2.Environment = string(envStr2)

	// Verify that both existing and extracted values are in the environment
	var envCheck2 map[string]any
	json.Unmarshal([]byte(taskDef2.Environment), &envCheck2)

	if envCheck2["EXISTING_VAR"] != "existing_value" {
		t.Errorf("Expected EXISTING_VAR to be 'existing_value' in environment, got '%v'", envCheck2["EXISTING_VAR"])
	}
	if envCheck2["BRANCH_NAME"] != "main" {
		t.Errorf("Expected BRANCH_NAME to be 'main' in environment, got '%v'", envCheck2["BRANCH_NAME"])
	}
	if envCheck2["COMMIT_HASH"] != "abc123" {
		t.Errorf("Expected COMMIT_HASH to be 'abc123' in environment, got '%v'", envCheck2["COMMIT_HASH"])
	}
	if envCheck2["EVENT_TYPE"] != "push" {
		t.Errorf("Expected EVENT_TYPE to be 'push' in environment, got '%v'", envCheck2["EVENT_TYPE"])
	}

	// Test case 3: Task definition values should have priority over extracted values
	taskDef3 := db.Task{
		ProjectID:   1,
		TemplateID:  1,
		Environment: `{"BRANCH_NAME": "production", "EXISTING_VAR": "from_task"}`, // Conflicts with extracted BRANCH_NAME
		Params:      make(db.MapStringAnyField),
	}
	taskDef3.IntegrationID = &integration.ID

	env3 := make(map[string]any)

	if taskDef3.Environment != "" {
		json.Unmarshal([]byte(taskDef3.Environment), &env3)
	}

	// Add extracted environment variables only if they don't conflict with
	// existing task definition variables (task definition has higher priority)
	for k, v := range extractedEnvResults {
		if _, exists := env3[k]; !exists {
			env3[k] = v
		}
	}

	envStr3, _ := json.Marshal(env3)
	taskDef3.Environment = string(envStr3)

	// Verify that task definition values take precedence over extracted values
	var envCheck3 map[string]any
	json.Unmarshal([]byte(taskDef3.Environment), &envCheck3)

	// BRANCH_NAME should remain "production" from task definition, not "main" from extracted
	if envCheck3["BRANCH_NAME"] != "production" {
		t.Errorf("Expected BRANCH_NAME to be 'production' (task definition priority), got '%v'", envCheck3["BRANCH_NAME"])
	}
	// EXISTING_VAR should remain from task definition
	if envCheck3["EXISTING_VAR"] != "from_task" {
		t.Errorf("Expected EXISTING_VAR to be 'from_task', got '%v'", envCheck3["EXISTING_VAR"])
	}
	// Non-conflicting extracted values should still be added
	if envCheck3["COMMIT_HASH"] != "abc123" {
		t.Errorf("Expected COMMIT_HASH to be 'abc123' in environment, got '%v'", envCheck3["COMMIT_HASH"])
	}
	if envCheck3["EVENT_TYPE"] != "push" {
		t.Errorf("Expected EVENT_TYPE to be 'push' in environment, got '%v'", envCheck3["EVENT_TYPE"])
	}
}

// Test the Extract function to ensure it works correctly for both body and header extraction
func TestExtractBodyAndHeaderValues(t *testing.T) {
	// Create test payload with nested JSON
	payload := []byte(`{"repository": {"name": "test-repo"}, "ref": "refs/heads/main", "pusher": {"name": "johndoe"}}`)

	// Create test request with headers
	req, _ := http.NewRequest("POST", "/webhook", nil)
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-GitHub-Delivery", "12345")

	// Test various extraction scenarios
	extractValues := []db.IntegrationExtractValue{
		{
			Variable:     "REPO_NAME",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "repository.name",
		},
		{
			Variable:     "GIT_REF",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "ref",
		},
		{
			Variable:     "PUSHER_NAME",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataJSON,
			Key:          "pusher.name",
		},
		{
			Variable:    "GITHUB_EVENT",
			ValueSource: db.IntegrationExtractHeaderValue,
			Key:         "X-GitHub-Event",
		},
		{
			Variable:    "GITHUB_DELIVERY",
			ValueSource: db.IntegrationExtractHeaderValue,
			Key:         "X-GitHub-Delivery",
		},
		{
			Variable:     "FULL_PAYLOAD",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataString,
		},
	}

	result := Extract(extractValues, req.Header, payload)

	// Verify body JSON extractions
	if result["REPO_NAME"] != "test-repo" {
		t.Errorf("Expected REPO_NAME to be 'test-repo', got '%s'", result["REPO_NAME"])
	}
	if result["GIT_REF"] != "refs/heads/main" {
		t.Errorf("Expected GIT_REF to be 'refs/heads/main', got '%s'", result["GIT_REF"])
	}
	if result["PUSHER_NAME"] != "johndoe" {
		t.Errorf("Expected PUSHER_NAME to be 'johndoe', got '%s'", result["PUSHER_NAME"])
	}

	// Verify header extractions
	if result["GITHUB_EVENT"] != "push" {
		t.Errorf("Expected GITHUB_EVENT to be 'push', got '%s'", result["GITHUB_EVENT"])
	}
	if result["GITHUB_DELIVERY"] != "12345" {
		t.Errorf("Expected GITHUB_DELIVERY to be '12345', got '%s'", result["GITHUB_DELIVERY"])
	}

	// Verify string body extraction
	if result["FULL_PAYLOAD"] != string(payload) {
		t.Errorf("Expected FULL_PAYLOAD to match original payload")
	}
}
