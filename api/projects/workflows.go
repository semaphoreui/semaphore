package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// WorkflowController loads workflow-related entities into the request context
// for the workflow route middleware chain.
type WorkflowController struct {
	workflowRepo db.WorkflowManager
}

func NewWorkflowController(workflowRepo db.WorkflowManager) *WorkflowController {
	return &WorkflowController{
		workflowRepo: workflowRepo,
	}
}

func (c *WorkflowController) WorkflowsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		workflowID, ok := helpers.GetIntParam("workflow_id", w, r)
		if !ok {
			return
		}

		workflow, err := c.workflowRepo.GetWorkflowTemplate(project.ID, workflowID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "workflow", workflow)
		next.ServeHTTP(w, r)
	})
}

func (c *WorkflowController) WorkflowRunsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

		runID, ok := helpers.GetIntParam("run_id", w, r)
		if !ok {
			return
		}

		run, err := c.workflowRepo.GetWorkflowRun(project.ID, workflow.ID, runID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "workflow_run", run)
		next.ServeHTTP(w, r)
	})
}
