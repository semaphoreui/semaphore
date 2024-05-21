package db_lib

import (
	"os"

	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
)

func getSensitiveEnvs() []string {
	return []string{
		"SEMAPHORE_ACCESS_KEY_ENCRYPTION",
		"SEMAPHORE_ADMIN_PASSWORD",
		"SEMAPHORE_DB_USER",
		"SEMAPHORE_DB_NAME",
		"SEMAPHORE_DB_HOST",
		"SEMAPHORE_DB_PASS",
		"SEMAPHORE_LDAP_PASSWORD",
	}
}

type LocalApp interface {
	SetLogger(logger task_logger.Logger) task_logger.Logger
	InstallRequirements() error
	Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error
}
