package workflows

import (
	"errors"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/tasks"
	log "github.com/sirupsen/logrus"
)

var (
	ErrWorkflowNotFound      = errors.New("workflow not found")
	ErrNoStartNode          = errors.New("workflow has no start node")
	ErrInvalidNodeType      = errors.New("invalid node type")
	ErrWorkflowRunNotFound  = errors.New("workflow run not found")
	ErrNodeAlreadyRunning   = errors.New("node is already running")
)

// WorkflowEngine handles workflow execution
type WorkflowEngine struct {
	store    db.Store
	taskPool *tasks.TaskPool
	runs     map[int]*WorkflowRunState
	mu       sync.RWMutex
}

// WorkflowRunState tracks the state of a running workflow
type WorkflowRunState struct {
	Run         db.WorkflowRun
	Nodes       map[int]*NodeState
	Links       []db.WorkflowLink
	ActiveNodes map[int]bool
	mu          sync.Mutex
}

// NodeState tracks the state of a node execution
type NodeState struct {
	RunNode db.WorkflowRunNode
	Node    db.WorkflowNode
	TaskID  *int
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(store db.Store, taskPool *tasks.TaskPool) *WorkflowEngine {
	return &WorkflowEngine{
		store:    store,
		taskPool: taskPool,
		runs:     make(map[int]*WorkflowRunState),
	}
}

// RunWorkflow starts a workflow execution
func (e *WorkflowEngine) RunWorkflow(projectID int, workflowID int, userID *int) (*db.WorkflowRun, error) {
	// Get workflow with nodes and links
	workflow, err := e.store.GetWorkflow(projectID, workflowID)
	if err != nil {
		return nil, ErrWorkflowNotFound
	}

	nodes, err := e.store.GetWorkflowNodes(workflowID)
	if err != nil {
		return nil, err
	}

	links, err := e.store.GetWorkflowLinks(workflowID)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, ErrNoStartNode
	}

	// Find start node (node with no incoming links)
	startNodeID := e.findStartNode(nodes, links)
	if startNodeID == 0 {
		return nil, ErrNoStartNode
	}

	// Create workflow run
	run := db.WorkflowRun{
		WorkflowID: workflowID,
		Status:     db.WorkflowRunStatusPending,
		UserID:     userID,
	}

	run, err = e.store.CreateWorkflowRun(run)
	if err != nil {
		return nil, err
	}

	// Create run nodes for all workflow nodes
	nodeStates := make(map[int]*NodeState)
	for _, node := range nodes {
		runNode := db.WorkflowRunNode{
			WorkflowRunID: run.ID,
			WorkflowNodeID: node.ID,
			Status:        db.WorkflowRunNodeStatusPending,
		}
		runNode, err = e.store.CreateWorkflowRunNode(runNode)
		if err != nil {
			return nil, err
		}

		nodeStates[node.ID] = &NodeState{
			RunNode: runNode,
			Node:    node,
		}
	}

	// Initialize run state
	runState := &WorkflowRunState{
		Run:         run,
		Nodes:       nodeStates,
		Links:       links,
		ActiveNodes: make(map[int]bool),
	}

	e.mu.Lock()
	e.runs[run.ID] = runState
	e.mu.Unlock()

	// Start execution asynchronously
	go e.executeWorkflow(runState, startNodeID)

	return &run, nil
}

// findStartNode finds the node with no incoming links
func (e *WorkflowEngine) findStartNode(nodes []db.WorkflowNode, links []db.WorkflowLink) int {
	incomingCount := make(map[int]int)
	for _, node := range nodes {
		incomingCount[node.ID] = 0
	}

	for _, link := range links {
		incomingCount[link.ToNodeID]++
	}

	for _, node := range nodes {
		if incomingCount[node.ID] == 0 {
			return node.ID
		}
	}

	return 0
}

// executeWorkflow executes the workflow starting from the start node
func (e *WorkflowEngine) executeWorkflow(runState *WorkflowRunState, startNodeID int) {
	runState.mu.Lock()
	runState.Run.Status = db.WorkflowRunStatusRunning
	now := tz.Now()
	runState.Run.Start = &now
	e.store.UpdateWorkflowRun(runState.Run)
	runState.mu.Unlock()

	// Start with the start node
	e.executeNode(runState, startNodeID)
}

