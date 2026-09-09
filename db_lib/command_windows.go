//go:build windows

package db_lib

import (
	"errors"
	"os"
	"os/exec"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// waitCommand takes ownership of waiting for and terminating an already-started
// command. The caller must not wait for or signal the command after this call.
func waitCommand(cmd *exec.Cmd, stopCh <-chan struct{}, logger task_logger.Logger) error {
	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	select {
	case err := <-waitCh:
		return err
	case <-stopCh:
		// Windows has no SIGTERM signal. Future graceful process-tree termination
		// should use a Job Object, instead of killing only the main process.
		if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			logger.Log(err.Error())
		}
		return <-waitCh
	}
}
