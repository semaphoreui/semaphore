package tasks

import "github.com/ansible-semaphore/semaphore/lib"

type RemoteRunner struct {
	// TODO: wrapper to TCP
}

func (r *RemoteRunner) CreateJob(playbook *lib.AnsiblePlaybook) (job AnsibleJob, err error) {
	// TODO: put job to queue

	return
}