// executeNode executes a single node
func (e *WorkflowEngine) executeNode(runState *WorkflowRunState, nodeID int) {
	runState.mu.Lock()
	nodeState, exists := runState.Nodes[nodeID]
	if !exists {
		runState.mu.Unlock()
		return
	}

	// Check if already running
	if runState.ActiveNodes[nodeID] {
		runState.mu.Unlock()
		return
	}

	// Check if already completed
	if nodeState.RunNode.Status == db.WorkflowRunNodeStatusSuccess ||
		nodeState.RunNode.Status == db.WorkflowRunNodeStatusError ||
		nodeState.RunNode.Status == db.WorkflowRunNodeStatusSkipped {
		runState.mu.Unlock()
		return
	}

	runState.ActiveNodes[nodeID] = true
	nodeState.RunNode.Status = db.WorkflowRunNodeStatusRunning
	now := tz.Now()
	nodeState.RunNode.Start = &now
	e.store.UpdateWorkflowRunNode(nodeState.RunNode)
	runState.mu.Unlock()

	// Execute based on node type
	var err error
	var taskID *int

	switch nodeState.Node.Type {
	case db.WorkflowNodeTypeTask:
		taskID, err = e.executeTaskNode(runState, nodeState)
	case db.WorkflowNodeTypePause:
		err = e.executePauseNode(runState, nodeState)
	case db.WorkflowNodeTypeApproval:
		err = e.executeApprovalNode(runState, nodeState)
	default:
		err = ErrInvalidNodeType
	}

	// Update node status
	runState.mu.Lock()
	now = tz.Now()
	nodeState.RunNode.End = &now
	nodeState.RunNode.TaskID = taskID

	if err != nil {
		nodeState.RunNode.Status = db.WorkflowRunNodeStatusError
		nodeState.RunNode.Message = err.Error()
	} else {
		nodeState.RunNode.Status = db.WorkflowRunNodeStatusSuccess
	}

	e.store.UpdateWorkflowRunNode(nodeState.RunNode)
	delete(runState.ActiveNodes, nodeID)
	runState.mu.Unlock()

	// Process outgoing links
	e.processOutgoingLinks(runState, nodeID, nodeState.RunNode.Status)
}

// executeTaskNode executes a task node
func (e *WorkflowEngine) executeTaskNode(runState *WorkflowRunState, nodeState *NodeState) (*int, error) {
	if nodeState.Node.TaskID == nil {
		return nil, errors.New("task node has no task template")
	}

	// Get workflow to get projectID
	workflow, err := e.store.GetWorkflow(0, runState.Run.WorkflowID)
	if err != nil {
		return nil, err
	}

	// Get template
	template, err := e.store.GetTemplate(workflow.ProjectID, *nodeState.Node.TaskID)
	if err != nil {
		return nil, err
	}

	// Create task
	task := db.Task{
		TemplateID: template.ID,
		ProjectID:  template.ProjectID,
		Status:     db.TaskStatusWaiting,
	}

	// Add task to pool
	newTask, err := e.taskPool.AddTask(
		task,
		runState.Run.UserID,
		"workflow",
		template.ProjectID,
		template.App.NeedTaskAlias(),
	)

	if err != nil {
		return nil, err
	}

	taskID := newTask.ID

	// Wait for task completion
	// In a real implementation, we'd subscribe to task events
	// For MVP, we'll poll
	for {
		task, err := e.store.GetTask(template.ProjectID, taskID)
		if err != nil {
			return nil, err
		}

		if task.Status == db.TaskStatusSuccess || task.Status == db.TaskStatusError {
			if task.Status == db.TaskStatusError {
				return &taskID, errors.New("task failed")
			}
			return &taskID, nil
		}

		time.Sleep(1 * time.Second)
	}
}

// executePauseNode executes a pause node (waits for a duration)
func (e *WorkflowEngine) executePauseNode(runState *WorkflowRunState, nodeState *NodeState) error {
	// For MVP, pause for 5 seconds
	// In production, this would read duration from config_json
	time.Sleep(5 * time.Second)
	return nil
}

