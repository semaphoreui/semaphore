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
	// SubmoduleCredentials maps submodule hosts to the access key used to
	// authenticate their clone, for repositories whose submodules live on a
	// different host/credentials than Repository itself.
	SubmoduleCredentials []db.RepositorySubmoduleCredential `json:"submodule_credentials,omitempty"`
	Environment          db.Environment                     `json:"environment" binding:"required"`
	JWT                  string                             `json:"jwt,omitempty"`
}

type RunnerState struct {
	CurrentJobs []JobState
	NewJobs     []JobData            `json:"new_jobs" binding:"required"`
	AccessKeys  map[int]db.AccessKey `json:"access_keys" binding:"required"`

	ClearCache          bool `json:"clear_cache,omitempty"`
	CacheCleanProjectID *int `json:"cache_clean_project_id,omitempty"`
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

// RunnerProgressResponse is the server's reply to a progress report (PUT).
// TerminatedJobs lists task IDs the server considers finished (e.g. stopped
// while the runner was offline) and whose results it no longer accepts; the
// runner must emergency-stop those jobs. Old servers reply 204 with an empty
// body and old runners ignore the body, so the field is backward compatible
// in both directions.
type RunnerProgressResponse struct {
	TerminatedJobs []int `json:"terminated_jobs,omitempty"`
}

type JobProgress struct {
	ID         int
	Status     task_logger.TaskStatus
	LogRecords []LogRecord
	Commit     *CommitInfo
}

type RunnerRegistration struct {
	// RegistrationToken is either the shared global registration token (which
	// creates a new runner) or a one-time token issued for a specific unregistered
	// runner (which registers that runner).
	RegistrationToken string   `json:"registration_token" binding:"required"`
	Webhook           string   `json:"webhook,omitempty"`
	Name              string   `json:"name,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	MaxParallelTasks  int      `json:"max_parallel_tasks"`
	Enabled           bool     `json:"enabled,omitempty"`
	ProjectID         *int     `json:"project_id,omitempty"`
}

type jobLogRecord struct {
	taskID int
	record LogRecord
}

type job struct {
	username        string
	incomingVersion *string
	alias           string

	// job is the executor that will run this task on the runner host. The concrete
	// type is selected by the factory at enqueue time (LocalExecutor for local
	// execution, KubernetesExecutor when the runner is configured to dispatch into
	// Pods). Field accesses that need the original db.Task/db.Template/... go
	// through taskID/template — kept here so progress reporting and orphan cleanup
	// don't need to type-assert back to the concrete executor.
	job    tasks.Executor
	taskID int
	status task_logger.TaskStatus
}
