package pro_interfaces

import "github.com/semaphoreui/semaphore/db"

// WorkflowService orchestrates workflow runs: starting a run, progressing it as
// upstream tasks finish, resolving approvals and merging run artifacts. It is a
// Pro feature — the open-source build wires a no-op stub
// (pro/services/server/workflow_svc.go); the licensed build provides the real
// implementation (pro_impl/services/server/workflow_svc.go).
type WorkflowService interface {
	StartWorkflow(workflow db.WorkflowTemplate, user *db.User) (db.WorkflowRun, error)
	ProgressWorkflowRun(projectID int, runID int, user *db.User) error
	// StopWorkflowRun stops a non-finished run: it signals every in-flight task
	// of the run to stop and marks the run as stopped (terminal).
	StopWorkflowRun(projectID int, runID int, user *db.User) (db.WorkflowRun, error)
	ResolveWorkflowApproval(projectID int, workflowID int, runID int, nodeID int, status db.WorkflowApprovalStatus, user *db.User) (db.WorkflowApproval, error)
	HandleWorkflowTaskCompletion(task db.Task) error
	GetWorkflowRunArtifacts(projectID int, runID int, currentTaskID *int) (map[string]any, error)
}

// WorkflowTaskEnqueuer is the slice of the task pool the workflow service needs
// to launch a node's task. It is implemented by *services/tasks.TaskPool.
// Declaring it here keeps the pro modules dependent only on pro_interfaces + db,
// avoiding an import of (and a cycle with) the services/tasks package.
type WorkflowTaskEnqueuer interface {
	AddTask(task db.Task, userID *int, username string, projectID int, needAlias bool) (db.Task, error)
	// StopTasksByWorkflowRun stops every active (queued or running) task that
	// belongs to the given workflow run. forceStop kills running tasks
	// immediately instead of letting them stop gracefully.
	StopTasksByWorkflowRun(projectID int, runID int, forceStop bool)
}
