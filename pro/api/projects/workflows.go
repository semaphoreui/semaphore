package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// workflowController is the open-source stub. Workflows are a Pro feature; the
// real implementation lives in pro_impl/api/projects/workflows.go. The stub
// keeps the open build compiling and the API surface present (returning empty
// collections / 404) while the feature is disabled via the Workflows feature
// flag (see pro/pkg/features).
type workflowController struct{}

func NewWorkflowController(svc pro_interfaces.WorkflowService, workflowRepo db.WorkflowManager) pro_interfaces.WorkflowController {
	return &workflowController{}
}

func (c *workflowController) GetWorkflows(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []struct{}{})
}

func (c *workflowController) AddWorkflow(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) RemoveWorkflow(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) RunWorkflow(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) StopWorkflowRun(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) GetWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []struct{}{})
}

func (c *workflowController) GetWorkflowRun(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) GetWorkflowRunArtifacts(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (c *workflowController) GetWorkflowApprovals(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, []struct{}{})
}

func (c *workflowController) ResolveWorkflowApproval(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
