package db

import (
	"time"
)

type Workflow struct {
	ID          int       `db:"id" json:"id"`
	ProjectID   int       `db:"project_id" json:"project_id"`
	Name        string    `db:"name" json:"name" binding:"required"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type WorkflowNode struct {
	ID                int    `db:"id" json:"id"`
	WorkflowID        int    `db:"workflow_id" json:"workflow_id"`
	ProjectTemplateID *int   `db:"project_template_id" json:"task_id"` // Mapped to task_id in JSON for frontend compatibility as requested
	Type              string `db:"type" json:"type"` // task, pause, approval
	PositionX         int    `db:"position_x" json:"position_x"`
	PositionY         int    `db:"position_y" json:"position_y"`
	ConfigJSON        string `db:"config_json" json:"config_json"`
}

type WorkflowLink struct {
	ID         int    `db:"id" json:"id"`
	WorkflowID int    `db:"workflow_id" json:"workflow_id"`
	FromNodeID int    `db:"from_node_id" json:"from_node_id"`
	ToNodeID   int    `db:"to_node_id" json:"to_node_id"`
	Condition  string `db:"condition" json:"condition"` // success, failure, always
}

type WorkflowRun struct {
	ID         int        `db:"id" json:"id"`
	WorkflowID int        `db:"workflow_id" json:"workflow_id"`
	Status     string     `db:"status" json:"status"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	FinishedAt *time.Time `db:"finished_at" json:"finished_at"`
}

type WorkflowNodeRun struct {
	ID             int        `db:"id" json:"id"`
	WorkflowRunID  int        `db:"workflow_run_id" json:"workflow_run_id"`
	WorkflowNodeID int        `db:"workflow_node_id" json:"workflow_node_id"`
	Status         string     `db:"status" json:"status"`
	TaskID         *int       `db:"task_id" json:"task_id"` // Reference to the actual task run
	StartedAt      time.Time  `db:"started_at" json:"started_at"`
	FinishedAt     *time.Time `db:"finished_at" json:"finished_at"`
}
