package db

type ProjectUserRole string

const (
	ProjectOwner      ProjectUserRole = "owner"
	ProjectManager    ProjectUserRole = "manager"
	ProjectTaskRunner ProjectUserRole = "task_runner"
	ProjectGuest      ProjectUserRole = "guest"
)

func (r ProjectUserRole) IsValid() bool {
	switch r {
	case ProjectOwner, ProjectManager, ProjectTaskRunner, ProjectGuest:
		return true
	}
	return false
}

type ProjectUserPermission int64

const (
	CanRunProjectTasks ProjectUserPermission = 1 << iota
	CanUpdateProject
	CanManageProjectResources
	CanManageProjectUsers
	CanRunWorkflows
	CanEditWorkflows
	CanApproveWorkflows
)

// The verbs a grant on a template or on a workflow may hold.
const (
	TemplateGrantPermissions = CanRunProjectTasks | CanManageProjectResources
	WorkflowGrantPermissions = CanRunWorkflows | CanEditWorkflows | CanApproveWorkflows
)

type ProjectUser struct {
	ProjectID int `db:"project_id" json:"project_id"`
	UserID    int `db:"user_id" json:"user_id"`
}
