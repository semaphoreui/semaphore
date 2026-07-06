package db

type Role struct {
	Slug        string                `db:"slug" json:"slug" backup:"-"`
	Name        string                `db:"name" json:"name"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
	ProjectID   *int                  `db:"project_id" json:"project_id"`
}

func ValidateRole(role Role) error {
	if role.Name == "" {
		return &ValidationError{Message: "Role name cannot be empty"}
	}
	if role.Slug == "" {
		return &ValidationError{Message: "Role slug cannot be empty"}
	}
	// Built-in role slugs are reserved. Allowing a custom role to reuse one lets
	// it shadow the built-in role and escalate the permissions of its members.
	if ProjectUserRole(role.Slug).IsValid() {
		return &ValidationError{Message: "Role slug is reserved and cannot be used: " + role.Slug}
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
