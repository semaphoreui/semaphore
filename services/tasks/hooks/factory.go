package hooks

import (
	"github.com/Digital-Data-Co/forge/db"
)

func GetHook(app db.TemplateApp) Hook {
	switch app {
	case db.AppAnsible:
		return &AnsibleHook{}
	default:
		return nil
	}
}
