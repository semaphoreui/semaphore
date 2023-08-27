package tasks

import (
	"github.com/ansible-semaphore/semaphore/lib"
	"os"
)

type AnsibleJob interface {
	RunGalaxy(args []string) error
	RunPlaybook(args []string, environmentVars *[]string, cb func(*os.Process)) error
}

type LocalAnsibleJob struct {
	playbook *lib.AnsiblePlaybook
}

func (j *LocalAnsibleJob) RunGalaxy(args []string) error {
	return j.playbook.RunGalaxy(args)
}

func (j *LocalAnsibleJob) RunPlaybook(args []string, environmentVars *[]string, cb func(*os.Process)) error {
	return j.playbook.RunPlaybook(args, environmentVars, cb)
}
