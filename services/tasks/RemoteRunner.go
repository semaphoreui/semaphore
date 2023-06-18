package tasks

import "github.com/ansible-semaphore/semaphore/lib"

type RemoteRunner struct {
}

func (r *RemoteRunner) CreateJob(playbook *lib.AnsiblePlaybook) (job AnsibleJob, err error) {
	return
}
