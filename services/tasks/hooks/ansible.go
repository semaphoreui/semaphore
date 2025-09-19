package hooks

import (
	"github.com/Digital-Data-Co/semaphore/db"
)

type AnsibleHook struct {
}

func (h *AnsibleHook) End(store db.Store, projectID int, taskID int) {
}
