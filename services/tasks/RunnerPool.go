package tasks

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
)

// RunnerPool is a collection of the registered runners.
type RunnerPool struct {
	store db.Store
}

func CreateRunnerPool(store db.Store) RunnerPool {
	return RunnerPool{
		store: store,
	}
}

func (p *RunnerPool) GetRunner(runnerID int) (*RemoteRunner, error) {
	return nil, nil
}

func (p *RunnerPool) CreateJob(playbook *lib.AnsiblePlaybook) (job LocalJob, err error) {

	runners, err := p.store.GetGlobalRunners()

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = fmt.Errorf("no runners")
	}

	job = LocalJob{}
	return
}
