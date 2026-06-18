package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