// executeApprovalNode executes an approval node (waits for manual approval)
func (e *WorkflowEngine) executeApprovalNode(runState *WorkflowRunState, nodeState *NodeState) error {
	// For MVP, approval nodes are auto-approved
	// In production, this would wait for user approval via API
	return nil
}

// processOutgoingLinks processes outgoing links from a completed node
func (e *WorkflowEngine) processOutgoingLinks(runState *WorkflowRunState, nodeID int, nodeStatus db.WorkflowRunNodeStatus) {
	runState.mu.Lock()
	defer runState.mu.Unlock()

	// Find outgoing links
	var nextNodes []int
	for _, link := range runState.Links {
		if link.FromNodeID == nodeID {
			// Check condition
			shouldExecute := false
			switch link.Condition {
			case db.WorkflowLinkConditionAlways:
				shouldExecute = true
			case db.WorkflowLinkConditionSuccess:
				shouldExecute = nodeStatus == db.WorkflowRunNodeStatusSuccess
			case db.WorkflowLinkConditionFailure:
				shouldExecute = nodeStatus == db.WorkflowRunNodeStatusError
			}

			if shouldExecute {
				nextNodes = append(nextNodes, link.ToNodeID)
			}
		}
	}

	// Check if all incoming nodes are completed before executing
	for _, nextNodeID := range nextNodes {
		if e.canExecuteNode(runState, nextNodeID) {
			go e.executeNode(runState, nextNodeID)
		}
	}

	// Check if workflow is complete
	if len(runState.ActiveNodes) == 0 && e.allNodesCompleted(runState) {
		e.finishWorkflow(runState)
	}
}

// canExecuteNode checks if a node can be executed (all incoming nodes are completed)
func (e *WorkflowEngine) canExecuteNode(runState *WorkflowRunState, nodeID int) bool {
	// Find all incoming links
	incomingNodes := make(map[int]bool)
	for _, link := range runState.Links {
		if link.ToNodeID == nodeID {
			incomingNodes[link.FromNodeID] = true
		}
	}

	// If no incoming links, it's a start node (shouldn't happen here, but handle it)
	if len(incomingNodes) == 0 {
		return true
	}

	// Check if all incoming nodes are completed
	for fromNodeID := range incomingNodes {
		nodeState, exists := runState.Nodes[fromNodeID]
		if !exists {
			return false
		}

		if nodeState.RunNode.Status != db.WorkflowRunNodeStatusSuccess &&
			nodeState.RunNode.Status != db.WorkflowRunNodeStatusError &&
			nodeState.RunNode.Status != db.WorkflowRunNodeStatusSkipped {
			return false
		}
	}

	return true
}

// allNodesCompleted checks if all nodes are completed
func (e *WorkflowEngine) allNodesCompleted(runState *WorkflowRunState) bool {
	for _, nodeState := range runState.Nodes {
		if nodeState.RunNode.Status != db.WorkflowRunNodeStatusSuccess &&
			nodeState.RunNode.Status != db.WorkflowRunNodeStatusError &&
			nodeState.RunNode.Status != db.WorkflowRunNodeStatusSkipped {
			return false
		}
	}
	return true
}

// finishWorkflow marks the workflow as completed
func (e *WorkflowEngine) finishWorkflow(runState *WorkflowRunState) {
	runState.mu.Lock()
	defer runState.mu.Unlock()

	now := tz.Now()
	runState.Run.End = &now

	// Determine final status
	hasError := false
	for _, nodeState := range runState.Nodes {
		if nodeState.RunNode.Status == db.WorkflowRunNodeStatusError {
			hasError = true
			break
		}
	}

	if hasError {
		runState.Run.Status = db.WorkflowRunStatusError
	} else {
		runState.Run.Status = db.WorkflowRunStatusSuccess
	}

	e.store.UpdateWorkflowRun(runState.Run)

	// Remove from active runs
	e.mu.Lock()
	delete(e.runs, runState.Run.ID)
	e.mu.Unlock()
}

// GetWorkflowRunState gets the current state of a workflow run
func (e *WorkflowEngine) GetWorkflowRunState(runID int) (*WorkflowRunState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	runState, exists := e.runs[runID]
	if !exists {
		return nil, ErrWorkflowRunNotFound
	}

	return runState, nil
}
