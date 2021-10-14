package db

import "time"

//Task is a model of a task which will be executed by the runner
type Task struct {
	ID         int `db:"id" json:"id"`
	TemplateID int `db:"template_id" json:"template_id" binding:"required"`
	ProjectID  int `db:"project_id" json:"project_id"`

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

	Version       *string `db:"version" json:"version"`
	CommitHash    *string `db:"commit_hash" json:"commit_hash"`
	CommitMessage *string `db:"commit_message" json:"commit_message"`
	Message       *string `db:"message" json:"message"`
}

// TaskWithTpl is the task data with additional fields
type TaskWithTpl struct {
	Task
	TemplatePlaybook string  `db:"tpl_playbook" json:"tpl_playbook"`
	TemplateAlias    string  `db:"tpl_alias" json:"tpl_alias"`
	TemplateType     string  `db:"tpl_type" json:"tpl_type"`
	UserName         *string `db:"user_name" json:"user_name"`
}

// TaskOutput is the ansible log output from the task
type TaskOutput struct {
	TaskID int       `db:"task_id" json:"task_id"`
	Task   string    `db:"task" json:"task"`
	Time   time.Time `db:"time" json:"time"`
	Output string    `db:"output" json:"output"`
}
