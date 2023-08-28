package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
)

type RemoteJob struct {
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment
	Playbook    *lib.AnsiblePlaybook
	Logger      lib.Logger

	RunnerPool RemoteRunnerPool
}

func (t *RemoteJob) Run(username string, incomingVersion *string) (err error) {

	var job *RemoteRunnerJob

	db.StoreSession(t.RunnerPool.store, "create job", func() {
		job, err = t.RunnerPool.CreateJob(username, incomingVersion, t)
	})

	if err != nil {
		return
	}

	return job.Wait()
}

func (t *RemoteJob) Kill() {
}
