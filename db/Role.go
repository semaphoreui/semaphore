package db

import (
	"regexp"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type Role struct {
	ID          int                   `db:"id" json:"id" backup:"-"`
	Slug        string                `db:"slug" json:"slug" backup:"-"`
	Name        string                `db:"name" json:"name"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	ProjectID   *int                  `db:"project_id" json:"project_id"`
	Scope       int                   `db:"scope" json:"-" backup:"-"`
	Builtin     bool                  `db:"builtin" json:"builtin" backup:"-"`
}

const KnownRolePermissions = CanRunProjectTasks | CanUpdateProject | CanManageProjectResources | CanManageProjectUsers

var roleSlugRE = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

// ValidateRole checks a role sent by a user. Slug rules apply only to new roles, so old slugs stay editable.
func ValidateRole(role Role, isNew bool) error {
	if role.Name == "" {
		return &common_errors.ValidationError{Message: "Role name cannot be empty"}
	}
	if role.Permissions&^KnownRolePermissions != 0 {
		return &common_errors.ValidationError{Message: "Role permissions contain unknown bits"}
	}
	if !isNew {
		return nil
	}
	if !roleSlugRE.MatchString(role.Slug) {
		return &common_errors.ValidationError{Message: "Role slug must be 1 to 100 characters: letters A-Z, digits, '_' and '-'"}
	}
	// "all" is taken by the /project/{id}/roles/all route.
	if ProjectUserRole(strings.ToLower(role.Slug)).IsValid() || strings.EqualFold(role.Slug, "all") {
		return &common_errors.ValidationError{Message: "Role slug is reserved: " + role.Slug}
	}
	return nil
}

type TemplateRolePerm struct {
	ID          int                   `db:"id" json:"id"`
	RoleID      int                   `db:"role_id" json:"-"`
	TemplateID  int                   `db:"template_id" json:"template_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	RoleSlug    string                `db:"role_slug" json:"role_slug"`
}

// RoleImpact counts what deleting a role together with its assignments removes.
type RoleImpact struct {
	Members        int `json:"members"`
	Projects       int `json:"projects"`
	TemplateGrants int `json:"template_grants"`
	Mappings       int `json:"mappings"`
}

// RoleMember is a project member who holds a role, with all roles the member holds in that project.
type RoleMember struct {
	ProjectID int
	UserID    int
	Username  string
	Roles     []Role
	// The member also holds roles through group mappings, so losing every manual role keeps the membership.
	Mapped bool
}

// RoleAssignments is everything that points to a role: the members who hold it, its template grants and its mappings.
type RoleAssignments struct {
	Members  []RoleMember
	Grants   []TemplateRolePerm
	Mappings []RoleMapping
	// Assignments that go with Mappings.
	MappedRoles []MappedRole
}

// Impact counts the assignments for the delete confirmation and the audit event.
func (a RoleAssignments) Impact() RoleImpact {
	type member struct{ projectID, userID int }
	members := map[member]bool{}
	projects := map[int]bool{}
	for _, m := range a.Members {
		members[member{m.ProjectID, m.UserID}] = true
		projects[m.ProjectID] = true
	}
	for _, m := range a.MappedRoles {
		members[member{m.ProjectID, m.UserID}] = true
		projects[m.ProjectID] = true
	}
	for _, g := range a.Grants {
		projects[g.ProjectID] = true
	}
	for _, m := range a.Mappings {
		projects[m.ProjectID] = true
	}
	return RoleImpact{Members: len(members), Projects: len(projects), TemplateGrants: len(a.Grants), Mappings: len(a.Mappings)}
}
