//go:build !pro

package hooks

import "github.com/semaphoreui/semaphore/db"

func GetHook(app db.TemplateApp) Hook {
	switch app {
	case db.AppAnsible:
		return &AnsibleHook{}
	default:
		return nil
	}
}
