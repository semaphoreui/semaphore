package db_lib

import (
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

func getEnvironmentVars() []string {
	res := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}

	for _, e := range util.Config.ForwardedEnvVars {
		v := os.Getenv(e)
		if v != "" {
			res = append(res, fmt.Sprintf("%s=%s", e, v))
		}
	}

	for k, v := range util.Config.EnvVars {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}

	return res
}

type LocalApp interface {
	SetLogger(logger task_logger.Logger) task_logger.Logger
	InstallRequirements(environmentVars *[]string) error
	Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error
}
