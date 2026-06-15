package projects

import (
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// workflowRunDetails is the JSON response for GET /runs/{run_id}.
type workflowRunDetails struct {
	Run   db.WorkflowRun        `json:"run"`
	Nodes []workflowNodeRunStatus `json:"nodes"`
}

// workflowNodeRunStatus pairs a workflow node with its optional task and
// approval status for a specific run.
type workflowNodeRunStatus struct {
	Node     db.WorkflowNode      `json:"node"`
	Task     *db.Task             `json:"task,omitempty"`
	Approval *db.WorkflowApproval `json:"approval,omitempty"`
}

// workflowApprovalResolveRequest is the body sent to the resolve endpoint.
type workflowApprovalResolveRequest struct {
	Status db.WorkflowApprovalStatus `json:"status"`
}

// workflowController implements the workflow HTTP API backed by the injected
// workflowRepo and workflowService.
type workflowController struct {
	workflowRepo    db.WorkflowManager
	workflowService pro_interfaces.WorkflowService
}

// NewWorkflowController creates a new workflowController.
func NewWorkflowController(svc pro_interfaces.WorkflowService, workflowRepo db.WorkflowManager) pro_interfaces.WorkflowController {
	return &workflowController{
		workflowRepo:    workflowRepo,
		workflowService: svc,
	}
}

// GetWorkflows returns all workflow templates for the current project.
func (c *workflowController) GetWorkflows(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflows, err := c.workflowRepo.GetWorkflowTemplates(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, workflows)
}

// AddWorkflow creates a new workflow template.
func (c *workflowController) AddWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var workflow db.WorkflowTemplate
	if !helpers.Bind(w, r, &workflow) {
		return
	}
	workflow.ProjectID = project.ID

	created, err := c.workflowRepo.CreateWorkflowTemplate(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, created)
}

// GetWorkflow returns the workflow template from the request context.
func (c *workflowController) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	helpers.WriteJSON(w, http.StatusOK, workflow)
}

// UpdateWorkflow replaces a workflow template's definition.
func (c *workflowController) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	existing := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	var workflow db.WorkflowTemplate
	if !helpers.Bind(w, r, &workflow) {
		return
	}
	workflow.ID = existing.ID
	workflow.ProjectID = project.ID

	if err := c.workflowRepo.UpdateWorkflowTemplate(workflow); err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveWorkflow deletes a workflow template.
func (c *workflowController) RemoveWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	if err := c.workflowRepo.DeleteWorkflowTemplate(project.ID, workflow.ID); err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RunWorkflow starts a new workflow run.
func (c *workflowController) RunWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	now := tz.Now()
	run, err := c.workflowRepo.CreateWorkflowRun(db.WorkflowRun{
		ProjectID:          project.ID,
		WorkflowTemplateID: workflow.ID,
		Status:             db.WorkflowRunRunning,
		Start:              &now,
	})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, run)
}

// StopWorkflowRun stops a running workflow run.
func (c *workflowController) StopWorkflowRun(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// GetWorkflowRuns returns all runs for the current workflow.
func (c *workflowController) GetWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	runs, err := c.workflowRepo.GetWorkflowRuns(project.ID, workflow.ID, db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, runs)
}

// GetWorkflowRun returns the details of a single workflow run.
func (c *workflowController) GetWorkflowRun(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	run := helpers.GetFromContext(r, "workflow_run").(db.WorkflowRun)

	approvals, err := c.workflowRepo.GetWorkflowApprovals(project.ID, run.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	approvalByNode := make(map[int]*db.WorkflowApproval, len(approvals))
	for i := range approvals {
		approvalByNode[approvals[i].WorkflowNodeID] = &approvals[i]
	}

	nodes := make([]workflowNodeRunStatus, 0, len(workflow.Nodes))
	for _, node := range workflow.Nodes {
		status := workflowNodeRunStatus{Node: node}
		if approval, ok := approvalByNode[node.ID]; ok {
			status.Approval = approval
		}
		nodes = append(nodes, status)
	}

	helpers.WriteJSON(w, http.StatusOK, workflowRunDetails{Run: run, Nodes: nodes})
}

// GetWorkflowRunArtifacts returns the merged artifacts of a workflow run.
func (c *workflowController) GetWorkflowRunArtifacts(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, map[string]any{})
}

// GetWorkflowApprovals returns all approvals for the current workflow run.
func (c *workflowController) GetWorkflowApprovals(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	run := helpers.GetFromContext(r, "workflow_run").(db.WorkflowRun)

	approvals, err := c.workflowRepo.GetWorkflowApprovals(project.ID, run.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, approvals)
}

// ResolveWorkflowApproval approves or rejects a pending workflow approval.
func (c *workflowController) ResolveWorkflowApproval(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	run := helpers.GetFromContext(r, "workflow_run").(db.WorkflowRun)

	nodeID, err := helpers.GetIntParam("node_id", w, r)
	if err != nil {
		return
	}

	var req workflowApprovalResolveRequest
	if !helpers.Bind(w, r, &req) {
		return
	}

	approval, err := c.workflowRepo.GetWorkflowApproval(project.ID, run.ID, nodeID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	now := time.Now()
	user := helpers.UserFromContext(r)
	approval.Status = req.Status
	approval.Resolved = &now
	approval.ResolvedByUserID = &user.ID

	if _, resolveErr := c.workflowRepo.ResolveWorkflowApprovalIfPending(approval); resolveErr != nil {
		helpers.WriteError(w, resolveErr)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, approval)
}
