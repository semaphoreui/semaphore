package hooks

import (
	"github.com/semaphoreui/semaphore/db"
)

type ProAnsibleHook struct {
}

func (h *ProAnsibleHook) End(store db.Store, projectID int, taskID int) {
}
