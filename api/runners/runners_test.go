package runners

import (
	"testing"

	"github.com/semaphoreui/semaphore/services/runners"
)

func TestRunnerRegistrationStructure(t *testing.T) {
	// Test that the RunnerRegistration struct has the new fields
	testActive := true
	registration := runners.RunnerRegistration{
		RegistrationToken: "test_token_123",
		Webhook:           "https://example.com/webhook",
		MaxParallelTasks:  2,
		Name:              "test-runner-name",
		Active:            &testActive,
		ProjectName:       "test-project",
		Tag:               "linux,docker,test",
	}

	// Verify all fields are set correctly
	if registration.RegistrationToken != "test_token_123" {
		t.Errorf("Expected RegistrationToken 'test_token_123', got '%s'", registration.RegistrationToken)
	}

	if registration.Webhook != "https://example.com/webhook" {
		t.Errorf("Expected Webhook 'https://example.com/webhook', got '%s'", registration.Webhook)
	}

	if registration.MaxParallelTasks != 2 {
		t.Errorf("Expected MaxParallelTasks 2, got %d", registration.MaxParallelTasks)
	}

	if registration.Name != "test-runner-name" {
		t.Errorf("Expected Name 'test-runner-name', got '%s'", registration.Name)
	}

	if registration.Active == nil || !*registration.Active {
		t.Error("Expected Active to be true")
	}

	if registration.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName 'test-project', got '%s'", registration.ProjectName)
	}

	if registration.Tag != "linux,docker,test" {
		t.Errorf("Expected Tag 'linux,docker,test', got '%s'", registration.Tag)
	}
}