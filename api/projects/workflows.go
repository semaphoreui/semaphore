package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

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

func GetWorkflows(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	workflows, err := helpers.Store(r).GetWorkflowTemplates(project.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, workflows)
}

func AddWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var workflow db.WorkflowTemplate
	if !helpers.Bind(w, r, &workflow) {
		return
	}

	workflow.ProjectID = project.ID

	newWorkflow, err := helpers.Store(r).CreateWorkflowTemplate(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventWorkflow,
		ObjectID:    newWorkflow.ID,
		Description: fmt.Sprintf("Workflow ID %d created", newWorkflow.ID),
	})

	helpers.WriteJSON(w, http.StatusCreated, newWorkflow)
}

func GetWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	helpers.WriteJSON(w, http.StatusOK, workflow)
}

func UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	oldWorkflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	var workflow db.WorkflowTemplate
	if !helpers.Bind(w, r, &workflow) {
		return
	}

	if workflow.ID != oldWorkflow.ID {
		helpers.WriteErrorStatus(w, "workflow id in URL and in body must be the same", http.StatusBadRequest)
		return
	}

	if workflow.ProjectID != oldWorkflow.ProjectID {
		helpers.WriteErrorStatus(w, "you can not move workflow to other project", http.StatusBadRequest)
		return
	}

	err := helpers.Store(r).UpdateWorkflowTemplate(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldWorkflow.ProjectID,
		ObjectType:  db.EventWorkflow,
		ObjectID:    oldWorkflow.ID,
		Description: fmt.Sprintf("Workflow ID %d updated", workflow.ID),
	})

	w.WriteHeader(http.StatusNoContent)
}

func RemoveWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	err := helpers.Store(r).DeleteWorkflowTemplate(workflow.ProjectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   workflow.ProjectID,
		ObjectType:  db.EventWorkflow,
		ObjectID:    workflow.ID,
		Description: fmt.Sprintf("Workflow ID %d deleted", workflow.ID),
	})

	w.WriteHeader(http.StatusNoContent)
}

func RunWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	user := helpers.UserFromContext(r)

	run, err := taskPool(r).StartWorkflow(workflow, user)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, run)
}

func GetWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)

	runs, err := helpers.Store(r).GetWorkflowRuns(project.ID, workflow.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, runs)
}

type workflowNodeRunStatus struct {
	Node     db.WorkflowNode      `json:"node"`
	Task     *db.TaskWithTpl      `json:"task,omitempty"`
	Approval *db.WorkflowApproval `json:"approval,omitempty"`
}

type workflowRunDetails struct {
	Run   db.WorkflowRun          `json:"run"`
	Nodes []workflowNodeRunStatus `json:"nodes"`
}

func GetWorkflowRun(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	run := helpers.GetFromContext(r, "workflow_run").(db.WorkflowRun)
	err := taskPool(r).ProgressWorkflowRun(project.ID, run.ID, helpers.UserFromContext(r))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	run, err = helpers.Store(r).GetWorkflowRunByID(project.ID, run.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	tasks, err := helpers.Store(r).GetProjectTasks(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	taskByNodeID := make(map[int]db.TaskWithTpl)
	for _, task := range tasks {
		if task.WorkflowRunID == nil || *task.WorkflowRunID != run.ID || task.WorkflowNodeID == nil {
			continue
		}
		if _, exists := taskByNodeID[*task.WorkflowNodeID]; exists {
			continue
		}
		taskByNodeID[*task.WorkflowNodeID] = task
	}

	approvals, err := helpers.Store(r).GetWorkflowApprovals(project.ID, run.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	approvalByNodeID := make(map[int]db.WorkflowApproval, len(approvals))
	for _, approval := range approvals {
		approvalByNodeID[approval.WorkflowNodeID] = approval
	}

	nodeStatuses := make([]workflowNodeRunStatus, 0, len(workflow.Nodes))
	for _, node := range workflow.Nodes {
		ns := workflowNodeRunStatus{Node: node}
		if task, ok := taskByNodeID[node.ID]; ok {
			t := task
			ns.Task = &t
		}
		if approval, ok := approvalByNodeID[node.ID]; ok {
			a := approval
			ns.Approval = &a
		}
		nodeStatuses = append(nodeStatuses, ns)
	}

	helpers.WriteJSON(w, http.StatusOK, workflowRunDetails{
		Run:   run,
		Nodes: nodeStatuses,
	})
}

func GetWorkflowApprovals(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	run := helpers.GetFromContext(r, "workflow_run").(db.WorkflowRun)
	err := taskPool(r).ProgressWorkflowRun(project.ID, run.ID, helpers.UserFromContext(r))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	approvals, err := helpers.Store(r).GetWorkflowApprovals(project.ID, run.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, approvals)
}

type workflowApprovalResolutionRequest struct {
	Status db.WorkflowApprovalStatus `json:"status"`
}

func ResolveWorkflowApproval(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.WorkflowTemplate)
	run := helpers.GetFromContext(r, "workflow_run").(db.WorkflowRun)
	user := helpers.UserFromContext(r)
	nodeID, err := helpers.GetIntParam("node_id", w, r)
	if err != nil {
		return
	}

	var req workflowApprovalResolutionRequest
	if !helpers.Bind(w, r, &req) {
		return
	}

	approval, err := taskPool(r).ResolveWorkflowApproval(project.ID, workflow.ID, run.ID, nodeID, req.Status, user)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      user.ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventWorkflow,
		ObjectID:    workflow.ID,
		Description: fmt.Sprintf("Workflow run #%d approval for node #%d set to %s", run.ID, nodeID, req.Status),
	})

	helpers.WriteJSON(w, http.StatusOK, approval)
}
