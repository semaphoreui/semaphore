package db

import (
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func TestTemplate_Validate_WorkingDirectory(t *testing.T) {
	originalConfig := util.Config
	t.Cleanup(func() {
		util.Config = originalConfig
	})
	util.Config = &util.ConfigType{
		Apps: map[string]util.App{
			"ansible": {},
			"bash":    {},
		},
	}

	inventoryID := 1
	tests := []struct {
		name             string
		app              TemplateApp
		workingDirectory string
		valid            bool
	}{
		{
			name:             "Ansible repository subdirectory",
			app:              AppAnsible,
			workingDirectory: "deploy/ansible",
			valid:            true,
		},
		{
			name:             "absolute path",
			app:              AppAnsible,
			workingDirectory: "/opt/deployment",
		},
		{
			name:             "repository escape",
			app:              AppAnsible,
			workingDirectory: "../deployment",
		},
		{
			name:             "Windows absolute path",
			app:              AppAnsible,
			workingDirectory: "C:\\deployment",
		},
		{
			name:             "Windows colon outside drive prefix",
			app:              AppAnsible,
			workingDirectory: "directory/file:stream",
		},
		{
			name:             "Windows repository escape",
			app:              AppAnsible,
			workingDirectory: "..\\deployment",
		},
		{
			name:             "Windows UNC path",
			app:              AppAnsible,
			workingDirectory: "\\\\server\\share",
		},
		{
			name:             "POSIX mixed-separator escape",
			app:              AppAnsible,
			workingDirectory: "x\\y\\z/../../script.sh",
		},
		{
			name:             "empty path",
			app:              AppAnsible,
			workingDirectory: "",
		},
		{
			name:             "non-Ansible template",
			app:              AppBash,
			workingDirectory: "deploy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := Template{
				Name:             "Deploy",
				Playbook:         "playbooks/site.yml",
				App:              tt.app,
				InventoryID:      &inventoryID,
				WorkingDirectory: &tt.workingDirectory,
			}

			err := tpl.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
