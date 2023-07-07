package db

type ProjectUserRole string

const (
	ProjectUserOwner  ProjectUserRole = "owner"
	ProjectUserRunner ProjectUserRole = "runner"
	ProjectUserGuest  ProjectUserRole = "guest"
)

type ProjectUserPermission int

const (
	ProjectUserCanRunTask ProjectUserPermission = 1 << iota
	ProjectCanEditProjectSettings
	ProjectCanRunTasks
)

var rolePermissions = map[ProjectUserRole]ProjectUserPermission{
	ProjectUserOwner:  ProjectUserCanRunTask | ProjectCanEditProjectSettings | ProjectCanRunTasks,
	ProjectUserRunner: ProjectCanRunTasks,
	ProjectUserGuest:  0,
}

type ProjectUser struct {
	ID        int             `db:"id" json:"-"`
	ProjectID int             `db:"project_id" json:"project_id"`
	UserID    int             `db:"user_id" json:"user_id"`
	Admin     bool            `db:"admin" json:"admin"`
	Role      ProjectUserRole `db:"role" json:"role"`
}

func (u *ProjectUser) Can(permissions ProjectUserPermission) bool {
	userPermissions := rolePermissions[u.Role]
	return (userPermissions & userPermissions) == permissions
}
