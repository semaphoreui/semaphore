package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// The workflow HTTP handlers are a Pro feature and live in the pro module
// (pro/api/projects + pro_impl/api/projects), wired via pro_interfaces.WorkflowController.
// Only the request-scoped context loaders remain here, since they rely on the
// open db.Store and are shared by the route middleware chain.

func WorkflowsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		workflowID, err := helpers.GetIntParam("workflow_id", w, r)
		if err != nil {
			return
		}

		workflow, err := helpers.Store(r).GetWorkflowTemplate(project.ID, workflowID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "workflow", workflow)
		next.ServeHTTP(w, r)
	})
}

func WorkflowRunsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

		runID, err := helpers.GetIntParam("run_id", w, r)
		if err != nil {
			return
		}

		run, err := helpers.Store(r).GetWorkflowRun(project.ID, workflow.ID, runID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "workflow_run", run)
		next.ServeHTTP(w, r)
	})
}
