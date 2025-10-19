package features

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func GetFeatures(user *db.User, plan string) map[string]bool {

	return map[string]bool{
		"project_runners":   false,
		"terraform_backend": false,
		"task_summary":      false,
		"secret_storages":   false,
	}
}
