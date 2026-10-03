package db_lib

import (
	"os/exec"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func runCommand(cmd *exec.Cmd, stopCh <-chan struct{}, logger task_logger.Logger) error {
	// This is the last pre-run cancellation check. The check and cmd.Start are
	// not atomic; if cancellation wins between them, waitCommand stops the process.
	select {
	case <-stopCh:
		return nil
	default:
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	return waitCommand(cmd, stopCh, logger)
}
