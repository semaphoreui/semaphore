package db

import (
	"errors"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type RoleKind uint8

const (
	RoleKindBuiltin RoleKind = 1 << iota
	RoleKindCustom
	RoleKindAll = RoleKindBuiltin | RoleKindCustom
)

type BuiltinRoleKey string

const (
	BuiltinRoleOwner      BuiltinRoleKey = "owner"
	BuiltinRoleManager    BuiltinRoleKey = "manager"
	BuiltinRoleTaskRunner BuiltinRoleKey = "task_runner"
	BuiltinRoleGuest      BuiltinRoleKey = "guest"
)

// RolesQuery is sealed by its unexported marker method so only query variants
// declared in this package can implement it.
type RolesQuery interface {
	isRolesQuery()
}

type GlobalRolesQuery struct {
	Kinds RoleKind
}

type ProjectRolesQuery struct {
	ProjectID int
}

type AvailableRolesQuery struct {
	ProjectID int
	Kinds     RoleKind
}

func (GlobalRolesQuery) isRolesQuery()    {}
func (ProjectRolesQuery) isRolesQuery()   {}
func (AvailableRolesQuery) isRolesQuery() {}

// RoleQuery describes the identity used to load one role.
type RoleQuery interface {
	isRoleQuery()
}

type RoleByIDQuery struct {
	ID int
}

type BuiltinRoleQuery struct {
	Key BuiltinRoleKey
}

func (RoleByIDQuery) isRoleQuery()    {}
func (BuiltinRoleQuery) isRoleQuery() {}

type Role struct {
	ID          int                   `db:"id" json:"id" backup:"-"`
	Name        string                `db:"name" json:"name"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	ProjectID   *int                  `db:"project_id" json:"project_id"`
	BuiltinKey  *BuiltinRoleKey       `db:"builtin_key" json:"builtin_key"`
}

func (r Role) IsBuiltin() bool {
	return r.BuiltinKey != nil
}

func (r Role) IsAvailableToProject(projectID int) bool {
	return r.ProjectID == nil || *r.ProjectID == projectID
}

// ResolveRoleForProject returns the requested role when it is global or belongs
// to projectID. A caller may accidentally or deliberately supply an ID owned by
// another project; missing and wrong-scope IDs return the same error so this
// check cannot disclose whether the role exists.
func ResolveRoleForProject(store RoleRepository, roleID int, projectID int) (Role, error) {
	role, err := store.GetRole(RoleByIDQuery{ID: roleID})

	switch {
	case errors.Is(err, ErrNotFound):
		return Role{}, roleUnavailableError()
	case err != nil:
		return Role{}, err
	case !role.IsAvailableToProject(projectID):
		return Role{}, roleUnavailableError()
	default:
		return role, nil
	}
}

func roleUnavailableError() error {
	return common_errors.NewValidationError("Role does not exist or is not available to this project")
}

func ValidateCustomRole(role Role) error {
	if role.Name == "" {
		return &common_errors.ValidationError{Message: "Role name cannot be empty"}
	}
	if role.IsBuiltin() {
		return &common_errors.ValidationError{Message: "Custom role cannot have a built-in key"}
	}
	return nil
}

type TemplateRolePerm struct {
	ID          int                   `db:"id" json:"id"`
	RoleID      int                   `db:"role_id" json:"role_id"`
	TemplateID  int                   `db:"template_id" json:"template_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}
