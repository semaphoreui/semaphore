package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePlaybookPath(t *testing.T) {
	tests := []struct {
		name     string
		playbook string
		valid    bool
	}{
		{"empty", "", true},
		{"simple filename", "site.yml", true},
		{"subdirectory", "playbooks/site.yml", true},
		{"dot-slash prefix", "./site.yml", true},
		{"internal dot-dot staying inside", "playbooks/../site.yml", true},
		{"terraform subdirectory", "environments/prod", true},

		{"absolute path", "/etc/cron.d/evil.yml", false},
		{"absolute script", "/usr/bin/script.sh", false},
		{"parent escape", "../outside.yml", false},
		{"deep parent escape", "playbooks/../../outside.yml", false},
		{"hidden parent escape", "./../outside.yml", false},
		{"only dot-dot", "..", false},
		{"windows absolute path", "C:\\Windows\\evil.ps1", false},
		{"windows drive forward slashes", "c:/windows/evil.ps1", false},
		{"windows parent escape", "..\\evil.ps1", false},
		{"windows unc path", "\\\\server\\share\\evil.ps1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlaybookPath(tt.playbook, "template")
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTemplate_Validate_AbsolutePlaybook_ReturnsError(t *testing.T) {
	tpl := &Template{
		Name:     "Deploy",
		Playbook: "/etc/evil.yml",
	}

	err := tpl.Validate()

	assert.EqualError(t, err, "template playbook must be a relative path inside the repository")
}

func TestTemplate_Validate_PlaybookOutsideRepo_ReturnsError(t *testing.T) {
	tpl := &Template{
		Name:     "Deploy",
		Playbook: "../outside.yml",
	}

	err := tpl.Validate()

	assert.EqualError(t, err, "template playbook must not point outside the repository")
}

func TestTemplate_Validate_RelativePlaybook_ReturnsNoError(t *testing.T) {
	tpl := &Template{
		Name:     "Deploy",
		Playbook: "playbooks/site.yml",
	}

	err := tpl.Validate()

	assert.NoError(t, err)
}

func TestTask_ValidateNewTask_AbsolutePlaybook_ReturnsError(t *testing.T) {
	task := &Task{
		Playbook: "/usr/bin/evil.sh",
	}

	err := task.ValidateNewTask(Template{})

	assert.EqualError(t, err, "task playbook must be a relative path inside the repository")
}

func TestTask_ValidateNewTask_RelativePlaybook_ReturnsNoError(t *testing.T) {
	task := &Task{
		Playbook: "site.yml",
	}

	err := task.ValidateNewTask(Template{})

	assert.NoError(t, err)
}
