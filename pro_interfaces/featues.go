package pro_interfaces

type Features struct {
	ProjectRunners            bool `json:"project_runners"`
	TerraformBackend          bool `json:"terraform_backend"`
	TaskSummary               bool `json:"task_summary"`
	SecretStorages            bool `json:"secret_storages"`
	SecretStorageManagement   bool `json:"secret_storage_management"`
	SecretStorageManagementEx bool `json:"secret_storage_management_ex"`
	CustomRolesManagement     bool `json:"custom_roles_management"`
	// GlobalRolesManagement enables instance-level roles with global permissions and their assignment to
	// users. In a Pro plan only Pro users have it; in Enterprise everyone.
	GlobalRolesManagement bool `json:"global_roles_management"`
	HighAvailability      bool `json:"high_availability"`
	Workflows             bool `json:"workflows"`
	DockerExecutor        bool `json:"docker_executor"`
	K8sExecutor           bool `json:"k8s_executor"`
	AuditLogFilters       bool `json:"audit_log_filters"`
}
