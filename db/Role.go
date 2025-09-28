package db

type Role struct {
	ID          int                   `db:"id" json:"id"`
	Slug        string                `db:"slug" json:"slug"`
	Name        string                `db:"name" json:"name"`
	ProjectID   *int                  `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}

type TemplateRole struct {
	ID          int                   `db:"id" json:"id"`
	RoleID      int                   `db:"role_id" json:"role_id"`
	TemplateID  int                   `db:"template_id" json:"template_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}

type ViewRole struct {
	ID          int                   `db:"id" json:"id"`
	RoleID      int                   `db:"role_id" json:"role_id"`
	ViewID      int                   `db:"view_id" json:"view_id"`
	ProjectID   int                   `db:"project_id" json:"project_id"`
	Permissions ProjectUserPermission `db:"permissions" json:"permissions"`
}
