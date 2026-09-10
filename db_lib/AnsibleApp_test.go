package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnsibleApp_skipGalaxyInstall(t *testing.T) {
	tests := []struct {
		name     string
		tpl      *db.AnsibleTemplateParams
		params   *db.AnsibleTaskParams
		expected bool
	}{
		{
			name:     "no template params",
			tpl:      nil,
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: true},
			expected: false,
		},
		{
			name:     "template skip enabled, override disabled",
			tpl:      &db.AnsibleTemplateParams{SkipGalaxyInstall: true},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: false},
			expected: true,
		},
		{
			name:     "template skip disabled, override disabled, task wants skip",
			tpl:      &db.AnsibleTemplateParams{SkipGalaxyInstall: false},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: true},
			expected: false,
		},
		{
			name: "override enabled, task disables skip",
			tpl: &db.AnsibleTemplateParams{
				SkipGalaxyInstall:              true,
				AllowOverrideSkipGalaxyInstall: true,
			},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: false},
			expected: false,
		},
		{
			name: "override enabled, task enables skip",
			tpl: &db.AnsibleTemplateParams{
				SkipGalaxyInstall:              false,
				AllowOverrideSkipGalaxyInstall: true,
			},
			params:   &db.AnsibleTaskParams{SkipGalaxyInstall: true},
			expected: true,
		},
		{
			name: "override enabled, nil task params falls back to template",
			tpl: &db.AnsibleTemplateParams{
				SkipGalaxyInstall:              true,
				AllowOverrideSkipGalaxyInstall: true,
			},
			params:   nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &AnsibleApp{}

			args := LocalAppInstallingArgs{}
			if tt.tpl != nil {
				args.TplParams = tt.tpl
			}
			if tt.params != nil {
				args.Params = tt.params
			}

			assert.Equal(t, tt.expected, app.skipGalaxyInstall(args))
		})
	}
}

func TestAnsibleApp_forceGalaxyInstall(t *testing.T) {
	tests := []struct {
		name     string
		tpl      *db.AnsibleTemplateParams
		params   *db.AnsibleTaskParams
		expected bool
	}{
		{
			name:     "no template params",
			tpl:      nil,
			params:   &db.AnsibleTaskParams{ForceGalaxyInstall: true},
			expected: false,
		},
		{
			name:     "template force enabled, override disabled",
			tpl:      &db.AnsibleTemplateParams{ForceGalaxyInstall: true},
			params:   &db.AnsibleTaskParams{ForceGalaxyInstall: false},
			expected: true,
		},
		{
			name:     "template force disabled, override disabled, task wants force",
			tpl:      &db.AnsibleTemplateParams{ForceGalaxyInstall: false},
			params:   &db.AnsibleTaskParams{ForceGalaxyInstall: true},
			expected: false,
		},
		{
			name: "override enabled, task disables force",
			tpl: &db.AnsibleTemplateParams{
				ForceGalaxyInstall:              true,
				AllowOverrideForceGalaxyInstall: true,
			},
			params:   &db.AnsibleTaskParams{ForceGalaxyInstall: false},
			expected: false,
		},
		{
			name: "override enabled, task enables force",
			tpl: &db.AnsibleTemplateParams{
				ForceGalaxyInstall:              false,
				AllowOverrideForceGalaxyInstall: true,
			},
			params:   &db.AnsibleTaskParams{ForceGalaxyInstall: true},
			expected: true,
		},
		{
			name: "override enabled, nil task params falls back to template",
			tpl: &db.AnsibleTemplateParams{
				ForceGalaxyInstall:              true,
				AllowOverrideForceGalaxyInstall: true,
			},
			params:   nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &AnsibleApp{}

			args := LocalAppInstallingArgs{}
			if tt.tpl != nil {
				args.TplParams = tt.tpl
			}
			if tt.params != nil {
				args.Params = tt.params
			}

			assert.Equal(t, tt.expected, app.forceGalaxyInstall(args))
		})
	}
}

func TestAnsibleApp_installGalaxyRequirementsFile_ForceBypassesCache(t *testing.T) {
	previousConfig := util.Config
	util.Config = &util.ConfigType{
		TmpPath: t.TempDir(),
		Process: &util.ConfigProcess{},
	}
	t.Cleanup(func() { util.Config = previousConfig })

	repoRoot := t.TempDir()
	requirementsFilePath := filepath.Join(repoRoot, "requirements.yml")
	require.NoError(t, os.WriteFile(requirementsFilePath, []byte("roles: []\n"), 0o644))

	capturedArgsPath := filepath.Join(t.TempDir(), "ansible-galaxy.args")
	binDir := t.TempDir()
	scriptPath := filepath.Join(binDir, "ansible-galaxy")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > "+capturedArgsPath+"\n"), 0o755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	app := &AnsibleApp{
		Logger: task_logger.NopLogger{},
		Template: db.Template{
			ID: 1,
		},
		Repository: db.Repository{
			ID:        1,
			ProjectID: 1,
			GitURL:    repoRoot,
			GitBranch: "main",
		},
		Playbook: &AnsiblePlaybook{
			TemplateID: 1,
			Repository: db.Repository{
				ID:        1,
				ProjectID: 1,
				GitURL:    repoRoot,
				GitBranch: "main",
			},
			Logger: task_logger.NopLogger{},
		},
	}

	requirementsHashFilePath := app.requirementsHashFilePath(GalaxyRole, requirementsFilePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(requirementsHashFilePath), 0o755))
	require.NoError(t, writeMD5Hash(requirementsFilePath, requirementsHashFilePath))

	tests := []struct {
		name      string
		force     bool
		expectRun bool
	}{
		{
			name:      "cached requirements are skipped when force is false",
			force:     false,
			expectRun: false,
		},
		{
			name:      "cached requirements are installed when force is true",
			force:     true,
			expectRun: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.RemoveAll(capturedArgsPath))

			err := app.installGalaxyRequirementsFile(GalaxyRole, requirementsFilePath, nil, nil, tt.force)
			require.NoError(t, err)

			capturedArgs, readErr := os.ReadFile(capturedArgsPath)
			if tt.expectRun {
				require.NoError(t, readErr)
				assert.Equal(t, "role\ninstall\n-r\n"+requirementsFilePath+"\n--force\n", string(capturedArgs))
				return
			}

			assert.ErrorIs(t, readErr, os.ErrNotExist)
		})
	}
}
