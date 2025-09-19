package tasks

import (
	"github.com/Digital-Data-Co/semaphore/services/tasks"
)

func NewTaskStateStore() tasks.TaskStateStore {
	return tasks.NewMemoryTaskStateStore()
}
