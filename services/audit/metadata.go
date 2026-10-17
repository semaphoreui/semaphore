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

// RoleRef names a role by its unique ID and by the slug it had at the time.
type RoleRef struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

func NewRoleRef(r db.Role) RoleRef {
	return RoleRef{ID: r.ID, Slug: TruncateName(r.Slug, MaxNameBytes)}
}

func NewRoleRefs(roles []db.Role) []RoleRef {
	refs := make([]RoleRef, 0, len(roles))
	for _, role := range roles {
		refs = append(refs, NewRoleRef(role))
	}
	return refs
}

type MembershipMetadata struct {
	Roles       []RoleRef `json:"roles"`
	SelfRemoval bool      `json:"self_removal"`
	// Set when the member lost the last role because the role was deleted.
	DeletedRole RoleRef `json:"deleted_role,omitzero"`
}

type ProjectRoleMetadata struct {
	OldRoles []RoleRef `json:"old_roles"`
	NewRoles []RoleRef `json:"new_roles,omitempty"`
	// Set when the role was taken away because it was deleted.
	DeletedRole RoleRef `json:"deleted_role,omitzero"`
}

// RoleMetadata describes a role definition: project permission names for a project role, global permission
// ids for a global role.
type RoleMetadata struct {
	Slug        string   `json:"slug"`
	Permissions []string `json:"permissions"`
}

// GlobalRoleMetadata describes a global role given to or taken from a user; the target is the user.
type GlobalRoleMetadata struct {
	Role              RoleRef  `json:"role"`
	GlobalPermissions []string `json:"global_permissions"`
}

type RoleDeleteMetadata struct {
	Slug string `json:"slug"`
	// Set only when the role was deleted together with its assignments.
	Impact *db.RoleImpact `json:"impact,omitempty"`
}

type TemplatePermissionMetadata struct {
	TemplateID  int      `json:"template_id"`
	Role        RoleRef  `json:"role,omitzero"`
	Permissions []string `json:"permissions,omitempty"`
}

type SettingsMetadata struct {
	// Keys only, never values.
	Keys []string `json:"keys"`
}

type LifecycleMetadata struct {
	Destinations []string `json:"destinations"`
}

// ExportMetadata describes a download of audit events: the query string filters and how many events were sent.
type ExportMetadata struct {
	Format   string              `json:"format"`
	Filters  map[string][]string `json:"filters"`
	Events   int64               `json:"events"`
	Complete bool                `json:"complete"`
}

type RetentionMetadata struct {
	Deleted       int64 `json:"deleted"`
	LastSeq       int64 `json:"last_seq"`
	RetentionDays int   `json:"retention_days"`
}

// GlobalPermissionNames lists the global permissions of a mask by their catalog ids.
func GlobalPermissionNames(p db.GlobalPermission) []string {
	names := []string{}
	for _, definition := range db.GlobalPermissionCatalog() {
		if p.Can(definition.Permission) {
			names = append(names, definition.ID)
		}
	}
	return names
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
