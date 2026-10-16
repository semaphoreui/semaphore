package util

import (
	"encoding/json"
	"os"
	"testing"
)

// TestAPIIntegration simulates how the API uses the Version() function
func TestAPIIntegration_Version(t *testing.T) {
	// Test case 1: Default behavior (no environment variable)
	os.Unsetenv("SEMAPHORE_BUILD_INFO")
	
	// Set some known values
	originalVer := Ver
	originalCommit := Commit
	originalDate := Date
	
	Ver = "v2.14.9"
	Commit = "66611eb"
	Date = "1746447063"
	
	defer func() {
		Ver = originalVer
		Commit = originalCommit
		Date = originalDate
	}()
	
	// Simulate what the API does
	systemInfo := map[string]interface{}{
		"version": Version(),
		"ansible": "some-ansible-version",
		"web_host": "localhost",
	}
	
	expected := "v2.14.9-66611eb-1746447063"
	if systemInfo["version"] != expected {
		t.Errorf("Expected version '%s', but got '%s'", expected, systemInfo["version"])
	}
	
	// Test case 2: With environment variable override
	customVersion := "v2.14.9-downstream-custom"
	err := os.Setenv("SEMAPHORE_BUILD_INFO", customVersion)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("SEMAPHORE_BUILD_INFO")
	
	// Simulate API call with environment variable set
	systemInfo2 := map[string]interface{}{
		"version": Version(),
		"ansible": "some-ansible-version", 
		"web_host": "localhost",
	}
	
	if systemInfo2["version"] != customVersion {
		t.Errorf("Expected version '%s', but got '%s'", customVersion, systemInfo2["version"])
	}
	
	// Test that it can be JSON serialized (as the API does)
	jsonData, err := json.Marshal(systemInfo2)
	if err != nil {
		t.Errorf("Failed to marshal system info to JSON: %v", err)
	}
	
	// Verify the JSON contains our custom version
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}
	
	if unmarshaled["version"] != customVersion {
		t.Errorf("JSON version mismatch. Expected '%s', got '%s'", customVersion, unmarshaled["version"])
	}
}