package workflows

import (
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
    "github.com/semaphoreui/semaphore/pkg/task_logger"
    log "github.com/sirupsen/logrus"
)

type WorkflowEngine struct {
	store    db.Store
	taskPool *tasks.TaskPool
}

func NewWorkflowEngine(store db.Store, taskPool *tasks.TaskPool) *WorkflowEngine {
	return &WorkflowEngine{
		store:    store,
		taskPool: taskPool,
	}
}

func (e *WorkflowEngine) RunWorkflow(projectID int, workflowID int) (db.WorkflowRun, error) {
    run := db.WorkflowRun{
        WorkflowID: workflowID,
        Status: "running",
        CreatedAt: time.Now(),
    }
    
    newRun, err := e.store.CreateWorkflowRun(run)
    if err != nil {
        return db.WorkflowRun{}, err
    }
    
    nodes, err := e.store.GetWorkflowNodes(projectID, workflowID)
    if err != nil {
        return newRun, err
    }
    
    links, err := e.store.GetWorkflowLinks(projectID, workflowID)
    if err != nil {
        return newRun, err
    }
    
    incoming := make(map[int]int)
    for _, l := range links {
        incoming[l.ToNodeID]++
    }
    
    var startNodes []db.WorkflowNode
    for _, n := range nodes {
        if incoming[n.ID] == 0 {
            startNodes = append(startNodes, n)
        }
    }
    
    if len(startNodes) == 0 && len(nodes) > 0 {
        startNodes = []db.WorkflowNode{nodes[0]} 
    }
    
    for _, node := range startNodes {
        e.triggerNode(newRun.ID, node)
    }
    
    return newRun, nil
}

func (e *WorkflowEngine) triggerNode(runID int, node db.WorkflowNode) {
    nodeRun := db.WorkflowNodeRun{
        WorkflowRunID: runID,
        WorkflowNodeID: node.ID,
        Status: "running",
        StartedAt: time.Now(),
    }
    
    newNodeRun, err := e.store.CreateWorkflowNodeRun(nodeRun)
    if err != nil {
        log.Error("Failed to create node run", err)
        return
    }
    
    if node.Type == "task" {
        if node.ProjectTemplateID == nil {
            log.Error("Task node missing template ID")
            e.failNodeRun(newNodeRun, "Missing template ID")
            return
        }
        
        taskObj := db.Task{
            TemplateID: *node.ProjectTemplateID,
        }
        
        // Fetch projectID from workflow via run
        run, err := e.store.GetWorkflowRun(0, runID) 
        if err != nil {
             log.Error("Failed to get workflow run", err)
             e.failNodeRun(newNodeRun, "Failed to get workflow run")
             return
        }

        workflow, err := e.store.GetWorkflow(0, run.WorkflowID) 
        if err != nil {
             log.Error("Failed to get workflow", err)
             e.failNodeRun(newNodeRun, "Failed to get workflow")
             return
        }
        
        projectID := workflow.ProjectID 
        
        newTask, err := e.taskPool.AddTask(taskObj, nil, "workflow", projectID, false)
        if err != nil {
            e.failNodeRun(newNodeRun, err.Error())
            return
        }
        
        newNodeRun.TaskID = &newTask.ID
        e.store.UpdateWorkflowNodeRun(newNodeRun)
        
    } else if node.Type == "pause" {
        // Just mark success for now as we don't have resume capability in API yet
        newNodeRun.Status = "success"
        now := time.Now()
        newNodeRun.FinishedAt = &now
        e.store.UpdateWorkflowNodeRun(newNodeRun)
        e.triggerNextNodes(runID, node.ID, "success")
    }
}

func (e *WorkflowEngine) failNodeRun(nodeRun db.WorkflowNodeRun, msg string) {
    nodeRun.Status = "error"
    now := time.Now()
    nodeRun.FinishedAt = &now
    e.store.UpdateWorkflowNodeRun(nodeRun)
}

func (e *WorkflowEngine) HandleTaskCompletion(taskRunner *tasks.TaskRunner) {
    taskID := taskRunner.Task.ID
    
    nodeRun, err := e.store.GetWorkflowNodeRunByTaskID(taskID)
    if err != nil {
        return
    }
    
    taskStatus := taskRunner.Task.Status
    var nodeStatus string
    if taskStatus == task_logger.TaskSuccessStatus {
        nodeStatus = "success"
    } else {
        nodeStatus = "error"
    }
    
    nodeRun.Status = nodeStatus
    now := time.Now()
    nodeRun.FinishedAt = &now
    e.store.UpdateWorkflowNodeRun(nodeRun)
    
    e.triggerNextNodes(nodeRun.WorkflowRunID, nodeRun.WorkflowNodeID, nodeStatus)
}

func (e *WorkflowEngine) triggerNextNodes(runID int, fromNodeID int, status string) {
    run, err := e.store.GetWorkflowRun(0, runID)
    if err != nil { log.Error(err); return }
    
    workflow, err := e.store.GetWorkflow(0, run.WorkflowID)
    if err != nil { log.Error(err); return }
    
    links, err := e.store.GetWorkflowLinks(workflow.ProjectID, workflow.ID)
    if err != nil { log.Error(err); return }
    
    nodes, err := e.store.GetWorkflowNodes(workflow.ProjectID, workflow.ID)
    nodeMap := make(map[int]db.WorkflowNode)
    for _, n := range nodes {
        nodeMap[n.ID] = n
    }
    
    triggered := false

    for _, link := range links {
        if link.FromNodeID == fromNodeID {
            shouldRun := false
            if link.Condition == "always" {
                shouldRun = true
            } else if link.Condition == "success" && status == "success" {
                shouldRun = true
            } else if link.Condition == "failure" && status == "error" {
                shouldRun = true
            }
            
            if shouldRun {
                if nextNode, ok := nodeMap[link.ToNodeID]; ok {
                    e.triggerNode(runID, nextNode)
                    triggered = true
                }
            }
        }
    }
    
    // Check if workflow is finished
    // If no nodes are running.
    // But triggering creates new running nodes.
    // So we check active nodes count.
    
    activeNodeRuns, err := e.store.GetWorkflowNodeRuns(workflow.ProjectID, runID)
    runningCount := 0
    for _, nr := range activeNodeRuns {
        if nr.Status == "running" {
            runningCount++
        }
    }
    
    if runningCount == 0 && !triggered {
        // No running nodes and we didn't trigger any new ones.
        // Mark workflow as finished.
        // Status? If any node failed, maybe fail workflow?
        // Or if all terminals are success?
        // Simple heuristic: if status was error and we didn't trigger handling, maybe fail?
        // But parallel branches exist.
        
        hasError := false
        for _, nr := range activeNodeRuns {
            if nr.Status == "error" {
                hasError = true
            }
        }
        
        wfStatus := "success"
        if hasError {
            wfStatus = "error"
        }
        
        run.Status = wfStatus
        now := time.Now()
        run.FinishedAt = &now
        e.store.UpdateWorkflowRun(run)
    }
}
