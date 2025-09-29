package db

type Role struct {
	ID          int                   `db:"id" json:"id"`
	Slug        string                `db:"slug" json:"slug"`
	Name        string                `db:"name" json:"name"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}

func ValidateRole(role Role) error {
	if role.Name == "" {
		return &ValidationError{Message: "Role name cannot be empty"}
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
