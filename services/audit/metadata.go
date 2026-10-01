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
