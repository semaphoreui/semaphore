package db_lib

import (
	"github.com/ansible-semaphore/semaphore/lib"
	"os"
)

type LocalApp interface {
	SetLogger(logger lib.Logger) lib.Logger
	InstallRequirements() error
	Run(args []string, environmentVars *[]string, cb func(*os.Process)) error
}
