package tasks

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
)

type RemoteRunnerJob struct {
	logger lib.Logger
}

func (c *RemoteRunnerJob) Wait() error {
	return nil
}

func (c *RemoteRunnerJob) WriteLogs(logRecords []LogRecord) {
	for _, record := range logRecords {
		c.logger.Log2(record.message, record.time)
	}
}

// RemoteRunnerPool is a collection of the registered runners.
type RemoteRunnerPool struct {
	store   db.Store
	runners map[int]*RemoteRunner
}

func CreateRunnerPool(store db.Store) RemoteRunnerPool {
	return RemoteRunnerPool{
		store: store,
	}
}

func (p *RemoteRunnerPool) GetOrAddRunner(runnerID int) (*RemoteRunner, error) {
	return nil, nil
}

func (p *RemoteRunnerPool) CreateJob(playbook *lib.AnsiblePlaybook) (job *RemoteRunnerJob, err error) {

	runners, err := p.store.GetGlobalRunners()

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = fmt.Errorf("no runners")
	}

	runner := runners[0] // TODO: get random active runner

	remoteRunner, err := p.GetOrAddRunner(runner.ID)

	if err != nil {
		return
	}

	job = &RemoteRunnerJob{}

	remoteRunner.AddJob(0, job)

	return
}
