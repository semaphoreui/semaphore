package k8s

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// Executor is a no-op stand-in that satisfies services/tasks.Executor so the runner's
// executor factory keeps type-checking. In OSS builds the JobPool never reaches New
// (Config initialisation fails first), so these methods exist only to keep the type
// graph consistent.
type Executor struct {
	killed bool
	logger task_logger.Logger
}

// New returns a stub Executor. The arguments mirror the proprietary signature so the
// factory in services/runners can call it unconditionally. None of the inputs are
// retained because the stub never does any work with them.
func New(_ Config, _ db.Task, _ db.Template, _ db.Inventory, _ db.Repository, _ db.Environment) *Executor {
	return &Executor{}
}

func (e *Executor) Prepare(_ string, _ *string, _ string) error { return ErrNotAvailable }

func (e *Executor) Run(_ string, _ *string, _ string) error { return ErrNotAvailable }

func (e *Executor) Cleanup() {}

func (e *Executor) Kill() { e.killed = true }

func (e *Executor) IsKilled() bool { return e.killed }

func (e *Executor) SetLogger(logger task_logger.Logger) {
	e.logger = logger
	if logger != nil {
		logger.Log(ErrNotAvailable.Error())
	}
}

func (e *Executor) SetStatus(status task_logger.TaskStatus) {
	if e.logger != nil {
		e.logger.SetStatus(status)
	}
}
