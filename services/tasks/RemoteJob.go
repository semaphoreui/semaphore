package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
)

type RemoteJob struct {
	Task     db.Task
	taskPool *TaskPool
}

type runnerWebhookPayload struct {
	Action     string `json:"action"`
	ProjectID  int    `json:"project_id"`
	TaskID     int    `json:"task_id"`
	TemplateID int    `json:"template_id"`
	RunnerID   int    `json:"runner_id"`
}

func callRunnerWebhook(runner *db.Runner, tsk *TaskRunner, action string) (err error) {
	if runner.Webhook == "" {
		return
	}

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(runnerWebhookPayload{
		Action:     action,
		ProjectID:  tsk.Task.ProjectID,
		TaskID:     tsk.Task.ID,
		TemplateID: tsk.Template.ID,
		RunnerID:   runner.ID,
	})
	if err != nil {
		return
	}

	client := &http.Client{}

	var req *http.Request
	req, err = http.NewRequest("POST", runner.Webhook, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		err = fmt.Errorf("webhook returned incorrect status")
		return
	}

	return
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
		runners, err = t.taskPool.store.GetGlobalRunners(true)
	})

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = fmt.Errorf("no runners available")
		return
	}

	var runner *db.Runner

	for _, r := range runners {
		n := t.taskPool.GetNumberOfRunningTasksOfRunner(r.ID)
		if n < r.MaxParallelTasks {
			runner = &r
			break
		}
	}

	if runner == nil {
		err = fmt.Errorf("no runners available")
		return
	}

	err = callRunnerWebhook(runner, tsk, "start")

	if err != nil {
		return
	}

	tsk.RunnerID = runner.ID

	startTime := time.Now()

	taskTimedOut := false

	for {
		if util.Config.MaxTaskDurationSec > 0 && int(time.Now().Sub(startTime).Seconds()) > util.Config.MaxTaskDurationSec {
			taskTimedOut = true
			break
		}

		time.Sleep(1_000_000_000)
		tsk = t.taskPool.GetTask(t.Task.ID)
		if tsk.Task.Status == task_logger.TaskSuccessStatus ||
			tsk.Task.Status == task_logger.TaskStoppedStatus ||
			tsk.Task.Status == task_logger.TaskFailStatus {
			break
		}
	}

	err = callRunnerWebhook(runner, tsk, "finish")

	if err != nil {
		return
	}

	if tsk.Task.Status == task_logger.TaskFailStatus {
		err = fmt.Errorf("task failed")
	} else if taskTimedOut {
		err = fmt.Errorf("task timed out")
	}

	return
}

func (t *RemoteJob) Kill() {
	// Do nothing because you can't kill remote process
}
