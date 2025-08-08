package tasks

import "github.com/semaphoreui/semaphore/util"

func NewTaskStateStore() TaskStateStore {
	if util.Config.HA.Enabled {
		return NewRedisTaskStateStore()
	}
	return NewMemoryTaskStateStore()
}
