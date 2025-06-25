package hooks

import (
	"github.com/semaphoreui/semaphore/db"
)

type AnsibleHook struct {
}

func (h *AnsibleHook) End(store db.Store, projectID int, taskID int) {
}
