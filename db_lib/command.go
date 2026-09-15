package db_lib

import (
	"os/exec"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func runCommand(cmd *exec.Cmd, stopCh <-chan struct{}, logger task_logger.Logger) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return waitCommand(cmd, stopCh, logger)
}
