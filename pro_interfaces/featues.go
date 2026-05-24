package pro_interfaces

type Features struct {
	ProjectRunners          bool `json:"project_runners"`
	TerraformBackend        bool `json:"terraform_backend"`
	TaskSummary             bool `json:"task_summary"`
	SecretStorages          bool `json:"secret_storages"`
	SecretStorageManagement bool `json:"secret_storage_management"`
	CustomRolesManagement   bool `json:"custom_roles_management"`
	HighAvailability        bool `json:"high_availability"`
}
