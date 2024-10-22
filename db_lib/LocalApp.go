package db_lib

import (
	"os"

	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
)

func getEnvironmentVars() []string {
	res := []string{}
	for v, k := range util.Config.EnvironmentVars {
		res = append(res, v+"="+k)
	}
	return res
}

type LocalApp interface {
	SetLogger(logger task_logger.Logger) task_logger.Logger
	InstallRequirements(environmentVars *[]string) error
	Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error
}
