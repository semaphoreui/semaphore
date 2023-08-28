package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
)

type RemoteJob struct {
	task        db.Task
	template    db.Template
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment
	playbook    *lib.AnsiblePlaybook
	logger      lib.Logger

	runnerPool RemoteRunnerPool
}

func (t *RemoteJob) Run(username string, incomingVersion *string) (err error) {

	var job *RemoteRunnerJob

	db.StoreSession(t.runnerPool.store, "create job", func() {
		job, err = t.runnerPool.CreateJob(username, incomingVersion, t)
	})

	if err != nil {
		return
	}

	return job.Wait()
}
