package db_lib

import (
	"fmt"
	"os"

	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
)

func getEnvironmentVars() []string {
	res := []string{}
	for k, v := range util.Config.EnvironmentVars {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}

type LocalApp interface {
	SetLogger(logger task_logger.Logger) task_logger.Logger
	InstallRequirements(environmentVars *[]string) error
	Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error
}
