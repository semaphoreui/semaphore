package tasks

import (
	"fmt"
	"time"
)

type LogRecord struct {
	time    time.Time
	message string
}

type RemoteRunner struct {
	jobs map[int]*RemoteRunnerJob
}

func (r *RemoteRunner) AddJob(job *RemoteRunnerJob) {
	if job == nil {
		panic("remote job cannot be nil")
	}

	r.jobs[job.job.task.ID] = job
}

func (r *RemoteRunner) WriteLogs(taskID int, logRecords []LogRecord) error {
	job, ok := r.jobs[taskID]
	if !ok {
		return fmt.Errorf("task not found")
	}

	job.WriteLogs(logRecords)

	return nil
}
