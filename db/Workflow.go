package db

import (
	"time"
)

// Workflow represents a workflow definition
type Workflow struct {
	ID          int        `db:"id" json:"id"`
	ProjectID   int        `db:"project_id" json:"project_id"`
	Name        string     `db:"name" json:"name"`
	Description *string    `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	Nodes       []WorkflowNode `db:"-" json:"nodes,omitempty"`
	Links       []WorkflowLink `db:"-" json:"links,omitempty"`
}

// WorkflowNodeType defines the type of workflow node
type WorkflowNodeType string

const (
	WorkflowNodeTypeTask     WorkflowNodeType = "task"
	WorkflowNodeTypePause    WorkflowNodeType = "pause"
	WorkflowNodeTypeApproval WorkflowNodeType = "approval"
)

// WorkflowNode represents a node in a workflow
type WorkflowNode struct {
	ID             int              `db:"id" json:"id"`
	WorkflowID     int              `db:"workflow_id" json:"workflow_id"`
	TaskTemplateID *int             `db:"task_template_id" json:"task_template_id,omitempty"`
	Type           WorkflowNodeType `db:"type" json:"type"`
	Name           string           `db:"name" json:"name"`
	PositionX      float64          `db:"position_x" json:"position_x"`
	PositionY      float64          `db:"position_y" json:"position_y"`
	Config         *string          `db:"config" json:"config,omitempty"`
}

// WorkflowLinkCondition defines when a link should be followed
type WorkflowLinkCondition string

const (
	WorkflowLinkConditionSuccess WorkflowLinkCondition = "success"
	WorkflowLinkConditionFailure WorkflowLinkCondition = "failure"
	WorkflowLinkConditionAlways  WorkflowLinkCondition = "always"
)

// WorkflowLink represents an edge between workflow nodes
type WorkflowLink struct {
	ID         int                   `db:"id" json:"id"`
	WorkflowID int                   `db:"workflow_id" json:"workflow_id"`
	FromNodeID int                   `db:"from_node_id" json:"from_node_id"`
	ToNodeID   int                   `db:"to_node_id" json:"to_node_id"`
	Condition  WorkflowLinkCondition `db:"condition" json:"condition"`
}

// WorkflowRunStatus defines the status of a workflow run
type WorkflowRunStatus string

const (
	WorkflowRunStatusPending WorkflowRunStatus = "pending"
	WorkflowRunStatusRunning WorkflowRunStatus = "running"
	WorkflowRunStatusSuccess WorkflowRunStatus = "success"
	WorkflowRunStatusFailure WorkflowRunStatus = "failure"
	WorkflowRunStatusStopped WorkflowRunStatus = "stopped"
)

// WorkflowRun represents a workflow execution instance
type WorkflowRun struct {
	ID         int               `db:"id" json:"id"`
	WorkflowID int               `db:"workflow_id" json:"workflow_id"`
	ProjectID  int               `db:"project_id" json:"project_id"`
	UserID     *int              `db:"user_id" json:"user_id,omitempty"`
	Status     WorkflowRunStatus `db:"status" json:"status"`
	Start      *time.Time        `db:"start" json:"start,omitempty"`
	End        *time.Time        `db:"end" json:"end,omitempty"`
	Message    *string           `db:"message" json:"message,omitempty"`
}

// WorkflowRunWithWorkflow includes workflow details
type WorkflowRunWithWorkflow struct {
	WorkflowRun
	WorkflowName string `db:"workflow_name" json:"workflow_name"`
}

// WorkflowNodeRunStatus defines the status of a workflow node run
type WorkflowNodeRunStatus string

const (
	WorkflowNodeRunStatusPending WorkflowNodeRunStatus = "pending"
	WorkflowNodeRunStatusRunning WorkflowNodeRunStatus = "running"
	WorkflowNodeRunStatusSuccess WorkflowNodeRunStatus = "success"
	WorkflowNodeRunStatusFailure WorkflowNodeRunStatus = "failure"
	WorkflowNodeRunStatusSkipped WorkflowNodeRunStatus = "skipped"
	WorkflowNodeRunStatusStopped WorkflowNodeRunStatus = "stopped"
)

// WorkflowNodeRun represents the execution of a single node in a workflow run
type WorkflowNodeRun struct {
	ID            int                   `db:"id" json:"id"`
	WorkflowRunID int                   `db:"workflow_run_id" json:"workflow_run_id"`
	NodeID        int                   `db:"node_id" json:"node_id"`
	TaskID        *int                  `db:"task_id" json:"task_id,omitempty"`
	Status        WorkflowNodeRunStatus `db:"status" json:"status"`
	Start         *time.Time            `db:"start" json:"start,omitempty"`
	End           *time.Time            `db:"end" json:"end,omitempty"`
	Message       *string               `db:"message" json:"message,omitempty"`
}

// Validate validates the workflow
func (w *Workflow) Validate() error {
	if w.Name == "" {
		return &ValidationError{"workflow name cannot be empty"}
	}
	return nil
}

// Validate validates the workflow node
func (n *WorkflowNode) Validate() error {
	if n.Name == "" {
		return &ValidationError{"workflow node name cannot be empty"}
	}
	if n.Type == WorkflowNodeTypeTask && n.TaskTemplateID == nil {
		return &ValidationError{"task node must have a template_id"}
	}
	return nil
}
