package db

type ProjectUserPermission int64

const (
	CanRunProjectTasks ProjectUserPermission = 1 << iota
	CanUpdateProject
	CanManageProjectResources
	CanManageProjectUsers
)

type ProjectUser struct {
	ID        int `db:"id" json:"-"`
	ProjectID int `db:"project_id" json:"project_id"`
	UserID    int `db:"user_id" json:"user_id"`
	RoleID    int `db:"role_id" json:"role_id"`
}
