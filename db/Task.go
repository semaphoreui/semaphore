package db

import "time"

//Task is a model of a task which will be executed by the runner
type Task struct {
	ID         int `db:"id" json:"id"`
	TemplateID int `db:"template_id" json:"template_id" binding:"required"`

	Status string `db:"status" json:"status"`
	Debug  bool   `db:"debug" json:"debug"`

	DryRun bool `db:"dry_run" json:"dry_run"`

	// override variables
	Playbook    string `db:"playbook" json:"playbook"`
	Environment string `db:"environment" json:"environment"`
	// to fit into []string
	Arguments *string `db:"arguments" json:"arguments"`

	UserID *int `db:"user_id" json:"user_id"`

	Created time.Time  `db:"created" json:"created"`
	Start   *time.Time `db:"start" json:"start"`
	End     *time.Time `db:"end" json:"end"`
}

// TaskOutput is the ansible log output from the task
type TaskOutput struct {
	TaskID int       `db:"task_id" json:"task_id"`
	Task   string    `db:"task" json:"task"`
	Time   time.Time `db:"time" json:"time"`
	Output string    `db:"output" json:"output"`
}
