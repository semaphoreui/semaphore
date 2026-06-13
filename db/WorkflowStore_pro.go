package db

// WorkflowManager is the Pro workflow store. Its implementation lives in
// pro_impl (see pro_impl/db/sql/workflow.go) and is wired through
// factory.NewWorkflowStore; the open-source build gets the no-op stub in
// pro/db/sql/workflow.go. It is intentionally NOT embedded in Store, mirroring
// TerraformStore (see TerraformInventoryStore_pro.go).
type WorkflowManager interface {
	GetWorkflowRunTasks(projectID int, runID int, params RetrieveQueryParams) ([]TaskWithTpl, error)

	GetWorkflowTemplates(projectID int, params RetrieveQueryParams) ([]WorkflowTemplate, error)
	GetWorkflowTemplate(projectID int, workflowID int) (WorkflowTemplate, error)
	CreateWorkflowTemplate(workflow WorkflowTemplate) (WorkflowTemplate, error)
	UpdateWorkflowTemplate(workflow WorkflowTemplate) error
	DeleteWorkflowTemplate(projectID int, workflowID int) error

	GetWorkflowRuns(projectID int, workflowTemplateID int, params RetrieveQueryParams) ([]WorkflowRun, error)
	GetWorkflowRun(projectID int, workflowTemplateID int, runID int) (WorkflowRun, error)
	GetWorkflowRunByID(projectID int, runID int) (WorkflowRun, error)
	// GetActiveWorkflowRuns returns every run in a non-terminal status
	// (running/approval) across all projects, for the workflow reconciler.
	GetActiveWorkflowRuns() ([]WorkflowRun, error)
	CreateWorkflowRun(run WorkflowRun) (WorkflowRun, error)
	UpdateWorkflowRun(run WorkflowRun) error
	// UpdateWorkflowRunStatusUnless updates the run's status and end time only
	// when its current persisted status is not in excluded. Returns whether a
	// row was updated. Used as a compare-and-set so concurrent HA nodes cannot
	// revive a stopped run or overwrite each other's whole-row writes.
	UpdateWorkflowRunStatusUnless(run WorkflowRun, excluded []WorkflowRunStatus) (bool, error)
	// SetWorkflowRunRootTask sets the run's root task only when none is
	// recorded yet. Returns whether a row was updated.
	SetWorkflowRunRootTask(projectID int, runID int, taskID int) (bool, error)

	GetWorkflowApprovals(projectID int, runID int) ([]WorkflowApproval, error)
	GetWorkflowApproval(projectID int, runID int, nodeID int) (WorkflowApproval, error)
	CreateWorkflowApproval(approval WorkflowApproval) (WorkflowApproval, error)
	UpdateWorkflowApproval(approval WorkflowApproval) error
	// ResolveWorkflowApprovalIfPending atomically resolves the approval
	// (status/resolved/resolver taken from approval) only when it is still
	// pending. Returns whether the update won; a false result means another
	// actor (user or timeout) resolved it first.
	ResolveWorkflowApprovalIfPending(approval WorkflowApproval) (bool, error)
}
