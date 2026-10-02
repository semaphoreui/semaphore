package audit

import "github.com/semaphoreui/semaphore/db"

type AuthMethodMetadata struct {
	Method   string `json:"method"`
	Provider string `json:"provider,omitempty"`
}

type DenyMetadata struct {
	Method     string `json:"method"`
	Permission string `json:"permission,omitempty"`
}

type UserCreateMetadata struct {
	Admin    bool `json:"admin"`
	Pro      bool `json:"pro"`
	External bool `json:"external"`
}

type BoolChange struct {
	Old bool `json:"old"`
	New bool `json:"new"`
}

type UserUpdateMetadata struct {
	// Names only, never values.
	Fields []string    `json:"fields"`
	Admin  *BoolChange `json:"admin,omitempty"`
	Pro    *BoolChange `json:"pro,omitempty"`
}

type MembershipMetadata struct {
	Role        string `json:"role"`
	SelfRemoval bool   `json:"self_removal"`
}

type ProjectRoleMetadata struct {
	OldRole string `json:"old_role"`
	NewRole string `json:"new_role,omitempty"`
}

type RoleMetadata struct {
	Permissions []string `json:"permissions"`
}

type TemplatePermissionMetadata struct {
	TemplateID  int      `json:"template_id"`
	RoleSlug    string   `json:"role_slug,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type SettingsMetadata struct {
	// Keys only, never values.
	Keys []string `json:"keys"`
}

type LifecycleMetadata struct {
	Destinations []string `json:"destinations"`
}

func PermissionNames(p db.ProjectUserPermission) []string {
	names := []string{}
	if p&db.CanRunProjectTasks != 0 {
		names = append(names, "run_tasks")
	}
	if p&db.CanUpdateProject != 0 {
		names = append(names, "update_project")
	}
	if p&db.CanManageProjectResources != 0 {
		names = append(names, "manage_resources")
	}
	if p&db.CanManageProjectUsers != 0 {
		names = append(names, "manage_users")
	}
	return names
}

type ProjectCreateMetadata struct {
	Demo    bool `json:"demo"`
	Partial bool `json:"partial,omitempty"`
}

type BackupRestoreMetadata struct {
	// Object counts per backup section, never the objects.
	Objects map[string]int `json:"objects"`
	Partial bool           `json:"partial,omitempty"`
}

type TemplateMetadata struct {
	App                string `json:"app"`
	CreatedInventoryID int    `json:"created_inventory_id,omitempty"`
	Partial            bool   `json:"partial,omitempty"`
}

type TemplateInventoryMetadata struct {
	InventoryID int `json:"inventory_id"`
}

type ScheduleMetadata struct {
	TemplateID int `json:"template_id"`
}

type IntegrationMetadata struct {
	TemplateID int    `json:"template_id"`
	AuthMethod string `json:"auth_method"`
}

type IntegrationPartMetadata struct {
	IntegrationID int `json:"integration_id,omitempty"`
}

type HostConfigMetadata struct {
	Type string `json:"type"`
}

type EnvironmentMetadata struct {
	SecretsCreated int  `json:"secrets_created"`
	SecretsUpdated int  `json:"secrets_updated"`
	SecretsDeleted int  `json:"secrets_deleted"`
	Partial        bool `json:"partial,omitempty"`
}

// DeleteMetadata marks a delete whose row is gone but whose secrets were not all removed.
type DeleteMetadata struct {
	Partial bool `json:"partial,omitempty"`
}

type CredentialMetadata struct {
	Type string `json:"type"`
}

type SecretStorageMetadata struct {
	Type string `json:"type"`
}

type TaskCreateMetadata struct {
	Trigger       string `json:"trigger"`
	TemplateID    int    `json:"template_id"`
	ScheduleID    int    `json:"schedule_id,omitempty"`
	IntegrationID int    `json:"integration_id,omitempty"`
	ParentTaskID  int    `json:"parent_task_id,omitempty"`
	WorkflowRunID int    `json:"workflow_run_id,omitempty"`
}

type TaskCompleteMetadata struct {
	// The final task status, for example success, error, stopped or stopping.
	Result      string `json:"result"`
	EndReason   string `json:"end_reason,omitempty"`
	TemplateID  int    `json:"template_id"`
	InitiatorID int    `json:"initiator_id,omitempty"`
	DurationMS  int64  `json:"duration_ms"`
}

type TaskMetadata struct {
	TemplateID int `json:"template_id"`
}

type RunnerRegisterMetadata struct {
	Token string `json:"token"`
}
