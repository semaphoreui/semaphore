package tasks

import "github.com/ansible-semaphore/semaphore/lib"

type RunnerPool struct {
	runners []RemoteRunner
}

func (p *RunnerPool) CreateJob(playbook *lib.AnsiblePlaybook) (AnsibleJob, error) {

	return p.runners[0].CreateJob(playbook)

	//return &LocalAnsibleJob{
	//	playbook: playbook,
	//}, nil
}
