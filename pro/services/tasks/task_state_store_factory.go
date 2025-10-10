package tasks

import (
	"github.com/semaphoreui/semaphore/services/tasks"
)

func NewTaskStateStore() tasks.TaskStateStore {
	return tasks.NewMemoryTaskStateStore()
}
