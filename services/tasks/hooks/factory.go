package hooks

import (
	"github.com/Digital-Data-Co/semaphore/db"
)

func GetHook(app db.TemplateApp) Hook {
	switch app {
	case db.AppAnsible:
		return &AnsibleHook{}
	default:
		return nil
	}
}
