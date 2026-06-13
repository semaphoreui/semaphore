package db

import (
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// Workflow represents a workflow template
type Workflow struct {
	ID          int       `db:"id" json:"id"`
	ProjectID  int       `db:"project_id" json:"project_id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	Created     time.Time `db:"created" json:"created"`
	Updated     time.Time `db:"updated" json:"updated"`
}

// WorkflowNodeType represents the type of a workflow node
type WorkflowNodeType string

const (
	WorkflowNodeTypeTask     WorkflowNodeType = "task"
	WorkflowNodeTypePause    WorkflowNodeType = "pause"
	WorkflowNodeTypeApproval WorkflowNodeType = "approval"
)

// WorkflowNode represents a node in a workflow
type WorkflowNode struct {
	ID        int             `db:"id" json:"id"`
	WorkflowID int           `db:"workflow_id" json:"workflow_id"`
	TaskID    *int           `db:"task_id" json:"task_id,omitempty"`
	Type      WorkflowNodeType `db:"type" json:"type"`
	PositionX float64        `db:"position_x" json:"position_x"`
	PositionY float64        `db:"position_y" json:"position_y"`
	ConfigJSON *string       `db:"config_json" json:"config_json,omitempty"`
}

// WorkflowLinkCondition represents the condition for a workflow link
type WorkflowLinkCondition string

const (
	WorkflowLinkConditionSuccess WorkflowLinkCondition = "success"
	WorkflowLinkConditionFailure WorkflowLinkCondition = "failure"
	WorkflowLinkConditionAlways   WorkflowLinkCondition = "always"
)

// WorkflowLink represents a connection between workflow nodes
type WorkflowLink struct {
	ID        int                    `db:"id" json:"id"`
	WorkflowID int                  `db:"workflow_id" json:"workflow_id"`
	FromNodeID int                  `db:"from_node_id" json:"from_node_id"`
	ToNodeID   int                  `db:"to_node_id" json:"to_node_id"`
	Condition WorkflowLinkCondition `db:"condition" json:"condition"`
}

// WorkflowRunStatus represents the status of a workflow run
type WorkflowRunStatus string

const (
	WorkflowRunStatusPending WorkflowRunStatus = "pending"
	WorkflowRunStatusRunning WorkflowRunStatus = "running"
	WorkflowRunStatusSuccess WorkflowRunStatus = "success"
	WorkflowRunStatusError   WorkflowRunStatus = "error"
	WorkflowRunStatusCanceled WorkflowRunStatus = "canceled"
)

// WorkflowRun represents an execution instance of a workflow
type WorkflowRun struct {
	ID        int              `db:"id" json:"id"`
	WorkflowID int            `db:"workflow_id" json:"workflow_id"`
	Status    WorkflowRunStatus `db:"status" json:"status"`
	UserID    *int             `db:"user_id" json:"user_id,omitempty"`
	Created   time.Time       `db:"created" json:"created"`
	Start     *time.Time       `db:"start" json:"start,omitempty"`
	End       *time.Time       `db:"end" json:"end,omitempty"`
	Message   string           `db:"message" json:"message,omitempty"`
}

// WorkflowRunNodeStatus represents the status of a workflow run node
type WorkflowRunNodeStatus string

const (
	WorkflowRunNodeStatusPending WorkflowRunNodeStatus = "pending"
	WorkflowRunNodeStatusRunning WorkflowRunNodeStatus = "running"
	WorkflowRunNodeStatusSuccess WorkflowRunNodeStatus = "success"
	WorkflowRunNodeStatusError   WorkflowRunNodeStatus = "error"
	WorkflowRunNodeStatusSkipped WorkflowRunNodeStatus = "skipped"
)

// WorkflowRunNode represents the execution state of a node in a workflow run
type WorkflowRunNode struct {
	ID            int                  `db:"id" json:"id"`
	WorkflowRunID int                 `db:"workflow_run_id" json:"workflow_run_id"`
	WorkflowNodeID int                `db:"workflow_node_id" json:"workflow_node_id"`
	TaskID       *int                 `db:"task_id" json:"task_id,omitempty"`
	Status       WorkflowRunNodeStatus `db:"status" json:"status"`
	Created      time.Time            `db:"created" json:"created"`
	Start        *time.Time           `db:"start" json:"start,omitempty"`
	End          *time.Time           `db:"end" json:"end,omitempty"`
	Message      string               `db:"message" json:"message,omitempty"`
}

// WorkflowWithNodes represents a workflow with its nodes and links
type WorkflowWithNodes struct {
	Workflow
	Nodes []WorkflowNode `json:"nodes"`
	Links []WorkflowLink `json:"links"`
}

// WorkflowRunWithNodes represents a workflow run with its node executions
type WorkflowRunWithNodes struct {
	WorkflowRun
	Nodes []WorkflowRunNode `json:"nodes"`
}
