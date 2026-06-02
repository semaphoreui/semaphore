package tasks

import (
	"errors"
	"fmt"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks/artifacts"
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

	var userID *int
	var username string
	if user != nil {
		userID = &user.ID
		username = user.Username
	}

	err = p.startWorkflowNode(run, rootNode, nil, userID, username)
	if err != nil {
		return
	}

	approvals, err := p.store.GetWorkflowApprovals(run.ProjectID, run.ID)
	if err != nil {
		return
	}

	workflowTasks, err := p.getWorkflowRunTasks(run.ProjectID, run.ID)
	if err != nil {
		return
	}
	run, err = p.store.GetWorkflowRunByID(run.ProjectID, run.ID)
	if err != nil {
		return
	}
	err = p.updateWorkflowRunStatus(run, workflowTasks, approvals)
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

// GetWorkflowRunArtifacts merges artifact JSON blobs from every finished task
// in the given WorkflowRun. The currentTaskID, when non-nil, is skipped so a
// running task does not feed its own artifacts back into itself. Later tasks
// (higher ID) override earlier ones, mirroring AWX's set_stats merge.
func (p *TaskPool) GetWorkflowRunArtifacts(projectID int, runID int, currentTaskID *int) (map[string]any, error) {
	tasks, err := p.getWorkflowRunTasks(projectID, runID)
	if err != nil {
		return nil, err
	}
	plain := make([]db.Task, 0, len(tasks))
	for _, t := range tasks {
		plain = append(plain, t.Task)
	}
	return artifacts.CollectFromTasks(plain, currentTaskID), nil
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

func (p *TaskPool) isWorkflowNodeReady(workflow db.WorkflowTemplate, destinationNode db.WorkflowNode, statusByNodeID map[int]task_logger.TaskStatus) (ready bool, blocked bool) {
	destinationNodeID := destinationNode.ID
	hasInbound := false
	hasMatchingInbound := false
	hasUnfinishedInbound := false

	for _, edge := range workflow.Edges {
		if edge.DestinationNodeID != destinationNodeID {
			continue
		}
		hasInbound = true

		status, exists := statusByNodeID[edge.SourceNodeID]
		if !exists || !status.IsFinished() {
			hasUnfinishedInbound = true
			if destinationNode.EffectiveConvergenceMode() == db.WorkflowConvergenceAll {
				return false, false
			}
			continue
		}
		if !db.WorkflowConditionMatches(status, edge.Condition) {
			if destinationNode.EffectiveConvergenceMode() == db.WorkflowConvergenceAll {
				return false, true
			}
			continue
		}

		hasMatchingInbound = true
		if destinationNode.EffectiveConvergenceMode() == db.WorkflowConvergenceAny {
			return true, false
		}
	}

	if !hasInbound {
		return false, true
	}

	if destinationNode.EffectiveConvergenceMode() == db.WorkflowConvergenceAny {
		if hasUnfinishedInbound {
			return false, false
		}
		return false, true
	}

	return hasMatchingInbound, !hasMatchingInbound
}

func (p *TaskPool) updateWorkflowRunStatus(run db.WorkflowRun, workflowTasks []db.TaskWithTpl, approvals []db.WorkflowApproval) error {
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
	for _, approval := range approvals {
		switch approval.Status {
		case db.WorkflowApprovalPending:
			hasUnfinished = true
		case db.WorkflowApprovalRejected:
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

	return p.progressWorkflowRunLocked(run, workflow, &task, nil)
}

func (p *TaskPool) evaluateWorkflowApprovalTimeouts(workflow db.WorkflowTemplate, approvals []db.WorkflowApproval) ([]db.WorkflowApproval, error) {
	nodeByID := make(map[int]db.WorkflowNode, len(workflow.Nodes))
	for _, node := range workflow.Nodes {
		nodeByID[node.ID] = node
	}

	now := time.Now().UTC()
	for i := range approvals {
		approval := approvals[i]
		if approval.Status != db.WorkflowApprovalPending {
			continue
		}

		node, ok := nodeByID[approval.WorkflowNodeID]
		if !ok || node.ApprovalTimeout == nil {
			continue
		}

		// Timeout resolution is evaluated lazily during run progression/detail reads.
		// This avoids introducing a background scheduler while keeping timeout behavior deterministic.
		timeoutAt := approval.Created.Add(time.Duration(*node.ApprovalTimeout) * time.Second)
		if now.Before(timeoutAt) {
			continue
		}

		approval.Status = db.WorkflowApprovalRejected
		approval.Resolved = &now
		approval.ResolvedByUserID = nil
		if err := p.store.UpdateWorkflowApproval(approval); err != nil {
			return nil, err
		}
		approvals[i] = approval
	}

	return approvals, nil
}

func mapLatestNodeApprovalStatus(approvals []db.WorkflowApproval) map[int]task_logger.TaskStatus {
	statusByNodeID := make(map[int]task_logger.TaskStatus)
	for _, approval := range approvals {
		switch approval.Status {
		case db.WorkflowApprovalApproved:
			statusByNodeID[approval.WorkflowNodeID] = task_logger.TaskSuccessStatus
		case db.WorkflowApprovalRejected:
			statusByNodeID[approval.WorkflowNodeID] = task_logger.TaskFailStatus
		}
	}
	return statusByNodeID
}

func mapNodeApproval(approvals []db.WorkflowApproval) map[int]db.WorkflowApproval {
	approvalsByNodeID := make(map[int]db.WorkflowApproval)
	for _, approval := range approvals {
		approvalsByNodeID[approval.WorkflowNodeID] = approval
	}
	return approvalsByNodeID
}

func findWorkflowNode(workflow db.WorkflowTemplate, nodeID int) *db.WorkflowNode {
	for i := range workflow.Nodes {
		if workflow.Nodes[i].ID == nodeID {
			return &workflow.Nodes[i]
		}
	}
	return nil
}

func (p *TaskPool) startWorkflowNode(
	run db.WorkflowRun,
	node db.WorkflowNode,
	buildTaskID *int,
	userID *int,
	username string,
) error {
	switch node.EffectiveKind() {
	case db.WorkflowNodeTaskKind:
		tpl, err := p.store.GetTemplate(run.ProjectID, node.TemplateID)
		if err != nil {
			return err
		}

		newTask, err := p.AddTask(db.Task{
			TemplateID:     node.TemplateID,
			ProjectID:      run.ProjectID,
			BuildTaskID:    buildTaskID,
			WorkflowRunID:  &run.ID,
			WorkflowNodeID: &node.ID,
		}, userID, username, run.ProjectID, tpl.App.NeedTaskAlias())
		if err != nil {
			return fmt.Errorf("failed to enqueue workflow node %d: %w", node.ID, err)
		}

		if run.RootTaskID == nil {
			run.RootTaskID = &newTask.ID
			if err = p.store.UpdateWorkflowRun(run); err != nil {
				return err
			}
		}

	case db.WorkflowNodeApprovalKind:
		_, err := p.store.GetWorkflowApproval(run.ProjectID, run.ID, node.ID)
		if err == nil {
			return nil
		}
		if !errors.Is(err, db.ErrNotFound) {
			return err
		}

		_, err = p.store.CreateWorkflowApproval(db.WorkflowApproval{
			ProjectID:      run.ProjectID,
			WorkflowRunID:  run.ID,
			WorkflowNodeID: node.ID,
			Status:         db.WorkflowApprovalPending,
			Created:        time.Now().UTC(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *TaskPool) progressWorkflowRunLocked(
	run db.WorkflowRun,
	workflow db.WorkflowTemplate,
	triggerTask *db.Task,
	user *db.User,
) error {
	workflowTasks, err := p.getWorkflowRunTasks(run.ProjectID, run.ID)
	if err != nil {
		return err
	}
	approvals, err := p.store.GetWorkflowApprovals(run.ProjectID, run.ID)
	if err != nil {
		return err
	}
	approvals, err = p.evaluateWorkflowApprovalTimeouts(workflow, approvals)
	if err != nil {
		return err
	}

	nodeStatusByID := mapLatestNodeTaskStatus(workflowTasks)
	for nodeID, status := range mapLatestNodeApprovalStatus(approvals) {
		nodeStatusByID[nodeID] = status
	}

	nodeTaskByID := mapLatestNodeTask(workflowTasks)
	nodeApprovalByID := mapNodeApproval(approvals)

	var buildTaskID *int
	var userID *int
	var username string
	if triggerTask != nil {
		buildTaskID = &triggerTask.ID
		userID = triggerTask.UserID
	}
	if user != nil {
		userID = &user.ID
		username = user.Username
	}

	for _, node := range workflow.Nodes {
		if _, exists := nodeTaskByID[node.ID]; exists {
			continue
		}
		if _, exists := nodeApprovalByID[node.ID]; exists {
			continue
		}

		ready, blocked := p.isWorkflowNodeReady(workflow, node, nodeStatusByID)
		if !ready || blocked {
			continue
		}

		if err = p.startWorkflowNode(run, node, buildTaskID, userID, username); err != nil {
			return err
		}
	}

	workflowTasks, err = p.getWorkflowRunTasks(run.ProjectID, run.ID)
	if err != nil {
		return err
	}
	approvals, err = p.store.GetWorkflowApprovals(run.ProjectID, run.ID)
	if err != nil {
		return err
	}
	return p.updateWorkflowRunStatus(run, workflowTasks, approvals)
}

func (p *TaskPool) ProgressWorkflowRun(projectID int, runID int, user *db.User) error {
	p.workflowMu.Lock()
	defer p.workflowMu.Unlock()

	run, err := p.store.GetWorkflowRunByID(projectID, runID)
	if err != nil {
		return err
	}

	workflow, err := p.store.GetWorkflowTemplate(projectID, run.WorkflowTemplateID)
	if err != nil {
		return err
	}

	return p.progressWorkflowRunLocked(run, workflow, nil, user)
}

func (p *TaskPool) ResolveWorkflowApproval(projectID int, workflowID int, runID int, nodeID int, status db.WorkflowApprovalStatus, user *db.User) (approval db.WorkflowApproval, err error) {
	if err = status.Validate(); err != nil {
		return
	}
	if status == db.WorkflowApprovalPending {
		err = db.NewValidationError("approval can not be resolved to pending")
		return
	}

	p.workflowMu.Lock()
	defer p.workflowMu.Unlock()

	run, err := p.store.GetWorkflowRun(projectID, workflowID, runID)
	if err != nil {
		return
	}
	workflow, err := p.store.GetWorkflowTemplate(projectID, workflowID)
	if err != nil {
		return
	}
	node := findWorkflowNode(workflow, nodeID)
	if node == nil || node.EffectiveKind() != db.WorkflowNodeApprovalKind {
		err = db.NewValidationError("workflow node is not an approval node")
		return
	}

	approval, err = p.store.GetWorkflowApproval(projectID, run.ID, nodeID)
	if err != nil {
		return
	}
	if approval.Status != db.WorkflowApprovalPending {
		err = db.NewValidationError("approval has already been resolved")
		return
	}

	now := time.Now().UTC()
	approval.Status = status
	approval.Resolved = &now
	if user != nil {
		approval.ResolvedByUserID = &user.ID
	}

	err = p.store.UpdateWorkflowApproval(approval)
	if err != nil {
		return
	}

	err = p.progressWorkflowRunLocked(run, workflow, nil, user)
	return
}
