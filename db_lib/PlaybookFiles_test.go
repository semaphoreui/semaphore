package db_lib

import (
	"os"
	"path/filepath"
	"testing"

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

			result, err := FindPlaybooks(root)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
