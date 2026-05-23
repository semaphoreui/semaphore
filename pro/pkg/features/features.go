package features

import (
	"github.com/semaphoreui/semaphore/db"
)

type Features struct {
	ProjectRunners          bool `json:"project_runners"`
	TerraformBackend        bool `json:"terraform_backend"`
	TaskSummary             bool `json:"task_summary"`
	SecretStorages          bool `json:"secret_storages"`
	SecretStorageManagement bool `json:"secret_storage_management"`
	CustomRolesManagement   bool `json:"custom_roles_management"`
}

func GetFeatures(user *db.User, plan string) Features {
	return Features{}
}
