package tasks

import (
	"time"
)

type LogRecord struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

type RemoteRunner struct {
	ID    int
	queue []RemoteRunnerJob
	pool  *RemoteRunnerPool
}

func (r *RemoteRunner) EnqueueJob(job RemoteRunnerJob) {
	r.queue = append(r.queue, job)
}

func (r *RemoteRunner) WriteLogs(taskID int, logRecords []LogRecord) error {

	task := r.pool.taskPool.GetTask(taskID)

	for _, record := range logRecords {
		task.Log2(record.Message, record.Time)
	}

	return nil
}
