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
	// CanReadProject on a project shows the whole project, on a grant only that object.
	CanReadProject
)

// The verbs a grant on a template or on a workflow may hold.
const (
	TemplateGrantPermissions = CanRunProjectTasks | CanManageProjectResources | CanReadProject
	WorkflowGrantPermissions = CanRunWorkflows | CanEditWorkflows | CanApproveWorkflows | CanReadProject
)

// LiftPermissions gives a mask written before the workflow verbs the same bits migration 2.20.17 gives
// stored roles: run tasks → run and approve workflows, manage resources → edit workflows.
func LiftPermissions(mask ProjectUserPermission) ProjectUserPermission {
	if mask&CanRunProjectTasks != 0 {
		mask |= CanRunWorkflows | CanApproveWorkflows
	}
	if mask&CanManageProjectResources != 0 {
		mask |= CanEditWorkflows
	}
	return mask
}

type ProjectUser struct {
	ProjectID int `db:"project_id" json:"project_id"`
	UserID    int `db:"user_id" json:"user_id"`
}
