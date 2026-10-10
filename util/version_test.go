package util

import (
	"os"
	"testing"
)

func TestVersion_Default(t *testing.T) {
	// Ensure no environment variable is set
	os.Unsetenv("SEMAPHORE_BUILD_INFO")
	
	// Set some test values for the build variables
	originalVer := Ver
	originalCommit := Commit
	originalDate := Date
	
	Ver = "v1.0.0"
	Commit = "abc123"
	Date = "1234567890"
	
	defer func() {
		Ver = originalVer
		Commit = originalCommit
		Date = originalDate
	}()
	
	expected := "v1.0.0-abc123-1234567890"
	actual := Version()
	
	if actual != expected {
		t.Errorf("Expected version '%s', but got '%s'", expected, actual)
	}
}

func TestVersion_EnvironmentOverride(t *testing.T) {
	// Set environment variable
	customVersion := "MyCustomVersion-2.0.0-beta"
	err := os.Setenv("SEMAPHORE_BUILD_INFO", customVersion)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("SEMAPHORE_BUILD_INFO")
	
	actual := Version()
	
	if actual != customVersion {
		t.Errorf("Expected version '%s', but got '%s'", customVersion, actual)
	}
}

func TestVersion_EmptyEnvironmentFallsBackToDefault(t *testing.T) {
	// Set empty environment variable
	err := os.Setenv("SEMAPHORE_BUILD_INFO", "")
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("SEMAPHORE_BUILD_INFO")
	
	// Set some test values for the build variables
	originalVer := Ver
	originalCommit := Commit
	originalDate := Date
	
	Ver = "v1.2.3"
	Commit = "def456"
	Date = "9876543210"
	
	defer func() {
		Ver = originalVer
		Commit = originalCommit
		Date = originalDate
	}()
	
	expected := "v1.2.3-def456-9876543210"
	actual := Version()
	
	if actual != expected {
		t.Errorf("Expected version '%s', but got '%s'", expected, actual)
	}
}