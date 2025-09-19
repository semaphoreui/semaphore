package hooks

import (
	"github.com/Digital-Data-Co/forge/db"
)

type AnsibleHook struct {
}

func (h *AnsibleHook) End(store db.Store, projectID int, taskID int) {
}
