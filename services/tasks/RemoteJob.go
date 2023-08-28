package tasks

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"math/rand"
	"time"
)

type RemoteJob struct {
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment
	Playbook    *lib.AnsiblePlaybook
	Logger      lib.Logger

	taskPool TaskPool
}

func (t *RemoteJob) Run(username string, incomingVersion *string) (err error) {

	tsk := t.taskPool.GetTask(t.Task.ID)

	if tsk == nil {
		return fmt.Errorf("task not found")
	}

	tsk.IncomingVersion = incomingVersion
	tsk.Username = username

	var runners []db.Runner
	db.StoreSession(t.taskPool.store, "run remote job", func() {
		runners, err = t.taskPool.store.GetGlobalRunners()
	})

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = fmt.Errorf("no runners")
		return
	}

	runner := runners[rand.Intn(len(runners))]

	if err != nil {
		return
	}

	tsk.RunnerID = runner.ID

	for {
		time.Sleep(100000)
		tsk = t.taskPool.GetTask(t.Task.ID)
		if tsk.task.Status == db.TaskSuccessStatus {
			break
		}
	}

	return
}

func (t *RemoteJob) Kill() {
}
