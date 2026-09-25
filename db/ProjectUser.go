package db

type ProjectUserRole string

const (
	ProjectOwner      ProjectUserRole = "owner"
	ProjectManager    ProjectUserRole = "manager"
	ProjectTaskRunner ProjectUserRole = "task_runner"
	ProjectGuest      ProjectUserRole = "guest"
	ProjectNone       ProjectUserRole = ""
)

type ProjectUserPermission int64

const (
	CanRunProjectTasks ProjectUserPermission = 1 << iota
	CanUpdateProject
	CanManageProjectResources
	CanManageProjectUsers
)

// IsBuiltin reports whether the role uses a reserved built-in slug.
func (r ProjectUserRole) IsBuiltin() bool {
	switch r {
	case ProjectOwner, ProjectManager, ProjectTaskRunner, ProjectGuest:
		return true
	default:
		return false
	}
}

type ProjectUser struct {
	ID        int             `db:"id" json:"-"`
	ProjectID int             `db:"project_id" json:"project_id"`
	UserID    int             `db:"user_id" json:"user_id"`
	Role      ProjectUserRole `db:"role" json:"role"`
}
