package db

import "time"

// RoleMapping gives a project role to the members of an IdP group.
type RoleMapping struct {
	ID        int       `db:"id" json:"id"`
	Provider  string    `db:"provider" json:"provider"`
	Group     string    `db:"group_name" json:"group"`
	ProjectID int       `db:"project_id" json:"project_id"`
	RoleID    int       `db:"role_id" json:"-"`
	Created   time.Time `db:"created" json:"created"`
	// Filled by reads for the mappings page.
	ProjectName string `db:"project_name" json:"project_name"`
	RoleSlug    string `db:"role_slug" json:"role"`
	RoleName    string `db:"role_name" json:"role_name"`
}

// MappedRole is a role a member holds through a mapping.
type MappedRole struct {
	ProjectID int    `db:"project_id"`
	UserID    int    `db:"user_id"`
	Username  string `db:"username"`
	MappingID int    `db:"mapping_id"`
	Provider  string `db:"provider"`
	Group     string `db:"group_name"`
	RoleID    int    `db:"role_id"`
	RoleSlug  string `db:"role_slug"`
	RoleName  string `db:"role_name"`
}

// ProjectMember is what a user holds in a project: manual roles and roles from group mappings.
type ProjectMember struct {
	Roles       []Role
	MappedRoles []MappedRole
}
