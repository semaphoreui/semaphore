package projects

import (
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/workflows"
)

var workflowEngine *workflows.WorkflowEngine

// InitWorkflowEngine initializes the workflow engine
func InitWorkflowEngine(engine *workflows.WorkflowEngine) {
	workflowEngine = engine
}

// WorkflowMiddleware loads workflow into context
func WorkflowMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		workflowID, err := helpers.GetIntParam("workflow_id", w, r)
		if err != nil {
			return
		}

		workflow, err := helpers.Store(r).GetWorkflow(project.ID, workflowID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "workflow", workflow)
		next.ServeHTTP(w, r)
	})
}

// GetWorkflows returns all workflows for a project
func GetWorkflows(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	//params, err := helpers.GetQueryParams(r.URL.Query(), db.WorkflowProps)
	//if err != nil {
	//	helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
	//		"error": err.Error(),
	//	})
	//	return
	//}

	workflows, err := helpers.Store(r).GetWorkflows(project.ID, db.RetrieveQueryParams{})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, workflows)
}

// GetWorkflow returns a specific workflow
func GetWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	helpers.WriteJSON(w, http.StatusOK, workflow)
}

// CreateWorkflow creates a new workflow
func CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var workflow db.Workflow
	if !helpers.Bind(w, r, &workflow) {
		return
	}

	workflow.ProjectID = project.ID

	newWorkflow, err := helpers.Store(r).CreateWorkflow(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Create nodes if provided
	if len(workflow.Nodes) > 0 {
		for _, node := range workflow.Nodes {
			node.WorkflowID = newWorkflow.ID
			_, err := helpers.Store(r).CreateWorkflowNode(node)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}

	// Create links if provided
	if len(workflow.Links) > 0 {
		for _, link := range workflow.Links {
			link.WorkflowID = newWorkflow.ID
			_, err := helpers.Store(r).CreateWorkflowLink(link)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}

	// Reload workflow with nodes and links
	workflow, err = helpers.Store(r).GetWorkflow(project.ID, newWorkflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, workflow)
}

// UpdateWorkflow updates an existing workflow
func UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	var body db.Workflow
	if !helpers.Bind(w, r, &body) {
		return
	}

	if body.ID != workflow.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Workflow ID mismatch",
		})
		return
	}

	body.ProjectID = workflow.ProjectID

	err := helpers.Store(r).UpdateWorkflow(body)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Update nodes if provided
	if len(body.Nodes) > 0 {
		// Get existing nodes
		existingNodes, err := helpers.Store(r).GetWorkflowNodes(workflow.ID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		// Create a map of existing node IDs
		existingNodeMap := make(map[int]bool)
		for _, node := range existingNodes {
			existingNodeMap[node.ID] = true
		}

		// Update or create nodes
		providedNodeMap := make(map[int]bool)
		for _, node := range body.Nodes {
			node.WorkflowID = workflow.ID
			providedNodeMap[node.ID] = true

			if node.ID > 0 && existingNodeMap[node.ID] {
				// Update existing node
				err := helpers.Store(r).UpdateWorkflowNode(node)
				if err != nil {
					helpers.WriteError(w, err)
					return
				}
			} else {
				// Create new node
				_, err := helpers.Store(r).CreateWorkflowNode(node)
				if err != nil {
					helpers.WriteError(w, err)
					return
				}
			}
		}

		// Delete nodes that are not in the provided list
		for _, node := range existingNodes {
			if !providedNodeMap[node.ID] {
				err := helpers.Store(r).DeleteWorkflowNode(workflow.ID, node.ID)
				if err != nil {
					helpers.WriteError(w, err)
					return
				}
			}
		}
	}

	// Update links if provided
	if len(body.Links) > 0 {
		// Get existing links
		existingLinks, err := helpers.Store(r).GetWorkflowLinks(workflow.ID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		// Delete all existing links (simpler approach for MVP)
		for _, link := range existingLinks {
			err := helpers.Store(r).DeleteWorkflowLink(workflow.ID, link.ID)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}

		// Create new links
		for _, link := range body.Links {
			link.WorkflowID = workflow.ID
			_, err := helpers.Store(r).CreateWorkflowLink(link)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteWorkflow deletes a workflow
func DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	err := helpers.Store(r).DeleteWorkflow(workflow.ProjectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RunWorkflow starts a workflow execution
func RunWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	user := helpers.GetFromContext(r, "user").(*db.User)

	// Create workflow run
	run := db.WorkflowRun{
		WorkflowID: workflow.ID,
		ProjectID:  workflow.ProjectID,
		UserID:     &user.ID,
		Status:     db.WorkflowRunStatusPending,
	}

	newRun, err := helpers.Store(r).CreateWorkflowRun(run)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Start workflow execution asynchronously
	if workflowEngine != nil {
		err = workflowEngine.ExecuteWorkflow(newRun.ID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	helpers.WriteJSON(w, http.StatusCreated, newRun)
}

// GetWorkflowRuns returns all runs for a workflow
func GetWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	//params, err := helpers.GetQueryParams(r.URL.Query(), db.WorkflowRunProps)
	//if err != nil {
	//	helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
	//		"error": err.Error(),
	//	})
	//	return
	//}

	runs, err := helpers.Store(r).GetWorkflowRuns(workflow.ID, db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, runs)
}

// GetProjectWorkflowRuns returns all workflow runs for a project
func GetProjectWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	//params, err := helpers.GetQueryParams(r.URL.Query(), db.WorkflowRunProps)
	//if err != nil {
	//	helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
	//		"error": err.Error(),
	//	})
	//	return
	//}

	runs, err := helpers.Store(r).GetProjectWorkflowRuns(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, runs)
}

// GetWorkflowRun returns a specific workflow run with node runs
func GetWorkflowRun(w http.ResponseWriter, r *http.Request) {
	workflowRunID, err := helpers.GetIntParam("run_id", w, r)
	if err != nil {
		return
	}

	run, err := helpers.Store(r).GetWorkflowRun(workflowRunID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Get node runs
	nodeRuns, err := helpers.Store(r).GetWorkflowNodeRuns(workflowRunID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	response := map[string]interface{}{
		"run":       run,
		"node_runs": nodeRuns,
	}

	helpers.WriteJSON(w, http.StatusOK, response)
}

// StopWorkflowRun stops a running workflow
func StopWorkflowRun(w http.ResponseWriter, r *http.Request) {
	workflowRunID, err := helpers.GetIntParam("run_id", w, r)
	if err != nil {
		return
	}

	run, err := helpers.Store(r).GetWorkflowRun(workflowRunID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if run.Status != db.WorkflowRunStatusRunning && run.Status != db.WorkflowRunStatusPending {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Workflow is not running",
		})
		return
	}

	// Update workflow run status to stopped
	run.Status = db.WorkflowRunStatusStopped
	now := db.GetParsedTime(time.Now())
	run.End = &now

	err = helpers.Store(r).UpdateWorkflowRun(run)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Stop all running node runs
	nodeRuns, err := helpers.Store(r).GetWorkflowNodeRuns(workflowRunID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	for _, nodeRun := range nodeRuns {
		if nodeRun.Status == db.WorkflowNodeRunStatusRunning || nodeRun.Status == db.WorkflowNodeRunStatusPending {
			nodeRun.Status = db.WorkflowNodeRunStatusStopped
			nodeRun.End = &now
			err = helpers.Store(r).UpdateWorkflowNodeRun(nodeRun)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}

	// Stop the workflow through the engine
	if workflowEngine != nil {
		err = workflowEngine.StopWorkflow(workflowRunID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
