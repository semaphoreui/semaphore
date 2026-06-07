package tasks

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// ErrAllRunnersBusy is returned when all available runners are busy. Used for logic
var ErrAllRunnersBusy = errors.New("all runners busy")

const runnerActiveThreshold = 30 * time.Minute

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

	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		err = fmt.Errorf("webhook returned incorrect status")
		return
	}

	return
}

func shuffleRunners(rs []db.Runner) []db.Runner {
	if len(rs) < 2 {
		return rs
	}

	// Work on a copy so that if randomness fails, we can safely return the original order.
	shuffled := make([]db.Runner, len(rs))
	copy(shuffled, rs)

	// Fisher–Yates shuffle using crypto/rand: for each i, pick j in [0, i].
	for i := len(shuffled) - 1; i > 0; i-- {
		max := big.NewInt(int64(i + 1))
		j, err := rand.Int(rand.Reader, max)
		if err != nil {
			log.WithError(err).Warn("failed to shuffle runners, using original order")
			return rs
		}

		ji := int(j.Int64())
		shuffled[i], shuffled[ji] = shuffled[ji], shuffled[i]
	}

	return shuffled
}

func (t *RemoteJob) Run(username string, incomingVersion *string, alias string) (err error) {
	tsk, err := t.taskPool.GetTask(t.Task.ID)

	if err != nil {
		return
	}

	if tsk == nil {
		return fmt.Errorf("task not found")
	}

	tsk.IncomingVersion = incomingVersion
	tsk.Username = username
	tsk.Alias = alias
	t.taskPool.state.UpdateRuntimeFields(tsk)

	var runners []db.Runner
	tagFilterMode := db.RunnerFilterTagCompleteMatch
	if t.RunnerTag == nil {
		tagFilterMode = db.RunnerFilterIsDefault
	}

	var projectRunners []db.Runner
	projectRunners, err = t.taskPool.store.GetRunners(t.Task.ProjectID, true, tagFilterMode, t.RunnerTag)
	if err != nil {
		return
	}

	var globalRunners []db.Runner
	globalRunners, err = t.taskPool.store.GetAllRunners(true, true, tagFilterMode, t.RunnerTag)
	if err != nil {
		return
	}

	runners = append(runners, shuffleRunners(projectRunners)...)
	runners = append(runners, shuffleRunners(globalRunners)...)

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = fmt.Errorf("no runners available")
		return
	}

	var runner *db.Runner
	now := tz.Now()

	// First pass: prefer runners with a recent heartbeat.
	// Second pass: fall back to runners that haven't reported recently.
	for pass := range 2 {
		for i := range runners {
			r := &runners[i]
			active := r.Touched != nil && now.Sub(*r.Touched) < runnerActiveThreshold || r.Webhook != ""
			if (pass == 0) == active {
				n := t.taskPool.GetNumberOfRunningTasksOfRunner(r.ID)
				if n < r.MaxParallelTasks || r.MaxParallelTasks == 0 {
					runner = r
					break
				}
			}
		}
		if runner != nil {
			break
		}
	}

	if runner == nil {
		err = ErrAllRunnersBusy
		return
	}

	err = callRunnerWebhook(runner, tsk, "start")

	if err != nil {
		return
	}

	tsk.Task.RunnerID = &runner.ID

	tsk.Logf("Task #%d is assigned to runner #%d", tsk.Task.ID, runner.ID)
	err = t.taskPool.store.UpdateTask(tsk.Task)

	if err != nil {
		return
	}

	t.taskPool.state.UpdateRuntimeFields(tsk)

	// The task now runs on the remote runner. Its completion is reported back
	// via the runner API (PUT /runners) and finalized by
	// TaskPool.FinalizeRemoteTask on whichever node receives the terminal
	// status. Returning here instead of polling means the task survives the
	// death or restart of the node that dispatched it: no node-local goroutine
	// owns its completion.
	t.scheduleTimeout(runner)
	return
}

// scheduleTimeout enforces util.Config.MaxTaskDurationSec for a dispatched
// remote task. The timer runs node-locally; if this node dies before it fires,
// the HA orphan cleaner applies the same limit as a backstop. Firing on an
// already-finished task is a no-op.
func (t *RemoteJob) scheduleTimeout(runner *db.Runner) {
	if util.Config.MaxTaskDurationSec <= 0 {
		return
	}
	d := time.Duration(util.Config.MaxTaskDurationSec) * time.Second
	taskID := t.Task.ID
	pool := t.taskPool
	time.AfterFunc(d, func() {
		tsk, err := pool.GetTask(taskID)
		if err != nil || tsk == nil {
			return
		}
		if util.HAEnabled() {
			pool.refreshTaskStatusFromDB(tsk)
		}
		if tsk.Task.Status.IsFinished() {
			return
		}
		tsk.Log("Task timed out")
		tsk.SetStatus(task_logger.TaskFailStatus)
		pool.FinalizeRemoteTask(tsk, runner)
	})
}

func (t *RemoteJob) Kill() {
	t.killed = true
	// Do nothing because you can't kill remote process
}

func (t *RemoteJob) IsKilled() bool {
	return t.killed
}

// Async is true: RemoteJob.Run only dispatches the task to a runner; its
// completion is reported asynchronously via the runner API.
func (t *RemoteJob) Async() bool {
	return true
}
