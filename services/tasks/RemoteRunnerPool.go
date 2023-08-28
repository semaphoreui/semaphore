package tasks

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"math/rand"
)

type RemoteRunnerJob struct {
	job             *RemoteJob
	username        string
	incomingVersion *string
}

func (c *RemoteRunnerJob) Wait() error {
	return nil
}

func (c *RemoteRunnerJob) WriteLogs(logRecords []LogRecord) {
	for _, record := range logRecords {
		c.job.Logger.Log2(record.Message, record.Time)
	}
}

// RemoteRunnerPool is a collection of the registered runners.
type RemoteRunnerPool struct {
	store    db.Store
	runners  map[int]*RemoteRunner
	taskPool *TaskPool
}

func CreateRunnerPool(store db.Store) RemoteRunnerPool {
	return RemoteRunnerPool{
		store: store,
	}
}

func (p *RemoteRunnerPool) GetRunner(runnerID int) (*RemoteRunner, error) {
	_, err := p.store.GetGlobalRunner(runnerID)

	if err != nil {
		if err == db.ErrNotFound {
			delete(p.runners, runnerID)
		}

		return nil, err
	}

	runner, ok := p.runners[runnerID]

	if !ok {
		runner = &RemoteRunner{
			queue: make([]RemoteRunnerJob, 0),
			pool:  p,
		}

		p.runners[runnerID] = runner
	}

	return runner, nil
}

func (p *RemoteRunnerPool) CreateJob(username string, incomingVersion *string, j *RemoteJob) (job *RemoteRunnerJob, err error) {

	runners, err := p.store.GetGlobalRunners()

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = fmt.Errorf("no runners")
		return
	}

	runner := runners[rand.Intn(len(runners))]

	remoteRunner, err := p.GetRunner(runner.ID)

	if err != nil {
		return
	}

	remoteRunner.EnqueueJob(RemoteRunnerJob{
		job:             j,
		incomingVersion: incomingVersion,
		username:        username,
	})

	return
}
