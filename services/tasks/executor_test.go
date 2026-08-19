package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestResolveGitBranch(t *testing.T) {
	tests := []struct {
		name       string
		repoBranch string
		template   db.Template
		task       db.Task
		expected   string
	}{
		{
			name:       "no overrides falls back to repository branch",
			repoBranch: "main",
			template:   db.Template{},
			task:       db.Task{},
			expected:   "main",
		},
		{
			name:       "template branch overrides repository branch",
			repoBranch: "main",
			template:   db.Template{GitBranch: new("release")},
			task:       db.Task{},
			expected:   "release",
		},
		{
			name:       "task branch ignored when override not allowed",
			repoBranch: "main",
			template:   db.Template{AllowOverrideBranchInTask: false},
			task:       db.Task{GitBranch: new("attacker")},
			expected:   "main",
		},
		{
			name:       "task branch ignored over pinned template branch when not allowed",
			repoBranch: "main",
			template:   db.Template{GitBranch: new("release"), AllowOverrideBranchInTask: false},
			task:       db.Task{GitBranch: new("attacker")},
			expected:   "release",
		},
		{
			name:       "task branch applied when override allowed",
			repoBranch: "main",
			template:   db.Template{AllowOverrideBranchInTask: true},
			task:       db.Task{GitBranch: new("feature")},
			expected:   "feature",
		},
		{
			name:       "empty task branch ignored even when override allowed",
			repoBranch: "main",
			template:   db.Template{GitBranch: new("release"), AllowOverrideBranchInTask: true},
			task:       db.Task{GitBranch: new("")},
			expected:   "release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, resolveGitBranch(tt.repoBranch, tt.template, tt.task))
		})
	}
}

// TestLocalExecutorImplementsExecutor is a compile-time guard: if LocalExecutor stops
// satisfying the Executor interface, this file fails to build. It documents the
// contract introduced in Phase 1 of the Kubernetes runner refactor.
func TestLocalExecutorImplementsExecutor(t *testing.T) {
	var _ Executor = (*LocalExecutor)(nil)
	var _ Job = (*LocalExecutor)(nil)
}

func TestLocalExecutorCleanup_ZeroValueSafe(t *testing.T) {
	// Cleanup runs in a defer after a potentially-failed Prepare, so it must tolerate
	// being called on a barely-initialized executor — in particular App may be nil
	// when Prepare didn't even get that far.
	exec := &LocalExecutor{}

	assert.NotPanics(t, func() {
		exec.Cleanup()
	})
}

func TestLocalExecutorIsKilled(t *testing.T) {
	exec := &LocalExecutor{}
	assert.False(t, exec.IsKilled(), "fresh executor should not be marked killed")

	exec.Kill() // Process is nil so this only flips the flag
	assert.True(t, exec.IsKilled(), "Kill() must flip the killed flag")
}
