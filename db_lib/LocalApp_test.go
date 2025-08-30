package db_lib

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, item) {
			return true
		}
	}
	return false
}

func TestGetEnvironmentVars(t *testing.T) {
	os.Setenv("SEMAPHORE_TEST", "test123")  //nolint:errcheck
	os.Setenv("SEMAPHORE_TEST2", "test222") //nolint:errcheck
	os.Setenv("PASSWORD", "test222")        //nolint:errcheck

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{"SEMAPHORE_TEST"},
		EnvVars: map[string]string{
			"ANSIBLE_FORCE_COLOR": "False",
		},
	}

	res := getEnvironmentVars()

	expected := []string{
		"SEMAPHORE_TEST=test123",
		"ANSIBLE_FORCE_COLOR=False",
		"PATH=",
	}

	if len(res) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, res)
	}

	for _, e := range expected {
		if !contains(res, e) {
			t.Errorf("Expected %v, got %v", expected, res)
		}
	}
}

func TestAnsiblePlaybookTemplateSpecificHome(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
		Process: &util.ConfigProcess{}, // Empty process config to avoid nil pointer
	}

	// Create two different playbooks with different template IDs
	playbook1 := AnsiblePlaybook{
		TemplateID: 123,
		Repository: db.Repository{
			ProjectID: 42,
		},
		Logger: &mockLogger{},
	}

	playbook2 := AnsiblePlaybook{
		TemplateID: 456,
		Repository: db.Repository{
			ProjectID: 42, // Same project but different template
		},
		Logger: &mockLogger{},
	}

	// Test that both playbooks get different HOME directories
	cmd1 := playbook1.makeCmd("test-command", []string{}, nil)
	cmd2 := playbook2.makeCmd("test-command", []string{}, nil)

	// Extract HOME environment variables
	var home1, home2 string
	for _, env := range cmd1.Env {
		if strings.HasPrefix(env, "HOME=") {
			home1 = env[5:] // Remove "HOME=" prefix
			break
		}
	}
	for _, env := range cmd2.Env {
		if strings.HasPrefix(env, "HOME=") {
			home2 = env[5:] // Remove "HOME=" prefix
			break
		}
	}

	// Verify HOME directories are different
	if home1 == home2 {
		t.Errorf("Expected different HOME directories for different templates, but got same: %s", home1)
	}

	// Verify HOME directories are template-specific
	expectedHome1 := "/tmp/project_42/template_123"
	expectedHome2 := "/tmp/project_42/template_456"

	if home1 != expectedHome1 {
		t.Errorf("Expected HOME for template 123 to be %s, got %s", expectedHome1, home1)
	}

	if home2 != expectedHome2 {
		t.Errorf("Expected HOME for template 456 to be %s, got %s", expectedHome2, home2)
	}
}

// mockLogger implements task_logger.Logger for testing
type mockLogger struct{}

func (l *mockLogger) Log(msg string)                                      {}
func (l *mockLogger) Logf(format string, a ...any)                        {}
func (l *mockLogger) LogWithTime(now time.Time, msg string)               {}
func (l *mockLogger) LogfWithTime(now time.Time, format string, a ...any) {}
func (l *mockLogger) LogCmd(cmd *exec.Cmd)                                 {}
func (l *mockLogger) SetStatus(status task_logger.TaskStatus)             {}
func (l *mockLogger) AddStatusListener(l2 task_logger.StatusListener)     {}
func (l *mockLogger) AddLogListener(l2 task_logger.LogListener)           {}
func (l *mockLogger) SetCommit(hash, message string)                      {}
func (l *mockLogger) WaitLog()                                            {}
