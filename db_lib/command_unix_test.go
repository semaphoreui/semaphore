//go:build !windows

package db_lib

import (
	"os/exec"
	"syscall"
	"testing"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStopCommandRespectsGracePeriod(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 0.1")
	require.NoError(t, cmd.Start())

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	require.NoError(t, err)
	// The command must not lead its group, so signaling -PID returns ESRCH.
	require.NotEqual(t, cmd.Process.Pid, pgid, "command must not lead its own process group")

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	// An immediate Process.Kill would make cmd.Wait return an *exec.ExitError.
	err = stopCommand(cmd, waitCh, task_logger.NopLogger{})
	assert.NoError(t, err)
}
