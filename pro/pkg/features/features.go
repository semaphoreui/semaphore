package features

import "github.com/Digital-Data-Co/forge/db"

func GetFeatures(user *db.User) map[string]bool {
	return map[string]bool{
		"project_runners":   false,
		"terraform_backend": false,
		"task_summary":      false,
		"secret_storages":   false,
	}
}
