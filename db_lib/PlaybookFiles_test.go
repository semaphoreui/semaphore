package db_lib

import (
	"fmt"
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
		dir      string
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
		{
			// Terraform and friends run a directory, so the picker offers one.
			name: "terraform lists top level directories only",
			app:  db.AppTerraform,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "main.tf")
				writeFile(t, root, "prod/main.tf")
				writeFile(t, root, "prod/modules/vpc/main.tf")
				writeFile(t, root, "staging/main.tf")
				writeFile(t, root, ".terraform/plugin.json")
				writeFile(t, root, ".git/config")
			},
			expected: []string{".", "prod", "staging"},
		},
		{
			name: "opentofu lists directories",
			app:  db.AppTofu,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "envs/dev/main.tf")
				writeFile(t, root, "main.tf")
			},
			expected: []string{".", "envs"},
		},
		{
			name: "terragrunt lists directories",
			app:  db.AppTerragrunt,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "live/terragrunt.hcl")
			},
			expected: []string{".", "live"},
		},
		{
			name: "pulumi lists directories",
			app:  db.AppPulumi,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "infra/index.ts")
			},
			expected: []string{".", "infra"},
		},
		{
			// A project with main.tf in the root still has somewhere to point.
			name: "a repository of only files offers its root",
			app:  db.AppTerraform,
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "main.tf")
			},
			expected: []string{"."},
		},
		{
			// Typing the separator asks for what is inside that directory.
			name: "terraform lists the directories inside dir",
			app:  db.AppTerraform,
			dir:  "prod/",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "prod/eu/main.tf")
				writeFile(t, root, "prod/us/main.tf")
				writeFile(t, root, "prod/main.tf")
				writeFile(t, root, "prod/eu/modules/vpc/main.tf")
				writeFile(t, root, "staging/main.tf")
			},
			expected: []string{"prod", "prod/eu", "prod/us"},
		},
		{
			name: "a trailing separator is not required",
			app:  db.AppTerraform,
			dir:  "prod",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "prod/eu/main.tf")
			},
			expected: []string{"prod", "prod/eu"},
		},
		{
			// Half a name while typing must not be an error.
			name: "a directory which does not exist yields nothing",
			app:  db.AppTerraform,
			dir:  "pro/",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "prod/eu/main.tf")
			},
			expected: nil,
		},
		{
			// dir comes from the request. Cleaning it against "/" resolves the
			// "..", so climbing out lands on the repository root.
			name: "climbing out of the repository lands on its root",
			app:  db.AppTerraform,
			dir:  "../../../../",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "prod/main.tf")
			},
			expected: []string{".", "prod"},
		},
		{
			// The absolute path of a real directory outside the repository is
			// read relative to it, so it finds nothing.
			name: "an absolute path outside the repository reads nothing",
			app:  db.AppTerraform,
			dir:  "/etc/",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "prod/main.tf")
			},
			expected: nil,
		},
		{
			name: "dir is ignored by the apps which list files",
			app:  db.AppBash,
			dir:  "scripts/",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "deploy.sh")
				writeFile(t, root, "scripts/release.sh")
			},
			expected: []string{"deploy.sh", "scripts/release.sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			result, err := FindRepositoryFiles(root, tt.app, tt.dir)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// The picker is a suggestion list: it stops at maxRepositoryFiles, and anything
// beyond it is still reachable by typing the path.
func TestFindRepositoryFiles_Limit(t *testing.T) {
	t.Run("files", func(t *testing.T) {
		root := t.TempDir()
		for i := 0; i < maxRepositoryFiles+10; i++ {
			writeFile(t, root, fmt.Sprintf("playbook-%03d.yml", i))
		}

		result, err := FindRepositoryFiles(root, db.AppAnsible, "")

		require.NoError(t, err)
		assert.Len(t, result, maxRepositoryFiles)
	})

	t.Run("directories", func(t *testing.T) {
		root := t.TempDir()
		for i := 0; i < maxRepositoryFiles+10; i++ {
			writeFile(t, root, fmt.Sprintf("env-%03d/main.tf", i))
		}

		result, err := FindRepositoryFiles(root, db.AppTerraform, "")

		require.NoError(t, err)
		assert.Len(t, result, maxRepositoryFiles)
	})
}

// A checkout can contain a directory symlink pointing out of it. Cleaning the
// requested path resolves "..", but following such a link would still list a
// directory outside the repository.
func TestFindRepositoryFiles_SymlinkDoesNotEscape(t *testing.T) {
	outside := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(outside, "secret-dir"), 0755))

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "prod"), 0755))

	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skip("symlinks are not available: " + err.Error())
	}

	result, err := FindRepositoryFiles(root, db.AppTerraform, "escape/")

	assert.Error(t, err, "listing through the symlink must be refused")
	assert.Empty(t, result)

	// The repository itself is unaffected, and the link is not offered: ReadDir
	// reports its own type, not the directory it points at.
	result, err = FindRepositoryFiles(root, db.AppTerraform, "")
	require.NoError(t, err)
	assert.Equal(t, []string{".", "prod"}, result)
}

// A .venv holding more matching files than the cap would fill the result before
// the walk reached the project's own, leaving the picker empty of what matters.
func TestFindRepositoryFiles_DotDirsDoNotFillTheLimit(t *testing.T) {
	root := t.TempDir()

	for i := 0; i < maxRepositoryFiles+10; i++ {
		writeFile(t, root, fmt.Sprintf(".venv/lib/mod-%03d.py", i))
	}
	writeFile(t, root, "main.py")

	result, err := FindRepositoryFiles(root, db.AppPython, "")

	require.NoError(t, err)
	assert.Equal(t, []string{"main.py"}, result)
}
