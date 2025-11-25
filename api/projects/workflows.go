package projects

import (
	"encoding/json"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/services/workflows"
)

type WorkflowController struct {
	engine *workflows.WorkflowEngine
}

func NewWorkflowController(store db.Store, taskPool *tasks.TaskPool) *WorkflowController {
	engine := workflows.NewWorkflowEngine(store, taskPool)
	return &WorkflowController{
		engine: engine,
	}
}

// GetWorkflows returns all workflows for a project
func GetWorkflows(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	workflows, err := helpers.Store(r).GetWorkflows(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, workflows)
}

// AddWorkflow creates a new workflow
func AddWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var workflow db.Workflow
	if err := helpers.Bind(w, r, &workflow); err != nil {
		return
	}

	workflow.ProjectID = project.ID

	newWorkflow, err := helpers.Store(r).CreateWorkflow(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newWorkflow)
}

// GetWorkflow returns a workflow by ID
func GetWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	// Get nodes and links
	nodes, err := helpers.Store(r).GetWorkflowNodes(workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	links, err := helpers.Store(r).GetWorkflowLinks(workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	workflowWithNodes := db.WorkflowWithNodes{
		Workflow: workflow,
		Nodes:    nodes,
		Links:    links,
	}

	helpers.WriteJSON(w, http.StatusOK, workflowWithNodes)
}

// UpdateWorkflow updates a workflow
func UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	var updatedWorkflow db.Workflow
	if err := helpers.Bind(w, r, &updatedWorkflow); err != nil {
		return
	}

	updatedWorkflow.ID = workflow.ID
	updatedWorkflow.ProjectID = project.ID

	if err := helpers.Store(r).UpdateWorkflow(updatedWorkflow); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, updatedWorkflow)
}

// DeleteWorkflow deletes a workflow
func DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	if err := helpers.Store(r).DeleteWorkflow(project.ID, workflow.ID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RunWorkflow starts a workflow execution
func (c *WorkflowController) RunWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	user := helpers.GetFromContext(r, "user").(*db.User)

	run, err := c.engine.RunWorkflow(project.ID, workflow.ID, &user.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, run)
}

// GetWorkflowRuns returns all runs for a workflow
func GetWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	runs, err := helpers.Store(r).GetWorkflowRuns(workflow.ID, db.RetrieveQueryParams{
		Count: 100,
	})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, runs)
}

// GetWorkflowRun returns a workflow run by ID
func GetWorkflowRun(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	runID, err := helpers.GetIntParam("run_id", w, r)
	if err != nil {
		return
	}

	run, err := helpers.Store(r).GetWorkflowRun(workflow.ID, runID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Get run nodes
	nodes, err := helpers.Store(r).GetWorkflowRunNodes(run.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	runWithNodes := db.WorkflowRunWithNodes{
		WorkflowRun: run,
		Nodes:       nodes,
	}

	helpers.WriteJSON(w, http.StatusOK, runWithNodes)
}

// WorkflowMiddleware gets a workflow by ID and sets it in context
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

// UpdateWorkflowNodes updates workflow nodes and links
func UpdateWorkflowNodes(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	var payload struct {
		Nodes []db.WorkflowNode `json:"nodes"`
		Links []db.WorkflowLink `json:"links"`
	}

	if err := helpers.Bind(w, r, &payload); err != nil {
		return
	}

	store := helpers.Store(r)

	// Get existing nodes and links
	existingNodes, err := store.GetWorkflowNodes(workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	existingLinks, err := store.GetWorkflowLinks(workflow.ID)
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
	for _, node := range payload.Nodes {
		node.WorkflowID = workflow.ID
		if node.ID > 0 && existingNodeMap[node.ID] {
			// Update existing node
			if err := store.UpdateWorkflowNode(node); err != nil {
				helpers.WriteError(w, err)
				return
			}
		} else {
			// Create new node
			node.ID = 0
			_, err := store.CreateWorkflowNode(node)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}

	// Delete nodes that are no longer in the payload
	for _, existingNode := range existingNodes {
		found := false
		for _, node := range payload.Nodes {
			if node.ID == existingNode.ID {
				found = true
				break
			}
		}
		if !found {
			if err := store.DeleteWorkflowNode(workflow.ID, existingNode.ID); err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}

	// Delete all existing links
	for _, link := range existingLinks {
		if err := store.DeleteWorkflowLink(workflow.ID, link.ID); err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	// Create new links
	for _, link := range payload.Links {
		link.WorkflowID = workflow.ID
		link.ID = 0
		_, err := store.CreateWorkflowLink(link)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	// Return updated workflow with nodes and links
	nodes, _ := store.GetWorkflowNodes(workflow.ID)
	links, _ := store.GetWorkflowLinks(workflow.ID)

	workflowWithNodes := db.WorkflowWithNodes{
		Workflow: workflow,
		Nodes:    nodes,
		Links:    links,
	}

	helpers.WriteJSON(w, http.StatusOK, workflowWithNodes)
}
