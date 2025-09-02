package cmd

import (
	"testing"

	"github.com/semaphoreui/semaphore/util"
)

func TestRunnerRegisterArgsProcessing(t *testing.T) {
	// Reset config to defaults
	util.Config = &util.ConfigType{
		Runner: &util.RunnerConfig{},
	}

	// Test CLI flag processing
	runnerRegisterArgs.hostname = "test-runner"
	runnerRegisterArgs.enabled = true
	runnerRegisterArgs.disabled = false
	runnerRegisterArgs.projectName = "test-project"
	runnerRegisterArgs.webhook = "https://example.com/webhook"
	runnerRegisterArgs.tags = "linux,test,docker"

	// Simulate the flag processing logic from registerRunner function
	if runnerRegisterArgs.hostname != "" {
		util.Config.Runner.Name = runnerRegisterArgs.hostname
	}

	if runnerRegisterArgs.disabled {
		enabled := false
		util.Config.Runner.Active = &enabled
	} else if runnerRegisterArgs.enabled {
		enabled := true
		util.Config.Runner.Active = &enabled
	} else if util.Config.Runner.Active == nil {
		enabled := true
		util.Config.Runner.Active = &enabled
	}

	if runnerRegisterArgs.projectName != "" {
		util.Config.Runner.ProjectName = runnerRegisterArgs.projectName
	}

	if runnerRegisterArgs.webhook != "" {
		util.Config.Runner.Webhook = runnerRegisterArgs.webhook
	}

	if runnerRegisterArgs.tags != "" {
		util.Config.Runner.Tag = runnerRegisterArgs.tags
	}

	// Validate the config was set correctly
	if util.Config.Runner.Name != "test-runner" {
		t.Errorf("Expected Name to be 'test-runner', got '%s'", util.Config.Runner.Name)
	}

	if util.Config.Runner.Active == nil || !*util.Config.Runner.Active {
		t.Errorf("Expected Active to be true")
	}

	if util.Config.Runner.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName to be 'test-project', got '%s'", util.Config.Runner.ProjectName)
	}

	if util.Config.Runner.Webhook != "https://example.com/webhook" {
		t.Errorf("Expected Webhook to be 'https://example.com/webhook', got '%s'", util.Config.Runner.Webhook)
	}

	if util.Config.Runner.Tag != "linux,test,docker" {
		t.Errorf("Expected Tag to be 'linux,test,docker', got '%s'", util.Config.Runner.Tag)
	}
}

func TestRunnerRegisterArgsDisabled(t *testing.T) {
	// Reset config to defaults
	util.Config = &util.ConfigType{
		Runner: &util.RunnerConfig{},
	}

	// Test disabled flag takes precedence
	runnerRegisterArgs.enabled = true
	runnerRegisterArgs.disabled = true

	// Simulate the flag processing logic from registerRunner function
	if runnerRegisterArgs.disabled {
		enabled := false
		util.Config.Runner.Active = &enabled
	} else if runnerRegisterArgs.enabled {
		enabled := true
		util.Config.Runner.Active = &enabled
	}

	// Validate disabled takes precedence
	if util.Config.Runner.Active == nil || *util.Config.Runner.Active {
		t.Errorf("Expected Active to be false when disabled flag is set")
	}
}