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
	return nil
}

type TemplateRolePerm struct {
	ID          int                   `db:"id" json:"id"`
	RoleSlug    string                `db:"role_slug" json:"role_slug"`
	TemplateID  int                   `db:"template_id" json:"template_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}
