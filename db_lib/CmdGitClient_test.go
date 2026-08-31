package db_lib

import (
	"strings"
	"testing"
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
				"def456\trefs/heads/env/test",
				"ghi789\trefs/heads/release/v1.0",
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
	// Repo with password in URL pointing to non-existent location
	repo := newTestGitRepo(t, "https://gitlab_user:my_secret_token@127.0.0.1:59999/test/repo.git", "main")

	t.Run("GetRemoteBranches_ReturnsDescriptiveSanitizedError", func(t *testing.T) {
		branches, err := client.GetRemoteBranches(repo)
		if err == nil {
			t.Fatalf("expected error, got branches: %v", branches)
		}
		errMsg := err.Error()
		if !strings.HasPrefix(errMsg, "git ls-remote failed:") {
			t.Errorf("expected error prefix 'git ls-remote failed:', got: %s", errMsg)
		}
		if strings.Contains(errMsg, "my_secret_token") {
			t.Errorf("expected secret token to be sanitized, got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "unable to access") && !strings.Contains(errMsg, "fatal:") {
			t.Errorf("expected git stderr details in error message, got: %s", errMsg)
		}
	})

	t.Run("Clone_ReturnsDescriptiveSanitizedError", func(t *testing.T) {
		repo.TmpDirName = "test_clone_fail"
		err := client.Clone(repo)
		if err == nil {
			t.Fatalf("expected clone error, got nil")
		}
		errMsg := err.Error()
		if !strings.HasPrefix(errMsg, "git clone failed:") {
			t.Errorf("expected error prefix 'git clone failed:', got: %s", errMsg)
		}
		if strings.Contains(errMsg, "my_secret_token") {
			t.Errorf("expected secret token to be sanitized, got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "unable to access") && !strings.Contains(errMsg, "fatal:") {
			t.Errorf("expected git stderr details in error message, got: %s", errMsg)
		}
	})
}