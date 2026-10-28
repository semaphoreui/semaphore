package db

import (
	"regexp"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// GlobalPermission is an instance-wide permission bit held by a global role. It is a separate type from
// ProjectUserPermission on purpose: the same numeric bit means different things in the two scopes.
type GlobalPermission int64

const (
	// CanManageGlobalUsers creates, edits and deletes users. Only an admin may grant the admin flag.
	CanManageGlobalUsers GlobalPermission = 1 << iota
	// CanManageGlobalRoles manages global roles and their assignments to users.
	CanManageGlobalRoles
	// CanManageGlobalSystem changes instance settings: options, apps, runners, cluster, active tasks.
	CanManageGlobalSystem
	// CanReadGlobalAudit reads the audit log.
	CanReadGlobalAudit
)

const AllGlobalPermissions = CanManageGlobalUsers | CanManageGlobalRoles | CanManageGlobalSystem | CanReadGlobalAudit

func (p GlobalPermission) Can(permission GlobalPermission) bool {
	return p&permission == permission
}

// GlobalPermissionDefinition describes one global permission bit for the API and the UI.
type GlobalPermissionDefinition struct {
	ID          string           `json:"id"`
	Permission  GlobalPermission `json:"permission"`
	Description string           `json:"description"`
}

// GlobalPermissionCatalog lists the global permissions in a stable order.
func GlobalPermissionCatalog() []GlobalPermissionDefinition {
	return []GlobalPermissionDefinition{
		{ID: "global.users.manage", Permission: CanManageGlobalUsers, Description: "Manage users"},
		{ID: "global.roles.manage", Permission: CanManageGlobalRoles, Description: "Manage global roles"},
		{ID: "global.system.manage", Permission: CanManageGlobalSystem, Description: "Manage system settings"},
		{ID: "global.audit.read", Permission: CanReadGlobalAudit, Description: "Read the audit log"},
	}
}

// Role grants project permissions. A role with no project is shared: it can be given to a member of any
// project. Instance-wide permissions live in GlobalRole, never here.
type Role struct {
	ID          int                   `db:"id" json:"id" backup:"-"`
	Slug        string                `db:"slug" json:"slug"`
	Name        string                `db:"name" json:"name"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	ProjectID   *int                  `db:"project_id" json:"project_id" backup:"-"`
	Scope       int                   `db:"scope" json:"-" backup:"-"`
	Builtin     bool                  `db:"builtin" json:"builtin" backup:"-"`
}

// GlobalRole grants instance-wide permissions to the users it is assigned to. It is kept apart from Role on
// purpose: project permissions and instance permissions never mix in one row.
type GlobalRole struct {
	ID          int              `db:"id" json:"id"`
	Slug        string           `db:"slug" json:"slug"`
	Name        string           `db:"name" json:"name"`
	Permissions GlobalPermission `db:"permissions" json:"permissions"`
}

const KnownRolePermissions = CanRunProjectTasks | CanUpdateProject | CanManageProjectResources | CanManageProjectUsers |
	CanRunWorkflows | CanEditWorkflows | CanApproveWorkflows

// conflictError is refused with 409 by the API: it matches ErrInvalidOperation and keeps its own message.
type conflictError struct{ message string }

func (e conflictError) Error() string { return e.message }

func (e conflictError) Is(target error) bool { return target == ErrInvalidOperation }

// ErrLastGlobalAdministrator refuses a change that would leave the instance without an administrator. The
// community store keeps the last admin-flag user; the Pro store also keeps the last holder of a global role
// with CanManageGlobalRoles when no admin-flag user exists.
var ErrLastGlobalAdministrator error = conflictError{"the instance must keep at least one administrator"}

// ErrGlobalRoleAssigned refuses to delete a global role that users hold until the caller asks to delete
// the assignments too.
var ErrGlobalRoleAssigned error = conflictError{"the role is assigned to users"}

var roleSlugRE = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

func validateRoleSlug(slug string) error {
	if !roleSlugRE.MatchString(slug) {
		return &common_errors.ValidationError{Message: "Role slug must be 1 to 100 characters: letters A-Z, digits, '_' and '-'"}
	}
	return nil
}

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
	if err := validateRoleSlug(role.Slug); err != nil {
		return err
	}
	if IsReservedRoleSlug(role.Slug) {
		return &common_errors.ValidationError{Message: "Role slug is reserved: " + role.Slug}
	}
	return nil
}

// IsReservedRoleSlug is true for the built-in slugs and "all", taken by the /project/{id}/roles/all route.
func IsReservedRoleSlug(slug string) bool {
	return ProjectUserRole(strings.ToLower(slug)).IsValid() || strings.EqualFold(slug, "all")
}

// ValidateGlobalRole checks a global role sent by a user. As with ValidateRole, the slug is checked only
// when the role is new.
func ValidateGlobalRole(role GlobalRole, isNew bool) error {
	if role.Name == "" {
		return &common_errors.ValidationError{Message: "Role name cannot be empty"}
	}
	if role.Permissions&^AllGlobalPermissions != 0 {
		return &common_errors.ValidationError{Message: "Role global permissions contain unknown bits"}
	}
	if !isNew {
		return nil
	}
	return validateRoleSlug(role.Slug)
}

type TemplateRolePerm struct {
	ID          int                   `db:"id" json:"id"`
	RoleID      int                   `db:"role_id" json:"-"`
	TemplateID  int                   `db:"template_id" json:"template_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	RoleSlug    string                `db:"role_slug" json:"role_slug"`
}

// GlobalRoleAssignment gives a user the permissions of a global role.
type GlobalRoleAssignment struct {
	ID     int `db:"id" json:"id"`
	UserID int `db:"user_id" json:"user_id"`
	RoleID int `db:"role_id" json:"role_id"`

	// Taken from the user and the global role when the assignment is read.
	Username          string           `db:"username" json:"username"`
	RoleSlug          string           `db:"role_slug" json:"role_slug"`
	RoleName          string           `db:"role_name" json:"role_name"`
	GlobalPermissions GlobalPermission `db:"global_permissions" json:"global_permissions"`
}

// GlobalPermissionDecision says whether a user holds one global permission and which role gives it.
type GlobalPermissionDecision struct {
	Permission string `json:"permission"`
	Allowed    bool   `json:"allowed"`
	RoleID     int    `json:"role_id,omitempty"`
	RoleName   string `json:"role_name,omitempty"`
}

// EffectiveGlobalPermissions is the mask a user holds and, per catalog entry, the first role that grants it.
type EffectiveGlobalPermissions struct {
	Permissions GlobalPermission           `json:"permissions"`
	Decisions   []GlobalPermissionDecision `json:"decisions"`
}

// EffectiveGlobalPermissionMask unions the assigned roles. An admin holds everything, whatever is assigned.
func EffectiveGlobalPermissionMask(isAdmin bool, assignments []GlobalRoleAssignment) GlobalPermission {
	if isAdmin {
		return AllGlobalPermissions
	}
	var permissions GlobalPermission
	for _, assignment := range assignments {
		permissions |= assignment.GlobalPermissions
	}
	return permissions
}

// ExplainEffectiveGlobalPermissions names, for every catalog entry, the role that makes it effective.
func ExplainEffectiveGlobalPermissions(isAdmin bool, assignments []GlobalRoleAssignment) EffectiveGlobalPermissions {
	catalog := GlobalPermissionCatalog()
	result := EffectiveGlobalPermissions{
		Permissions: EffectiveGlobalPermissionMask(isAdmin, assignments),
		Decisions:   make([]GlobalPermissionDecision, 0, len(catalog)),
	}
	for _, definition := range catalog {
		decision := GlobalPermissionDecision{Permission: definition.ID}
		if isAdmin {
			decision.Allowed, decision.RoleName = true, "Administrator"
		} else {
			for _, assignment := range assignments {
				if assignment.GlobalPermissions.Can(definition.Permission) {
					decision.Allowed, decision.RoleID, decision.RoleName = true, assignment.RoleID, assignment.RoleName
					break
				}
			}
		}
		result.Decisions = append(result.Decisions, decision)
	}
	return result
}

// RoleImpact counts what deleting a role together with its assignments removes. For a global role only
// Members is set: the users who hold it.
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
	// The member also holds roles through mappings of other roles, so losing every manual role keeps the membership.
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
	return RoleImpact{
		Members:        len(members),
		Projects:       len(projects),
		TemplateGrants: len(a.Grants),
		Mappings:       len(a.Mappings),
	}
}
