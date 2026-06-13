package db

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
}
