package db_lib

import (
	"os"

	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
)

type LocalApp interface {
	SetLogger(logger task_logger.Logger) task_logger.Logger
	InstallRequirements() error
	Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error
}
