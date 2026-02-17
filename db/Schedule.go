package db

import "time"

const (
	ScheduleTypeCron  = ""
	ScheduleTypeRunAt = "run_at"
)

type Schedule struct {
	ID             int    `db:"id" json:"id" backup:"-"`
	ProjectID      int    `db:"project_id" json:"project_id" backup:"-"`
	TemplateID     int    `db:"template_id" json:"template_id" backup:"-"`
	CronFormat     string `db:"cron_format" json:"cron_format"`
	Name           string `db:"name" json:"name"`
	Active         bool   `db:"active" json:"active"`
	Type           string `db:"type" json:"type"`
	DeleteAfterRun bool   `db:"delete_after_run" json:"delete_after_run"`

	LastCommitHash *string    `db:"last_commit_hash" json:"-" backup:"-"`
	RepositoryID   *int       `db:"repository_id" json:"repository_id" backup:"-"`
	RunAt          *time.Time `db:"run_at" json:"run_at,omitempty"`

	TaskParamsID *int        `db:"task_params_id" json:"-" backup:"-"`
	TaskParams   *TaskParams `db:"-" json:"task_params,omitempty" backup:"task_params"`
}

type ScheduleWithTpl struct {
	Schedule
	TemplateName string `db:"tpl_name" json:"tpl_name"`
}
