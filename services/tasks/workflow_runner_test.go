package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func TestWorkflowTraversal_OnFailureDoesNotTraverseOnSuccess(t *testing.T) {
	pool := &TaskPool{}
	workflow := db.WorkflowTemplate{
		Nodes: []db.WorkflowNode{{ID: 1}, {ID: 2}, {ID: 4}},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 1, DestinationNodeID: 4, Condition: db.WorkflowEdgeOnFailure},
		},
	}

	statusByNodeID := map[int]task_logger.TaskStatus{1: task_logger.TaskFailStatus}

	readyB, blockedB := pool.isWorkflowNodeReady(workflow, 2, statusByNodeID)
	if readyB || !blockedB {
		t.Fatal("node B should not be traversed when A fails")
	}

	readyD, blockedD := pool.isWorkflowNodeReady(workflow, 4, statusByNodeID)
	if !readyD || blockedD {
		t.Fatal("node D should be traversed when A fails via on_failure")
	}
}

func TestWorkflowConvergence(t *testing.T) {
	pool := &TaskPool{}
	workflow := db.WorkflowTemplate{
		Nodes: []db.WorkflowNode{{ID: 1}, {ID: 2}, {ID: 3}},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 2, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
		},
	}

	ready, blocked := pool.isWorkflowNodeReady(workflow, 3, map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
	})
	if ready || blocked {
		t.Fatal("node should wait for all parents")
	}

	ready, blocked = pool.isWorkflowNodeReady(workflow, 3, map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
		2: task_logger.TaskFailStatus,
	})
	if ready || !blocked {
		t.Fatal("node should be blocked when a parent does not satisfy the edge condition")
	}

	ready, blocked = pool.isWorkflowNodeReady(workflow, 3, map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
		2: task_logger.TaskSuccessStatus,
	})
	if !ready || blocked {
		t.Fatal("node should run after all converging parents satisfy conditions")
	}
}
