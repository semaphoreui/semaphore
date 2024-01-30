package db_lib

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"os"
)

type TerraformApp struct {
	Logger     lib.Logger
	Playbook   *AnsiblePlaybook
	Template   db.Template
	Repository db.Repository
}

func (t *TerraformApp) SetLogger(logger lib.Logger) {
	t.Logger = logger
}

func (t *TerraformApp) InstallRequirements() error {
	return nil
}

func (t *TerraformApp) Run(args []string, environmentVars *[]string, cb func(*os.Process)) error {
	return nil
}
