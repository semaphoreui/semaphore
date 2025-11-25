package workflows

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/tasks"
	log "github.com/sirupsen/logrus"
)

// WorkflowEngine manages workflow execution
type WorkflowEngine struct {
	store    db.Store
	taskPool *tasks.TaskPool
	mu       sync.RWMutex
	runs     map[int]*WorkflowRunContext
}

// WorkflowRunContext holds the runtime context for a workflow execution
type WorkflowRunContext struct {
	Run         db.WorkflowRun
	Workflow    db.Workflow
	NodeRuns    map[int]db.WorkflowNodeRun
	CompletedNodes map[int]bool
	RunningNodes   map[int]bool
	mu          sync.RWMutex
	stopped     bool
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(store db.Store, taskPool *tasks.TaskPool) *WorkflowEngine {
	return &WorkflowEngine{
		store:    store,
		taskPool: taskPool,
		runs:     make(map[int]*WorkflowRunContext),
	}
}

// ExecuteWorkflow starts executing a workflow
func (e *WorkflowEngine) ExecuteWorkflow(runID int) error {
	run, err := e.store.GetWorkflowRun(runID)
	if err != nil {
		return fmt.Errorf("failed to get workflow run: %w", err)
	}

	workflow, err := e.store.GetWorkflow(run.ProjectID, run.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Create run context
	ctx := &WorkflowRunContext{
		Run:         run,
		Workflow:    workflow,
		NodeRuns:    make(map[int]db.WorkflowNodeRun),
		CompletedNodes: make(map[int]bool),
		RunningNodes:   make(map[int]bool),
	}

	e.mu.Lock()
	e.runs[runID] = ctx
	e.mu.Unlock()

	// Update run status to running
	run.Status = db.WorkflowRunStatusRunning
	now := tz.Now()
	run.Start = &now
	err = e.store.UpdateWorkflowRun(run)
	if err != nil {
		return fmt.Errorf("failed to update workflow run: %w", err)
	}

	// Execute workflow
	go e.executeWorkflowAsync(ctx)

	return nil
}

// executeWorkflowAsync executes the workflow asynchronously
func (e *WorkflowEngine) executeWorkflowAsync(ctx *WorkflowRunContext) {
	defer func() {
		e.mu.Lock()
		delete(e.runs, ctx.Run.ID)
		e.mu.Unlock()
	}()

	// Find start nodes (nodes with no incoming links)
	startNodes := e.findStartNodes(ctx.Workflow)

	if len(startNodes) == 0 {
		e.failWorkflow(ctx, "No start nodes found in workflow")
		return
	}

	// Execute start nodes
	for _, node := range startNodes {
		if ctx.stopped {
			return
		}
		e.executeNode(ctx, node)
	}

	// Wait for all nodes to complete
	e.waitForCompletion(ctx)

	// Update final workflow status
	e.finalizeWorkflow(ctx)
}

// findStartNodes finds all nodes with no incoming links
func (e *WorkflowEngine) findStartNodes(workflow db.Workflow) []db.WorkflowNode {
	hasIncoming := make(map[int]bool)

	for _, link := range workflow.Links {
		hasIncoming[link.ToNodeID] = true
	}

	var startNodes []db.WorkflowNode
	for _, node := range workflow.Nodes {
		if !hasIncoming[node.ID] {
			startNodes = append(startNodes, node)
		}
	}

	return startNodes
}

// executeNode executes a single workflow node
func (e *WorkflowEngine) executeNode(ctx *WorkflowRunContext, node db.WorkflowNode) {
	ctx.mu.Lock()
	if ctx.RunningNodes[node.ID] || ctx.CompletedNodes[node.ID] {
		ctx.mu.Unlock()
		return
	}
	ctx.RunningNodes[node.ID] = true
	ctx.mu.Unlock()

	// Create node run
	nodeRun := db.WorkflowNodeRun{
		WorkflowRunID: ctx.Run.ID,
		NodeID:        node.ID,
		Status:        db.WorkflowNodeRunStatusRunning,
	}

	newNodeRun, err := e.store.CreateWorkflowNodeRun(nodeRun)
	if err != nil {
		log.WithError(err).Error("Failed to create workflow node run")
		ctx.mu.Lock()
		ctx.RunningNodes[node.ID] = false
		ctx.mu.Unlock()
		return
	}

	ctx.mu.Lock()
	ctx.NodeRuns[node.ID] = newNodeRun
	ctx.mu.Unlock()

	// Execute node based on type
	var success bool
	var message string

	switch node.Type {
	case db.WorkflowNodeTypeTask:
		success, message = e.executeTaskNode(ctx, node, &newNodeRun)
	case db.WorkflowNodeTypePause:
		success, message = e.executePauseNode(ctx, node)
	case db.WorkflowNodeTypeApproval:
		success, message = e.executeApprovalNode(ctx, node)
	default:
		success = false
		message = fmt.Sprintf("Unknown node type: %s", node.Type)
	}

	// Update node run status
	now := tz.Now()
	newNodeRun.End = &now
	if success {
		newNodeRun.Status = db.WorkflowNodeRunStatusSuccess
	} else {
		newNodeRun.Status = db.WorkflowNodeRunStatusFailure
		newNodeRun.Message = &message
	}

	err = e.store.UpdateWorkflowNodeRun(newNodeRun)
	if err != nil {
		log.WithError(err).Error("Failed to update workflow node run")
	}

	ctx.mu.Lock()
	ctx.NodeRuns[node.ID] = newNodeRun
	delete(ctx.RunningNodes, node.ID)
	ctx.CompletedNodes[node.ID] = true
	ctx.mu.Unlock()

	// Find and execute next nodes
	if !ctx.stopped {
		e.executeNextNodes(ctx, node, success)
	}
}

// executeTaskNode executes a task node
func (e *WorkflowEngine) executeTaskNode(ctx *WorkflowRunContext, node db.WorkflowNode, nodeRun *db.WorkflowNodeRun) (bool, string) {
	if node.TaskTemplateID == nil {
		return false, "Task node has no template"
	}

	// Get template
	template, err := e.store.GetTemplate(ctx.Workflow.ProjectID, *node.TaskTemplateID)
	if err != nil {
		return false, fmt.Sprintf("Failed to get template: %v", err)
	}

	// Parse node config for task parameters
	var taskParams map[string]interface{}
	if node.Config != nil {
		err := json.Unmarshal([]byte(*node.Config), &taskParams)
		if err != nil {
			log.WithError(err).Warn("Failed to parse node config")
		}
	}

	// Create task
	task := db.Task{
		TemplateID: template.ID,
		ProjectID:  ctx.Workflow.ProjectID,
		Status:     task_logger.TaskWaitingStatus,
		UserID:     ctx.Run.UserID,
	}

	// Apply task params from node config
	if taskParams != nil {
		if environment, ok := taskParams["environment"].(map[string]interface{}); ok {
			task.Environment = db.ObjectToJSON(environment)
		}
	}

	// Create the task
	maxTasks := 1000 // Default max tasks
	newTask, err := e.store.CreateTask(task, maxTasks)
	if err != nil {
		return false, fmt.Sprintf("Failed to create task: %v", err)
	}

	// Update node run with task ID
	nodeRun.TaskID = &newTask.ID
	err = e.store.UpdateWorkflowNodeRun(*nodeRun)
	if err != nil {
		log.WithError(err).Error("Failed to update node run with task ID")
	}

	// Add task to pool for execution
	e.taskPool.AddTask(newTask, ctx.Run.UserID)

	// Wait for task completion
	for {
		if ctx.stopped {
			return false, "Workflow stopped"
		}

		time.Sleep(2 * time.Second)

		// Get task status
		task, err := e.store.GetTask(ctx.Workflow.ProjectID, newTask.ID)
		if err != nil {
			return false, fmt.Sprintf("Failed to get task status: %v", err)
		}

		// Check if task is complete
		if task.Status == task_logger.TaskSuccessStatus {
			return true, "Task completed successfully"
		} else if task.Status == task_logger.TaskFailStatus {
			return false, "Task failed"
		} else if task.Status == task_logger.TaskStoppedStatus {
			return false, "Task was stopped"
		}
		// Task is still running, continue waiting
	}
}

// executePauseNode executes a pause node
func (e *WorkflowEngine) executePauseNode(ctx *WorkflowRunContext, node db.WorkflowNode) (bool, string) {
	// Parse pause duration from config
	var config map[string]interface{}
	pauseDuration := 5 * time.Second // Default

	if node.Config != nil {
		err := json.Unmarshal([]byte(*node.Config), &config)
		if err == nil {
			if duration, ok := config["duration"].(float64); ok {
				pauseDuration = time.Duration(duration) * time.Second
			}
		}
	}

	time.Sleep(pauseDuration)
	return true, fmt.Sprintf("Paused for %v", pauseDuration)
}

// executeApprovalNode executes an approval node (MVP: auto-approve)
func (e *WorkflowEngine) executeApprovalNode(ctx *WorkflowRunContext, node db.WorkflowNode) (bool, string) {
	// For MVP, auto-approve after a short delay
	// In production, this would wait for manual approval
	time.Sleep(1 * time.Second)
	return true, "Auto-approved (MVP)"
}

// executeNextNodes finds and executes the next nodes based on the current node's result
func (e *WorkflowEngine) executeNextNodes(ctx *WorkflowRunContext, currentNode db.WorkflowNode, success bool) {
	// Find outgoing links
	var nextNodes []db.WorkflowNode

	for _, link := range ctx.Workflow.Links {
		if link.FromNodeID != currentNode.ID {
			continue
		}

		// Check if link condition matches
		shouldFollow := false
		switch link.Condition {
		case db.WorkflowLinkConditionAlways:
			shouldFollow = true
		case db.WorkflowLinkConditionSuccess:
			shouldFollow = success
		case db.WorkflowLinkConditionFailure:
			shouldFollow = !success
		}

		if shouldFollow {
			// Find the target node
			for _, node := range ctx.Workflow.Nodes {
				if node.ID == link.ToNodeID {
					nextNodes = append(nextNodes, node)
					break
				}
			}
		}
	}

	// Execute next nodes (parallel execution)
	for _, node := range nextNodes {
		go e.executeNode(ctx, node)
	}
}

// waitForCompletion waits for all nodes to complete
func (e *WorkflowEngine) waitForCompletion(ctx *WorkflowRunContext) {
	for {
		ctx.mu.RLock()
		running := len(ctx.RunningNodes)
		stopped := ctx.stopped
		ctx.mu.RUnlock()

		if stopped || running == 0 {
			break
		}

		time.Sleep(1 * time.Second)
	}
}

// finalizeWorkflow updates the final workflow status
func (e *WorkflowEngine) finalizeWorkflow(ctx *WorkflowRunContext) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	// Determine final status
	var status db.WorkflowRunStatus
	var hasFailures bool

	for _, nodeRun := range ctx.NodeRuns {
		if nodeRun.Status == db.WorkflowNodeRunStatusFailure {
			hasFailures = true
			break
		}
	}

	if ctx.stopped {
		status = db.WorkflowRunStatusStopped
	} else if hasFailures {
		status = db.WorkflowRunStatusFailure
	} else {
		status = db.WorkflowRunStatusSuccess
	}

	// Update workflow run
	run := ctx.Run
	run.Status = status
	now := tz.Now()
	run.End = &now

	err := e.store.UpdateWorkflowRun(run)
	if err != nil {
		log.WithError(err).Error("Failed to finalize workflow run")
	}
}

// failWorkflow marks the workflow as failed
func (e *WorkflowEngine) failWorkflow(ctx *WorkflowRunContext, message string) {
	run := ctx.Run
	run.Status = db.WorkflowRunStatusFailure
	run.Message = &message
	now := tz.Now()
	run.End = &now

	err := e.store.UpdateWorkflowRun(run)
	if err != nil {
		log.WithError(err).Error("Failed to fail workflow run")
	}
}

// StopWorkflow stops a running workflow
func (e *WorkflowEngine) StopWorkflow(runID int) error {
	e.mu.RLock()
	ctx, exists := e.runs[runID]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("workflow run not found")
	}

	ctx.mu.Lock()
	ctx.stopped = true
	ctx.mu.Unlock()

	return nil
}
