package db_lib

import (
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnsiblePlaybookResolveWorkingDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	repository := db.Repository{GitURL: repoRoot}

	t.Run("repository root by default", func(t *testing.T) {
		playbook := AnsiblePlaybook{Repository: repository}

		workingDirectory, err := playbook.resolveWorkingDirectory()

		require.NoError(t, err)
		assert.Equal(t, repoRoot, workingDirectory)
	})

	t.Run("configured repository subdirectory", func(t *testing.T) {
		relativePath := "deploy/ansible"
		expectedPath := filepath.Join(repoRoot, "deploy", "ansible")
		playbook := AnsiblePlaybook{
			Repository:       repository,
			WorkingDirectory: &relativePath,
		}

		workingDirectory, err := playbook.resolveWorkingDirectory()

		require.NoError(t, err)
		assert.Equal(t, expectedPath, workingDirectory)
	})

	// Backslash is literal on POSIX and a separator on Windows.
	t.Run("uses runner-native separators", func(t *testing.T) {
		relativePath := "deploy\\windows"
		expectedPath := filepath.Join(repoRoot, `deploy\windows`)
		if filepath.Separator == '\\' {
			expectedPath = filepath.Join(repoRoot, "deploy", "windows")
		}
		playbook := AnsiblePlaybook{
			Repository:       repository,
			WorkingDirectory: &relativePath,
		}

		workingDirectory, err := playbook.resolveWorkingDirectory()

		require.NoError(t, err)
		assert.Equal(t, expectedPath, workingDirectory)
	})

	t.Run("repository escape", func(t *testing.T) {
		relativePath := "../outside"
		playbook := AnsiblePlaybook{
			Repository:       repository,
			WorkingDirectory: &relativePath,
		}

		_, err := playbook.resolveWorkingDirectory()

		assert.Error(t, err)
	})
}
