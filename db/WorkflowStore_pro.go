package db

type WorkflowManager interface {
	GetWorkflowRunTasks(projectID int, runID int, params RetrieveQueryParams) ([]TaskWithTpl, error)

	GetWorkflowTemplates(projectID int, params RetrieveQueryParams) ([]WorkflowTemplate, error)
	GetWorkflowTemplate(projectID int, workflowID int) (WorkflowTemplate, error)
	CreateWorkflowTemplate(workflow WorkflowTemplate) (WorkflowTemplate, error)
	UpdateWorkflowTemplate(workflow WorkflowTemplate) error
	DeleteWorkflowTemplate(projectID int, workflowID int) error

	// GetWorkflowRevisions lists the surviving revisions of a template, newest
	// first, with HasRuns filled.
	GetWorkflowRevisions(projectID int, workflowID int) ([]WorkflowRevision, error)
	// GetWorkflowRevisionGraph returns the template with the nodes and edges of
	// the given revision instead of the latest one. Runs read their graph here.
	GetWorkflowRevisionGraph(projectID int, revisionID int) (WorkflowTemplate, error)

	GetWorkflowRuns(projectID int, workflowTemplateID int, params RetrieveQueryParams) ([]WorkflowRun, error)
	GetWorkflowRun(projectID int, workflowTemplateID int, runID int) (WorkflowRun, error)
	GetWorkflowRunByID(projectID int, runID int) (WorkflowRun, error)

	GetActiveWorkflowRuns() ([]WorkflowRun, error)
	CreateWorkflowRun(run WorkflowRun) (WorkflowRun, error)
	UpdateWorkflowRun(run WorkflowRun) error

	UpdateWorkflowRunStatusUnless(run WorkflowRun, excluded []WorkflowRunStatus) (bool, error)

	SetWorkflowRunRootTask(projectID int, runID int, taskID int) (bool, error)

	GetWorkflowApprovals(projectID int, runID int) ([]WorkflowApproval, error)
	GetWorkflowApproval(projectID int, runID int, nodeID int) (WorkflowApproval, error)
	CreateWorkflowApproval(approval WorkflowApproval) (WorkflowApproval, error)
	UpdateWorkflowApproval(approval WorkflowApproval) error
	ResolveWorkflowApprovalIfPending(approval WorkflowApproval) (bool, error)

	GetWorkflowDelays(projectID int, runID int) ([]WorkflowDelay, error)
	GetWorkflowDelay(projectID int, runID int, nodeID int) (WorkflowDelay, error)
	CreateWorkflowDelay(delay WorkflowDelay) (WorkflowDelay, error)
	UpdateWorkflowDelay(delay WorkflowDelay) error
	ResolveWorkflowDelayIfWaiting(delay WorkflowDelay) (bool, error)
	GetExpiredWorkflowDelays() ([]WorkflowDelay, error)
}
