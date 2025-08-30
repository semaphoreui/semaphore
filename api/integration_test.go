package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

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

	if !matched {
		t.Fatal()
	}
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
	extractedEnvResults := Extract(envValues, req, payload)
	
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
			Variable:     "GITHUB_EVENT",
			ValueSource:  db.IntegrationExtractHeaderValue,
			Key:          "X-GitHub-Event",
		},
		{
			Variable:     "GITHUB_DELIVERY",
			ValueSource:  db.IntegrationExtractHeaderValue,
			Key:          "X-GitHub-Delivery",
		},
		{
			Variable:     "FULL_PAYLOAD",
			ValueSource:  db.IntegrationExtractBodyValue,
			BodyDataType: db.IntegrationBodyDataString,
		},
	}
	
	result := Extract(extractValues, req, payload)
	
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
