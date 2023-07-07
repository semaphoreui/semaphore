package db

type ProjectUserRole string

const (
	ProjectOwner      ProjectUserRole = "owner"
	ProjectTaskRunner ProjectUserRole = "task_runner"
	ProjectGuest      ProjectUserRole = "guest"
)

type ProjectUserPermission int

const (
	ProjectUserCanRunTask ProjectUserPermission = 1 << iota
	ProjectCanEditProjectSettings
	ProjectCanRunTasks
)

var rolePermissions = map[ProjectUserRole]ProjectUserPermission{
	ProjectOwner:      ProjectUserCanRunTask | ProjectCanEditProjectSettings | ProjectCanRunTasks,
	ProjectTaskRunner: ProjectCanRunTasks,
	ProjectGuest:      0,
}

type ProjectUser struct {
	ID        int             `db:"id" json:"-"`
	ProjectID int             `db:"project_id" json:"project_id"`
	UserID    int             `db:"user_id" json:"user_id"`
	Role      ProjectUserRole `db:"role" json:"role"`
}

func (u *ProjectUser) Can(permissions ProjectUserPermission) bool {
	userPermissions := rolePermissions[u.Role]
	return (userPermissions & userPermissions) == permissions
}
