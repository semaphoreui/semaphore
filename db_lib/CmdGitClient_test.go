package db_lib

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepositoryBranchNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "simple branch names",
			input: []string{
				"abc123\trefs/heads/main",
				"def456\trefs/heads/develop",
			},
			expected: []string{"main", "develop"},
		},
		{
			name: "branch names with slashes",
			input: []string{
				"abc123\trefs/heads/env/test",
				"def456\trefs/heads/feature/my-feature",
			},
			expected: []string{"env/test", "feature/my-feature"},
		},
		{
			name: "mixed branch names",
			input: []string{
				"abc123\trefs/heads/main",
				"456def\trefs/heads/release/v1.0",
				"789ghi\trefs/tags/v1.0",
				"jkl012\trefs/heads/feature/awesome",
			},
			expected: []string{"main", "release/v1.0", "feature/awesome"},
		},
		{
			name: "real world git output with various branch patterns",
			input: []string{
				"7f83b1657ff1fc53b92dc18148b1d6d9c2472dc5\trefs/heads/main",
				"a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0\trefs/heads/env/test",
				"b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0a1\trefs/heads/release/v1.0",
				"c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0a1b2\trefs/pull/1/head",
				"d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0a1b2c3\trefs/tags/v1.0.0",
			},
			expected: []string{"main", "env/test", "release/v1.0"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "skip lines without tab",
			input: []string{
				"invalid line",
				"abc123\trefs/heads/main",
			},
			expected: []string{"main"},
		},
		{
			name: "skip non-heads refs",
			input: []string{
				"abc123\trefs/tags/v1.0",
				"def456\trefs/heads/main",
			},
			expected: []string{"main"},
		},
		{
			name: "trailing whitespace in ref path",
			input: []string{
				"abc123\trefs/heads/env/test\n",
			},
			expected: []string{"env/test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRepositoryBranchNames(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d branches, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, branch := range result {
				if branch != tt.expected[i] {
					t.Errorf("branch[%d]: expected %q, got %q", i, tt.expected[i], branch)
				}
			}
		})
	}
}

func TestCmdGitClient_ErrorSanitizationAndStderr(t *testing.T) {
	setupGitClientTest(t)

	client := CreateCmdGitClient(nopKeyInstaller{})
	// Repo with password in URL pointing to unreachable port
	repo := newTestGitRepo(t, "https://gitlab_user:my_secret_token@127.0.0.1:59999/test/repo.git", "main")

	t.Run("GetRemoteBranches_ReturnsDescriptiveSanitizedError", func(t *testing.T) {
		branches, err := client.GetRemoteBranches(repo)
		require.Error(t, err)
		assert.Empty(t, branches)

		var userErr *common_errors.UserVisibleError
		require.ErrorAs(t, err, &userErr, "error should be a UserVisibleError")

		errMsg := err.Error()
		assert.NotContains(t, errMsg, "my_secret_token", "secret token must be redacted")
		assert.NotEqual(t, "exit status 128", errMsg, "error must not be a bare exit-status error")
		assert.True(t,
			strings.Contains(errMsg, "Connection refused: Git server is unreachable") ||
				strings.Contains(errMsg, "Could not resolve host: Unable to connect to Git server") ||
				strings.Contains(errMsg, "unreachable") ||
				strings.Contains(errMsg, "git ls-remote failed"),
			"expected descriptive git failure, got: %s", errMsg,
		)
	})

	t.Run("Clone_ReturnsDescriptiveSanitizedError", func(t *testing.T) {
		repo.TmpDirName = "test_clone_fail"
		err := client.Clone(repo)
		require.Error(t, err)

		var userErr *common_errors.UserVisibleError
		require.ErrorAs(t, err, &userErr, "error should be a UserVisibleError")

		errMsg := err.Error()
		assert.NotContains(t, errMsg, "my_secret_token", "secret token must be redacted")
		assert.NotEqual(t, "exit status 128", errMsg, "error must not be a bare exit-status error")
		assert.True(t,
			strings.Contains(errMsg, "Connection refused: Git server is unreachable") ||
				strings.Contains(errMsg, "Could not resolve host: Unable to connect to Git server") ||
				strings.Contains(errMsg, "unreachable") ||
				strings.Contains(errMsg, "git clone failed"),
			"expected descriptive git failure, got: %s", errMsg,
		)
	})

	t.Run("Clone_WithPointerNopLogger_ReturnsDescriptiveSanitizedError", func(t *testing.T) {
		repo.TmpDirName = "test_clone_fail_ptr"
		repo.Logger = &task_logger.NopLogger{}
		err := client.Clone(repo)
		require.Error(t, err)

		var userErr *common_errors.UserVisibleError
		require.ErrorAs(t, err, &userErr, "error should be a UserVisibleError")

		errMsg := err.Error()
		assert.NotContains(t, errMsg, "my_secret_token", "secret token must be redacted")
		assert.NotEqual(t, "exit status 128", errMsg, "error must not be a bare exit-status error")
	})

	t.Run("AuthenticationFailure_ReturnsDescriptiveSanitizedError", func(t *testing.T) {
		// HTTP server returning 401 Unauthorized
		authServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Git"`)
			http.Error(w, "HTTP Basic: Access denied", http.StatusUnauthorized)
		}))
		defer authServer.Close()

		util.Config.ForwardedEnvVars = []string{"GIT_SSL_NO_VERIFY"}
		t.Setenv("GIT_SSL_NO_VERIFY", "true")

		authRepo := newTestGitRepo(t, authServer.URL+"/test/repo.git", "main")
		authRepo.Repository.SSHKey = db.AccessKey{
			Type: db.AccessKeyLoginPassword,
			LoginPassword: db.LoginPassword{
				Login:    "gitlab_user",
				Password: "my_secret_token",
			},
		}

		_, err := client.GetRemoteBranches(authRepo)
		require.Error(t, err)

		var userErr *common_errors.UserVisibleError
		require.ErrorAs(t, err, &userErr, "error should be a UserVisibleError")

		errMsg := err.Error()
		assert.NotContains(t, errMsg, "my_secret_token", "secret token must be redacted")
		assert.NotEqual(t, "exit status 128", errMsg, "error must not be a bare exit-status error")
		assert.Equal(t, "Authentication failed: Access denied (check token or password)", errMsg)
	})
}