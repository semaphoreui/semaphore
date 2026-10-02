package db_lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile creates a file (and its parent directories) with some content.
func writeFile(t *testing.T, root string, relPath string) {
	t.Helper()
	full := filepath.Join(root, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0755))
	require.NoError(t, os.WriteFile(full, []byte("---\n"), 0644))
}

func TestFindRepositoryFiles(t *testing.T) {
	tests := []struct {
		name     string
		app      db.TemplateApp
		setup    func(t *testing.T, root string)
		expected []string
	}{
		{
			name: "nested directories",
			app:  db.AppAnsible,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "site.yml")
				writeFile(t, root, "playbooks/deploy.yml")
				writeFile(t, root, "playbooks/nested/inner.yaml")
			},
			expected: []string{"playbooks/deploy.yml", "playbooks/nested/inner.yaml", "site.yml"},
		},
		{
			name: "excluded directories are skipped",
			app:  db.AppAnsible,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "site.yml")
				writeFile(t, root, "roles/common/tasks/main.yml")
				writeFile(t, root, "group_vars/all.yml")
				writeFile(t, root, "host_vars/host1.yml")
				writeFile(t, root, ".git/config.yml")
				writeFile(t, root, "templates/config.yml")
				writeFile(t, root, "molecule/default/molecule.yml")
			},
			expected: []string{"site.yml"},
		},
		{
			name: "mixed extensions only yml and yaml counted",
			app:  db.AppAnsible,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "site.yml")
				writeFile(t, root, "readme.txt")
				writeFile(t, root, "vars.YAML")
				writeFile(t, root, "script.sh")
				writeFile(t, root, "inventory.ini")
			},
			expected: []string{"site.yml", "vars.YAML"},
		},
		{
			name:     "empty directory",
			app:      db.AppAnsible,
			setup:    func(t *testing.T, root string) {},
			expected: nil,
		},
		{
			// SEM-239: a Bash template was offered playbooks and never its own
			// scripts, and could not be given a path to one by hand.
			name: "bash gets scripts, not playbooks",
			app:  db.AppBash,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "site.yml")
				writeFile(t, root, "requirements.yml")
				writeFile(t, root, "deploy.sh")
				writeFile(t, root, "scripts/release.sh")
				writeFile(t, root, "tool.py")
			},
			expected: []string{"deploy.sh", "scripts/release.sh"},
		},
		{
			// roles/ and tests/ are Ansible conventions, not script ones.
			name: "ansible layout directories are not skipped for a script app",
			app:  db.AppBash,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "roles/deploy.sh")
				writeFile(t, root, "tests/smoke.sh")
				writeFile(t, root, ".git/hook.sh")
			},
			expected: []string{"roles/deploy.sh", "tests/smoke.sh"},
		},
		{
			name: "python",
			app:  db.AppPython,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "main.py")
				writeFile(t, root, "site.yml")
			},
			expected: []string{"main.py"},
		},
		{
			name: "powershell",
			app:  db.AppPowerShell,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "task.ps1")
				writeFile(t, root, "site.yml")
			},
			expected: []string{"task.ps1"},
		},
		{
			// Nothing is known about a user-defined app, so nothing is hidden.
			name: "user defined app lists everything",
			app:  db.TemplateApp("my-custom-runner"),
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "run.custom")
				writeFile(t, root, "site.yml")
				writeFile(t, root, ".git/config")
			},
			expected: []string{"run.custom", "site.yml"},
		},
		{
			// An unset app is what an older client sends; it must behave as before.
			name: "no app behaves as ansible",
			app:  "",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "site.yml")
				writeFile(t, root, "deploy.sh")
				writeFile(t, root, "roles/common/tasks/main.yml")
			},
			expected: []string{"site.yml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			result, err := FindRepositoryFiles(root, tt.app)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
