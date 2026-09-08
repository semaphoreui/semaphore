//go:build !windows

package tasks

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

const gracePeriod = 15 * time.Second

// stopProcess must return promptly because OnProcessStarted may call it
// synchronously. SIGKILL is therefore scheduled instead of waiting here.
func (t *LocalExecutor) stopProcess(process *os.Process, exitCh <-chan struct{}) {
	if process == nil {
		return
	}

	if err := syscall.Kill(-process.Pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return
		}

		t.Log(fmt.Sprintf("failed to send SIGTERM to process group: %v", err))
		if err := killProcess(process); err != nil {
			t.Log(err.Error())
		}
		return
	}

	go func() {
		timer := time.NewTimer(gracePeriod)
		defer timer.Stop()

		select {
		case <-exitCh:
		case <-timer.C:
		}

		// The group leader may have been reaped and its PGID reused before the
		// SIGKILL. We explicitly accept this risk for process-group cleanup, as
		// discussed for Bazel: https://github.com/bazelbuild/bazel/issues/11910.
		if err := syscall.Kill(-process.Pid, syscall.SIGKILL); err != nil &&
			!errors.Is(err, syscall.ESRCH) {
			t.Log(fmt.Sprintf("failed to send SIGKILL to process group: %v", err))
		}
	}()
}
