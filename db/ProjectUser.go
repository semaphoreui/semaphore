package db

type ProjectUserRole string

const (
	ProjectOwner      ProjectUserRole = "owner"
	ProjectManager    ProjectUserRole = "manager"
	ProjectTaskRunner ProjectUserRole = "task_runner"
	ProjectGuest      ProjectUserRole = "guest"
)

type ProjectUserPermission int

const (
	CanRunProjectTasks ProjectUserPermission = 1 << iota
	CanUpdateProject
	CanManageProjectResources
	CanManageProjectUsers
)

var rolePermissions = map[ProjectUserRole]ProjectUserPermission{
	ProjectOwner:      CanRunProjectTasks | CanUpdateProject | CanManageProjectResources,
	ProjectManager:    CanRunProjectTasks | CanManageProjectResources,
	ProjectTaskRunner: CanRunProjectTasks,
	ProjectGuest:      0,
}

func (r ProjectUserRole) IsValid() bool {
	_, ok := rolePermissions[r]
	return ok
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
