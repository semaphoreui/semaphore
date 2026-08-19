package runners

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	log "github.com/sirupsen/logrus"
)

type runningJob struct {
	// mu guards status, logRecords, commit and the listener slices. All of
	// these are touched concurrently by the job's own goroutine, the log-pipe
	// goroutines and the runner's progress/poll goroutine, so every access must
	// hold the lock. Listener callbacks are invoked WITHOUT the lock held
	// because they re-enter this type (e.g. LocalJob.SetStatus -> SetStatus and
	// the Terraform log listener calls SetStatus); Go mutexes are not reentrant.
	mu         sync.Mutex
	status     task_logger.TaskStatus
	logRecords []LogRecord

	// taskID is captured at enqueue time so log/error paths don't need to reach into
	// the executor for it. Necessary because executor is now an interface and the
	// id-bearing db.Task is owned by the concrete type.
	taskID int
	job    tasks.Executor
	commit *CommitInfo

	statusListeners []task_logger.StatusListener
	logListeners    []task_logger.LogListener

	logWG sync.WaitGroup
}

func (p *runningJob) AddStatusListener(l task_logger.StatusListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.statusListeners = append(p.statusListeners, l)
}

func (p *runningJob) AddLogListener(l task_logger.LogListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logListeners = append(p.logListeners, l)
}

func (p *runningJob) Log(msg string) {
	p.LogWithTime(tz.Now(), msg)
}

func (p *runningJob) Logf(format string, a ...any) {
	p.LogfWithTime(tz.Now(), format, a...)
}

func (p *runningJob) LogWithTime(now time.Time, msg string) {
	p.mu.Lock()
	p.logRecords = append(
		p.logRecords,
		LogRecord{
			Time:    now,
			Message: msg,
		},
	)
	listeners := make([]task_logger.LogListener, len(p.logListeners))
	copy(listeners, p.logListeners)
	p.mu.Unlock()

	for _, l := range listeners {
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
	p.mu.Lock()
	defer p.mu.Unlock()
	p.commit = &CommitInfo{
		Hash:    hash,
		Message: message,
	}
}

func (p *runningJob) SetStatus(status task_logger.TaskStatus) {
	p.mu.Lock()
	if p.status == status {
		p.mu.Unlock()
		return
	}
	p.status = status
	listeners := make([]task_logger.StatusListener, len(p.statusListeners))
	copy(listeners, p.statusListeners)
	p.mu.Unlock()

	// Invoked without the lock: p.job.SetStatus re-enters SetStatus (via
	// LocalJob.SetStatus -> Logger.SetStatus) and the listeners may do the same.
	p.job.SetStatus(status)

	for _, l := range listeners {
		l(status)
	}
}

// getStatus returns the current status under the lock.
func (p *runningJob) getStatus() task_logger.TaskStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}

// getProgress atomically snapshots the data needed to report progress to the
// server. The returned slice is a copy, so the caller can read it freely while
// the job keeps appending records.
func (p *runningJob) getProgress() (status task_logger.TaskStatus, logRecords []LogRecord, commit *CommitInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	status = p.status
	logRecords = make([]LogRecord, len(p.logRecords))
	copy(logRecords, p.logRecords)
	commit = p.commit
	return
}

// ackLogRecords drops the first sent records (those the server acknowledged),
// keeping any appended in the meantime, and returns the number still pending.
func (p *runningJob) ackLogRecords(sent int) (pending int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if sent <= len(p.logRecords) {
		p.logRecords = p.logRecords[sent:]
	} else {
		p.logRecords = nil
	}
	return len(p.logRecords)
}

func (p *runningJob) logPipe(reader io.Reader) {
	p.logWG.Add(1)
	defer p.logWG.Done()

	scanner := bufio.NewScanner(reader)
	const maxCapacity = 10 * 1024 * 1024 // 10 MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		p.Log(line)
	}

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

		p.job.Kill() // kill the job because stdout cannot be read.

		log.WithError(err).WithFields(log.Fields{
			"task_id": p.taskID,
			"context": "task_logger",
		}).Error(msg)

		p.Log("Fatal error: " + msg)
	}
}
