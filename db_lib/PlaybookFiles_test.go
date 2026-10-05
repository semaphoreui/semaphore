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

func TestFindPlaybooks(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, root string)
		expected []string
	}{
		{
			name: "nested directories",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "site.yml")
				writeFile(t, root, "playbooks/deploy.yml")
				writeFile(t, root, "playbooks/nested/inner.yaml")
			},
			expected: []string{"playbooks/deploy.yml", "playbooks/nested/inner.yaml", "site.yml"},
		},
		{
			name: "excluded directories are skipped",
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
			setup:    func(t *testing.T, root string) {},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			result, err := FindPlaybooks(root, nil)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindPlaybooks_PerApp(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{
		"site.yml",
		"main.tf",
		"envs/prod/main.tf",
		"envs/prod/variables.tf",
		"envs/dev/main.tofu",
		"envs/dev/.terraform/modules/vpc/main.tf",
		"live/eu/terragrunt.hcl",
		"live/root.hcl",
		"scripts/deploy.sh",
		"scripts/build.bash",
		"scripts/tool.py",
		"scripts/__pycache__/tool.py",
		"scripts/setup.PS1",
		"node_modules/pkg/install.sh",
		".git/hooks/pre-commit.sh",
		".github/workflows/ci.yml",
		"envs/dev/.venv/bin/activate.py",
	} {
		writeFile(t, root, f)
	}

	tests := []struct {
		app      db.TemplateApp
		expected []string
	}{
		{"", []string{"site.yml"}},
		{db.AppTerraform, []string{"envs/dev", "envs/prod"}},
		{db.AppTofu, []string{"envs/dev", "envs/prod"}},
		{db.AppTerragrunt, []string{"live/eu"}},
		{db.AppBash, []string{"scripts/build.bash", "scripts/deploy.sh"}},
		{db.AppPython, []string{"scripts/tool.py"}},
		{db.AppPowerShell, []string{"scripts/setup.PS1"}},
		{db.AppPulumi, []string{
			"envs/dev/main.tofu",
			"envs/prod/main.tf",
			"envs/prod/variables.tf",
			"live/eu/terragrunt.hcl",
			"live/root.hcl",
			"main.tf",
			"scripts/build.bash",
			"scripts/deploy.sh",
			"scripts/setup.PS1",
			"scripts/tool.py",
			"site.yml",
		}},
	}

	for _, tt := range tests {
		t.Run(string(tt.app), func(t *testing.T) {
			result, err := FindPlaybooks(root, &tt.app)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindPlaybooks_DoesNotFollowSymlinks(t *testing.T) {
	outside := t.TempDir()
	writeFile(t, outside, "secret.yml")

	root := t.TempDir()
	writeFile(t, root, "site.yml")
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "linked")))

	result, err := FindPlaybooks(root, nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"site.yml"}, result)
}

func TestFindPlaybooks_Capped(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < maxPlaybookFiles+10; i++ {
		writeFile(t, root, filepath.Join("p", fmt.Sprintf("%04d.yml", i)))
	}

	result, err := FindPlaybooks(root, nil)

	require.NoError(t, err)
	assert.Len(t, result, maxPlaybookFiles)
}
