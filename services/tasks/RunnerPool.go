package tasks

import "github.com/ansible-semaphore/semaphore/lib"

// RunnerPool is a collection of the registered runners.
type RunnerPool struct {
}

func (p *RunnerPool) CreateJob(playbook *lib.AnsiblePlaybook) (AnsibleJob, error) {

	return &LocalAnsibleJob{playbook: playbook}, nil
}
