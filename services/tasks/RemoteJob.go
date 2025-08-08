package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"
	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type RemoteJob struct {
	RunnerTag *string
	Task      db.Task
	taskPool  *TaskPool
	killed    bool
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

	log.WithFields(log.Fields{
		"runner_id": runner.ID,
		"task_id":   tsk.Task.ID,
		"action":    action,
	}).Infof("Calling runner webhook")

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

	log.WithFields(log.Fields{
		"runner_id": runner.ID,
		"task_id":   tsk.Task.ID,
		"action":    action,
	}).Infof("Runner webhook returned %d", resp.StatusCode)

	return
}

func (t *RemoteJob) Run(username string, incomingVersion *string, alias string) (err error) {

	tsk := t.taskPool.GetTask(t.Task.ID)

	if tsk == nil {
		return fmt.Errorf("task not found")
	}

	tsk.IncomingVersion = incomingVersion
	tsk.Username = username
	tsk.Alias = alias
	if t.taskPool != nil && t.taskPool.state != nil {
		t.taskPool.state.UpdateRuntimeFields(tsk)
	}

	var runners []db.Runner
	db.StoreSession(t.taskPool.store, "run remote job", func() {
		var projectRunners []db.Runner
		projectRunners, err = t.taskPool.store.GetRunners(t.Task.ProjectID, true, t.RunnerTag)
		if err != nil {
			return
		}
		var globalRunners []db.Runner
		globalRunners, err = t.taskPool.store.GetAllRunners(true, true)
		if err != nil {
			return
		}
		runners = append(runners, projectRunners...)
		runners = append(runners, globalRunners...)
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
		if n < r.MaxParallelTasks || r.MaxParallelTasks == 0 {
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
	if t.taskPool != nil && t.taskPool.state != nil {
		t.taskPool.state.UpdateRuntimeFields(tsk)
	}

	startTime := tz.Now()

	taskTimedOut := false

	for {
		if util.Config.MaxTaskDurationSec > 0 && int(tz.Now().Sub(startTime).Seconds()) > util.Config.MaxTaskDurationSec {
			taskTimedOut = true
			break
		}

		time.Sleep(1_000_000_000)
		tsk = t.taskPool.GetTask(t.Task.ID)

		if tsk == nil {
			err = fmt.Errorf("task %d not found", t.Task.ID)
			return
		}

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
	t.killed = true
	// Do nothing because you can't kill remote process
}

func (t *RemoteJob) IsKilled() bool {
	return t.killed
}
