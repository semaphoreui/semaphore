package db

type WorkflowManager interface {
	// GetWorkflowRunTasks lists the tasks that make up the state of a run,
	// newest first: the tasks the run launched and, for a failed-only re-run,
	// the tasks of the nodes it carries over from earlier runs.
	GetWorkflowRunTasks(projectID int, runID int, params RetrieveQueryParams) ([]TaskWithTpl, error)

	// GetWorkflowTemplates lists the workflows the viewer may see; a nil viewer lists them all.
	GetWorkflowTemplates(projectID int, viewer *Viewer, params RetrieveQueryParams) ([]WorkflowTemplate, error)
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
	// CreateWorkflowRerun creates a run together with the nodes it carries
	// over, in one transaction: a progression pass on any node of a cluster
	// sees either no run or the run with every carried node in place, never a
	// run whose carried nodes look not started.
	CreateWorkflowRerun(run WorkflowRun, carries []WorkflowRunCarry) (WorkflowRun, error)
	// GetWorkflowRunCarries lists the nodes the run carries over, by node id.
	GetWorkflowRunCarries(projectID int, runID int) ([]WorkflowRunCarry, error)
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
