package runners

import (
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
)

type JobData struct {
	Username            string
	IncomingVersion     *string
	Alias               string
	Task                db.Task        `json:"task" binding:"required"`
	Template            db.Template    `json:"template" binding:"required"`
	Inventory           db.Inventory   `json:"inventory" binding:"required"`
	InventoryRepository *db.Repository `json:"inventory_repository" binding:"required"`
	Repository          db.Repository  `json:"repository" binding:"required"`
	Environment         db.Environment `json:"environment" binding:"required"`
}

type RunnerState struct {
	CurrentJobs []JobState
	NewJobs     []JobData            `json:"new_jobs" binding:"required"`
	AccessKeys  map[int]db.AccessKey `json:"access_keys" binding:"required"`
}

type JobState struct {
	ID     int                    `json:"id" binding:"required"`
	Status task_logger.TaskStatus `json:"status" binding:"required"`
}

type LogRecord struct {
	Time    time.Time `json:"time" binding:"required"`
	Message string    `json:"message" binding:"required"`
}

type CommitInfo struct {
	Hash    string `json:"hash" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type RunnerProgress struct {
	Jobs []JobProgress
}

type JobProgress struct {
	ID         int
	Status     task_logger.TaskStatus
	LogRecords []LogRecord
	Commit     *CommitInfo
}

type RunnerRegistration struct {
	RegistrationToken string `json:"registration_token" binding:"required"`
	Webhook           string `json:"webhook"`
	MaxParallelTasks  int    `db:"max_parallel_tasks" json:"max_parallel_tasks"`
}

type jobLogRecord struct {
	taskID int
	record LogRecord
}

type job struct {
	username        string
	incomingVersion *string
	alias           string

	// job presents remote or local job information
	job    *tasks.LocalJob
	status task_logger.TaskStatus
}
