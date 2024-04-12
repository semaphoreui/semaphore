package tasks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func (t *TaskRunner) Log(msg string) {
	t.LogWithTime(time.Now(), msg)
}

func (t *TaskRunner) Logf(format string, a ...any) {
	t.LogfWithTime(time.Now(), format, a...)
}

func (t *TaskRunner) LogWithTime(now time.Time, msg string) {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":       "log",
			"output":     msg,
			"time":       now,
			"task_id":    t.Task.ID,
			"project_id": t.Task.ProjectID,
		})

		util.LogPanic(err)
		sockets.Message(user, b)
	}

	t.pool.logger <- logRecord{
		task:   t,
		output: msg,
		time:   now,
	}
}

func (t *TaskRunner) LogfWithTime(now time.Time, format string, a ...any) {
	t.LogWithTime(now, fmt.Sprintf(format, a...))
}

func (t *TaskRunner) LogCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go t.logPipe(bufio.NewReader(stderr))
	go t.logPipe(bufio.NewReader(stdout))
}

func (t *TaskRunner) SetStatus(status lib.TaskStatus) {
	if status == t.Task.Status {
		return
	}

	switch t.Task.Status { // check old status
	case lib.TaskConfirmed:
		if status == lib.TaskWaitingConfirmation {
			return
		}
	case lib.TaskRunningStatus:
		if status == lib.TaskWaitingStatus {
			return
		}
	case lib.TaskStoppingStatus:
		if status == lib.TaskWaitingStatus || status == lib.TaskRunningStatus {
			//panic("stopping TaskRunner cannot be " + status)
			return
		}
	case lib.TaskSuccessStatus:
	case lib.TaskFailStatus:
	case lib.TaskStoppedStatus:
		return
	}

	t.Task.Status = status

	if status == lib.TaskRunningStatus {
		now := time.Now()
		t.Task.Start = &now
	}

	t.saveStatus()

	if localJob, ok := t.job.(*LocalJob); ok {
		localJob.SetStatus(status)
	}

	if status == lib.TaskFailStatus {
		t.sendMailAlert()
	}

	if status.IsNotifiable() {
		t.sendTelegramAlert()
		t.sendSlackAlert()
		t.sendRocketChatAlert()
		t.sendMicrosoftTeamsAlert()
	}
}

func (t *TaskRunner) panicOnError(err error, msg string) {
	if err != nil {
		t.Log(msg)
		util.LogPanicWithFields(err, log.Fields{"error": msg})
	}
}

func (t *TaskRunner) logPipe(reader *bufio.Reader) {
	line, err := Readln(reader)

	for err == nil {
		t.Log(line)
		line, err = Readln(reader)
	}

	if err != nil && err.Error() != "EOF" {
		//don't panic on these errors, sometimes it throws not dangerous "read |0: file already closed" error
		util.LogWarningWithFields(err, log.Fields{"error": "Failed to read TaskRunner output"})
	}
}

// Readln reads from the pipe
func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix = true
		err      error
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}
