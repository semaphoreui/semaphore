package db

import "github.com/semaphoreui/semaphore/pkg/common_errors"

type RoleKind uint8

const (
	RoleKindBuiltin RoleKind = 1 << iota
	RoleKindCustom
	RoleKindAll = RoleKindBuiltin | RoleKindCustom
)

// RoleQuery is sealed by its unexported marker method so only query variants
// declared in this package can implement it.
type RoleQuery interface {
	isRoleQuery()
}

type GlobalRoleQuery struct {
	Kinds RoleKind
}

type ProjectRoleQuery struct {
	ProjectID int
}

type AvailableRoleQuery struct {
	ProjectID int
	Kinds     RoleKind
}

func (GlobalRoleQuery) isRoleQuery()    {}
func (ProjectRoleQuery) isRoleQuery()   {}
func (AvailableRoleQuery) isRoleQuery() {}

type Role struct {
	Slug        string                `db:"slug" json:"slug" backup:"-"`
	Name        string                `db:"name" json:"name"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	ProjectID   *int                  `db:"project_id" json:"project_id"`
	IsBuiltin   bool                  `db:"is_builtin" json:"is_builtin"`
}

func ValidateRole(role Role) error {
	if role.Name == "" {
		return &common_errors.ValidationError{Message: "Role name cannot be empty"}
	}
	if role.Slug == "" {
		return &common_errors.ValidationError{Message: "Role slug cannot be empty"}
	}
	if role.IsBuiltin {
		return &common_errors.ValidationError{Message: "Custom role cannot be marked as built-in"}
	}
	// Built-in role slugs are reserved. Allowing a custom role to reuse one lets
	// it shadow the built-in role and escalate the permissions of its members.
	if ProjectUserRole(role.Slug).IsValid() {
		return &common_errors.ValidationError{Message: "Role slug is reserved and cannot be used: " + role.Slug}
	}
	return nil
}

type TemplateRolePerm struct {
	ID          int                   `db:"id" json:"id"`
	RoleSlug    string                `db:"role_slug" json:"role_slug"`
	TemplateID  int                   `db:"template_id" json:"template_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}
