package runners

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type runningJob struct {
	status     task_logger.TaskStatus
	logRecords []LogRecord
	job        *tasks.LocalJob
	commit     *CommitInfo

	statusListeners []task_logger.StatusListener
	logListeners    []task_logger.LogListener

	logWG sync.WaitGroup
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

	go p.logPipe(stderr)
	go p.logPipe(stdout)
}

func (p *runningJob) WaitLog() {
	p.logWG.Wait()
}

func (p *runningJob) SetCommit(hash, message string) {
	p.commit = &CommitInfo{
		Hash:    hash,
		Message: message,
	}
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

func (p *runningJob) logPipe(reader io.Reader) {
	p.logWG.Add(1)
	defer p.logWG.Done()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		p.Log(line)
	}

	if scanner.Err() != nil && scanner.Err().Error() != "EOF" {
		//don't panic on these errors, sometimes it throws not dangerous "read |0: file already closed" error
		util.LogWarningF(scanner.Err(), log.Fields{"error": "Failed to read TaskRunner output"})
	}
}
