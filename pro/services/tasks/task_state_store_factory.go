package tasks

import (
	"github.com/Digital-Data-Co/forge/services/tasks"
)

func NewTaskStateStore() tasks.TaskStateStore {
	return tasks.NewMemoryTaskStateStore()
}
