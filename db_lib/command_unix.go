//go:build !windows

package db_lib

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

const gracePeriod = 15 * time.Second

// waitCommand takes ownership of waiting for and terminating an already-started
// command. The caller must not wait for or signal the command after this call.
func waitCommand(cmd *exec.Cmd, stopCh <-chan struct{}, logger task_logger.Logger) error {
	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	select {
	case err := <-waitCh:
		killProcessGroup(cmd, logger)
		return err
	case <-stopCh:
		return stopCommand(cmd, waitCh, logger, gracePeriod)
	}
}

func stopCommand(
	cmd *exec.Cmd,
	exitCh <-chan error,
	logger task_logger.Logger,
	termGracePeriod time.Duration,
) error {
	// Request a graceful shutdown for the process group.
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil &&
		!errors.Is(err, syscall.ESRCH) {
		logger.Logf("failed to send SIGTERM to process group: %v", err)
	}

	timer := time.NewTimer(termGracePeriod)
	defer timer.Stop()

	var exitErr error
	gracePeriodExpired := false
	select {
	case exitErr = <-exitCh:
	case <-timer.C:
		gracePeriodExpired = true
	}

	killProcessGroup(cmd, logger)
	if gracePeriodExpired {
		// Kill the main process directly as a fallback if it was not started as
		// the leader of a dedicated process group.
		if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			logger.Logf("failed to kill process: %v", err)
		}
		// If the process is still alive after the termination attempts,
		// waiting for cmd.Wait below may block indefinitely.
		exitErr = <-exitCh
	}
	return exitErr
}

func killProcessGroup(cmd *exec.Cmd, logger task_logger.Logger) {
	// The group leader may have been reaped and its PGID reused before the
	// SIGKILL. We explicitly accept this risk for process-group cleanup, as
	// discussed for Bazel: https://github.com/bazelbuild/bazel/issues/11910.
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err != nil && !errors.Is(err, syscall.ESRCH) {
		logger.Logf("failed to send SIGKILL to process group: %v", err)
	}
}
