package tasks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"

	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func (t *TaskRunner) Log(msg string) {
	t.LogWithTime(tz.Now(), msg)
}

func (t *TaskRunner) Logf(format string, a ...any) {
	t.LogfWithTime(tz.Now(), format, a...)
}

func (t *TaskRunner) LogWithTime(now time.Time, msg string) {
	t.sendToWs(now, msg)

	t.pool.logger <- logRecord{
		task:   t,
		output: msg,
		time:   now,
	}

	for _, l := range t.logListeners {
		l(now, msg)
	}
}

func (t *TaskRunner) sendToWs(now time.Time, msg string) {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]any{
			"type":       "log",
			"output":     msg,
			"time":       now,
			"task_id":    t.Task.ID,
			"project_id": t.Task.ProjectID,
		})

		util.LogPanic(err)
		sockets.Message(user, b)
	}
}

func (t *TaskRunner) LogfWithTime(now time.Time, format string, a ...any) {
	t.LogWithTime(now, fmt.Sprintf(format, a...))
}

func (t *TaskRunner) LogCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go t.logPipe(stderr)
	go t.logPipe(stdout)
}

func (t *TaskRunner) WaitLog() {
	t.logWG.Wait()
}

func (t *TaskRunner) SetCommit(hash, message string) {

	t.Task.CommitHash = &hash
	t.Task.CommitMessage = message

	if err := t.pool.store.UpdateTask(t.Task); err != nil {
		t.panicOnError(err, "Failed to update task commit")
	}
}

func (t *TaskRunner) SetStatus(status task_logger.TaskStatus) {
	if status == t.Task.Status {
		return
	}

	switch t.Task.Status { // check old status
	case task_logger.TaskConfirmed:
		if status == task_logger.TaskWaitingConfirmation {
			return
		}
	case task_logger.TaskRunningStatus:
		if status == task_logger.TaskWaitingStatus {
			return
		}
	case task_logger.TaskStoppingStatus:
		if status == task_logger.TaskWaitingStatus || status == task_logger.TaskRunningStatus {
			//panic("stopping TaskRunner cannot be " + status)
			return
		}
	case task_logger.TaskSuccessStatus:
	case task_logger.TaskFailStatus:
	case task_logger.TaskStoppedStatus:
		return
	}

	t.Task.Status = status

	if status == task_logger.TaskRunningStatus {
		now := tz.Now()
		t.Task.Start = &now
	}

	t.saveStatus()

	if localJob, ok := t.job.(*LocalJob); ok {
		localJob.SetStatus(status)
	}

	if status == task_logger.TaskFailStatus {
		t.sendMailAlert()
	}

	if status.IsNotifiable() {
		t.sendTelegramAlert()
		t.sendSlackAlert()
		t.sendRocketChatAlert()
		t.sendMicrosoftTeamsAlert()
		t.sendDingTalkAlert()
		t.sendGotifyAlert()
	}

	for _, l := range t.statusListeners {
		l(status)
	}
}

func (t *TaskRunner) panicOnError(err error, msg string) {
	if err == nil {
		return
	}

	t.Log(msg)
	util.LogPanicF(err, log.Fields{"error": msg})
}

func (t *TaskRunner) logPipe(reader io.Reader) {
	t.logWG.Add(1)

	linesCh := make(chan string, 100000)

	go func() {
		defer t.logWG.Done()

		for line := range linesCh {
			t.Log(line)
		}
	}()

	scanner := bufio.NewScanner(reader)
	const maxCapacity = 10 * 1024 * 1024 // 10 MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		linesCh <- line
	}

	close(linesCh)

	err := scanner.Err()

	if err != nil {
		msg := "Failed to read TaskRunner output"

		switch err.Error() {
		case "EOF",
			"os: process already finished",
			"read |0: file already closed":
			return // it is ok
		case "bufio.Scanner: token too long":
			msg = "TaskRunner output exceeds the maximum allowed size of 10MB"
		}

		t.kill() // kill the job because stdout cannot be read.

		log.WithError(err).WithFields(log.Fields{
			"task_id": t.Task.ID,
			"context": "task_logger",
		}).Error(msg)

		t.Log("Fatal error: " + msg)
	}
}
