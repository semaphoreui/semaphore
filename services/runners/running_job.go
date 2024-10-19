package runners

import (
	"bufio"
	"fmt"
	"os/exec"
	"time"

	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type runningJob struct {
	status     task_logger.TaskStatus
	logRecords []LogRecord
	job        *tasks.LocalJob

	statusListeners []task_logger.StatusListener
	logListeners    []task_logger.LogListener
}

func (p *runningJob) AddStatusListener(l task_logger.StatusListener) {
	p.statusListeners = append(p.statusListeners, l)
}

func (p *runningJob) AddLogListener(l task_logger.LogListener) {
	p.logListeners = append(p.logListeners, l)
}

func (p *runningJob) Log(msg string) {
	p.LogWithTime(time.Now(), msg)
}

func (p *runningJob) Logf(format string, a ...any) {
	p.LogfWithTime(time.Now(), format, a...)
}

func (p *runningJob) LogWithTime(now time.Time, msg string) {
	p.logRecords = append(
		p.logRecords,
		LogRecord{
			Time:    now,
			Message: msg,
		},
	)
	for _, l := range p.logListeners {
		l(now, msg)
	}
}

func (p *runningJob) LogfWithTime(now time.Time, format string, a ...any) {
	p.LogWithTime(now, fmt.Sprintf(format, a...))
}

func (p *runningJob) LogCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go p.logPipe(bufio.NewReader(stderr))
	go p.logPipe(bufio.NewReader(stdout))
}

func (p *runningJob) SetStatus(status task_logger.TaskStatus) {
	if p.status == status {
		return
	}

	p.status = status
	p.job.SetStatus(status)

	for _, l := range p.statusListeners {
		l(status)
	}
}

func (p *runningJob) logPipe(reader *bufio.Reader) {
	line, err := tasks.Readln(reader)
	for err == nil {
		p.Log(line)
		line, err = tasks.Readln(reader)
	}

	if err != nil && err.Error() != "EOF" {
		//don't panic on these errors, sometimes it throws not dangerous "read |0: file already closed" error
		util.LogWarningWithFields(err, log.Fields{"error": "Failed to read TaskRunner output"})
	}
}
