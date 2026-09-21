package db

import (
	"time"
)

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

	// AlertMode is inherit (whatever the template resolves to) or ids
	// (only AlertIDs, which may be empty to make the schedule silent).
	AlertMode string `db:"alert_mode" json:"alert_mode"`
	// AlertIDs are used when AlertMode is ids. nil on update means the
	// client omitted the field and the existing bindings are kept.
	AlertIDs []int `db:"-" json:"alert_ids" backup:"-"`
}

// NormalizeAlerts fills the default mode and validates it.
func (s *Schedule) NormalizeAlerts() error {
	if s.AlertMode == "" {
		s.AlertMode = AlertModeInherit
	}
	return ValidateAlertMode(s.AlertMode, AlertModeInherit, AlertModeIDs)
}

type ScheduleWithTpl struct {
	Schedule
	TemplateName string `db:"tpl_name" json:"tpl_name"`
}
