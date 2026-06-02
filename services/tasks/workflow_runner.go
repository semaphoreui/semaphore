package tasks

import (
	"fmt"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func (p *TaskPool) StartWorkflow(workflow db.WorkflowTemplate, user *db.User) (run db.WorkflowRun, err error) {
	p.workflowMu.Lock()
	defer p.workflowMu.Unlock()

	if err = db.ValidateWorkflowTemplate(p.store, workflow); err != nil {
		return
	}

	rootNode, err := db.WorkflowRootNode(workflow)
	if err != nil {
		return run, db.NewValidationError(err.Error())
	}

	now := time.Now().UTC()
	run, err = p.store.CreateWorkflowRun(db.WorkflowRun{
		ProjectID:          workflow.ProjectID,
		WorkflowTemplateID: workflow.ID,
		Status:             db.WorkflowRunRunning,
		Start:              &now,
	})
	if err != nil {
		return
	}

	tpl, err := p.store.GetTemplate(workflow.ProjectID, rootNode.TemplateID)
	if err != nil {
		return
	}

	var userID *int
	var username string
	if user != nil {
		userID = &user.ID
		username = user.Username
	}

	createdTask, err := p.AddTask(db.Task{
		TemplateID:     rootNode.TemplateID,
		ProjectID:      workflow.ProjectID,
		WorkflowRunID:  &run.ID,
		WorkflowNodeID: &rootNode.ID,
	}, userID, username, workflow.ProjectID, tpl.App.NeedTaskAlias())
	if err != nil {
		return
	}

	run.RootTaskID = &createdTask.ID
	err = p.store.UpdateWorkflowRun(run)
	return
}

func (p *TaskPool) getWorkflowRunTasks(projectID int, runID int) (tasks []db.TaskWithTpl, err error) {
	allTasks, err := p.store.GetProjectTasks(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	tasks = make([]db.TaskWithTpl, 0, len(allTasks))
	for _, task := range allTasks {
		if task.WorkflowRunID == nil || *task.WorkflowRunID != runID {
			continue
		}
		tasks = append(tasks, task)
	}

	return
}

func mapLatestNodeTaskStatus(tasks []db.TaskWithTpl) map[int]task_logger.TaskStatus {
	// GetProjectTasks returns tasks sorted by ID descending, so the first seen task
	// for each node is the latest one for that node.
	statusByNodeID := make(map[int]task_logger.TaskStatus)
	for _, task := range tasks {
		if task.WorkflowNodeID == nil {
			continue
		}
		if _, exists := statusByNodeID[*task.WorkflowNodeID]; exists {
			continue
		}
		statusByNodeID[*task.WorkflowNodeID] = task.Status
	}

	return statusByNodeID
}

func mapLatestNodeTask(tasks []db.TaskWithTpl) map[int]db.TaskWithTpl {
	// GetProjectTasks returns tasks sorted by ID descending, so the first seen task
	// for each node is the latest one for that node.
	taskByNodeID := make(map[int]db.TaskWithTpl)
	for _, task := range tasks {
		if task.WorkflowNodeID == nil {
			continue
		}
		if _, exists := taskByNodeID[*task.WorkflowNodeID]; exists {
			continue
		}
		taskByNodeID[*task.WorkflowNodeID] = task
	}

	return taskByNodeID
}

func (p *TaskPool) isWorkflowNodeReady(workflow db.WorkflowTemplate, destinationNodeID int, statusByNodeID map[int]task_logger.TaskStatus) (ready bool, blocked bool) {
	hasInbound := false
	for _, edge := range workflow.Edges {
		if edge.DestinationNodeID != destinationNodeID {
			continue
		}
		hasInbound = true

		status, exists := statusByNodeID[edge.SourceNodeID]
		if !exists || !status.IsFinished() {
			return false, false
		}
		if !db.WorkflowConditionMatches(status, edge.Condition) {
			return false, true
		}
	}

	if !hasInbound {
		return false, true
	}

	return true, false
}

func (p *TaskPool) updateWorkflowRunStatus(run db.WorkflowRun, workflowTasks []db.TaskWithTpl) error {
	if len(workflowTasks) == 0 {
		return nil
	}

	hasUnfinished := false
	hasFailed := false
	for _, task := range workflowTasks {
		if !task.Status.IsFinished() {
			hasUnfinished = true
			continue
		}
		if task.Status != task_logger.TaskSuccessStatus {
			hasFailed = true
		}
	}

	var runStatus db.WorkflowRunStatus
	switch {
	case hasUnfinished:
		runStatus = db.WorkflowRunRunning
	case hasFailed:
		runStatus = db.WorkflowRunFailed
	default:
		runStatus = db.WorkflowRunSuccess
	}

	if run.Status == runStatus {
		if runStatus != db.WorkflowRunRunning || run.End != nil {
			return nil
		}
	}

	run.Status = runStatus
	if runStatus == db.WorkflowRunRunning {
		run.End = nil
	} else if run.End == nil {
		now := time.Now().UTC()
		run.End = &now
	}

	return p.store.UpdateWorkflowRun(run)
}

func (p *TaskPool) HandleWorkflowTaskCompletion(task db.Task) error {
	if task.WorkflowRunID == nil || task.WorkflowNodeID == nil {
		return nil
	}

	p.workflowMu.Lock()
	defer p.workflowMu.Unlock()

	run, err := p.store.GetWorkflowRunByID(task.ProjectID, *task.WorkflowRunID)
	if err != nil {
		return err
	}

	workflow, err := p.store.GetWorkflowTemplate(task.ProjectID, run.WorkflowTemplateID)
	if err != nil {
		return err
	}

	workflowTasks, err := p.getWorkflowRunTasks(task.ProjectID, run.ID)
	if err != nil {
		return err
	}

	nodeStatusByID := mapLatestNodeTaskStatus(workflowTasks)
	nodeTaskByID := mapLatestNodeTask(workflowTasks)

	for _, edge := range workflow.Edges {
		if edge.SourceNodeID != *task.WorkflowNodeID {
			continue
		}
		if !db.WorkflowConditionMatches(task.Status, edge.Condition) {
			continue
		}
		if _, exists := nodeTaskByID[edge.DestinationNodeID]; exists {
			continue
		}

		ready, blocked := p.isWorkflowNodeReady(workflow, edge.DestinationNodeID, nodeStatusByID)
		if !ready || blocked {
			continue
		}

		var destinationNode *db.WorkflowNode
		for i := range workflow.Nodes {
			if workflow.Nodes[i].ID == edge.DestinationNodeID {
				destinationNode = &workflow.Nodes[i]
				break
			}
		}
		if destinationNode == nil {
			continue
		}

		tpl, err2 := p.store.GetTemplate(task.ProjectID, destinationNode.TemplateID)
		if err2 != nil {
			return err2
		}

		_, err2 = p.AddTask(db.Task{
			TemplateID:     destinationNode.TemplateID,
			ProjectID:      task.ProjectID,
			BuildTaskID:    &task.ID,
			WorkflowRunID:  task.WorkflowRunID,
			WorkflowNodeID: &destinationNode.ID,
		}, task.UserID, "", task.ProjectID, tpl.App.NeedTaskAlias())
		if err2 != nil {
			return fmt.Errorf("failed to enqueue workflow node %d: %w", destinationNode.ID, err2)
		}
	}

	workflowTasks, err = p.getWorkflowRunTasks(task.ProjectID, run.ID)
	if err != nil {
		return err
	}

	return p.updateWorkflowRunStatus(run, workflowTasks)
}
