//go:build !windows

package tasks

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

const terminationGracePeriod = 15 * time.Second

// stopProcess must return promptly because OnProcessStarted may call it
// synchronously. SIGKILL is therefore scheduled instead of waiting here.
func (t *LocalExecutor) stopProcess(process *os.Process) {
	if process == nil {
		return
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return
		}

		t.Log(fmt.Sprintf("failed to send SIGTERM: %v", err))
		t.killProcess(process)
		return
	}

	time.AfterFunc(terminationGracePeriod, func() {
		t.killProcess(process)
	})
}
