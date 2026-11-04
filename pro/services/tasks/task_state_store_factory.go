package tasks

import (
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/tasks"
)

// NewTaskStateStore returns the in-memory store; the Redis-backed store that
// uses the shared HA connection is supplied by pro_impl.
func NewTaskStateStore(_ pro_interfaces.HARedis) tasks.TaskStateStore {
	return tasks.NewMemoryTaskStateStore()
}
