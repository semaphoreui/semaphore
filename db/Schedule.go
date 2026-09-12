package db

import (
	"time"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
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

	// AlertMode is inherit (use the template list) or ids (use AlertIDs, which may be empty).
	AlertMode      string `db:"alert_mode" json:"alert_mode"`
	AlertIDs       []int  `db:"-" json:"alert_ids" backup:"-"`
	AlertOnSuccess *bool  `db:"alert_on_success" json:"alert_on_success"`
	AlertOnError   *bool  `db:"alert_on_error" json:"alert_on_error"`
}

type ScheduleWithTpl struct {
	Schedule
	TemplateName string `db:"tpl_name" json:"tpl_name"`
}

func (s *Schedule) NormalizeAlerts() error {
	if s.AlertMode == "" {
		s.AlertMode = AlertModeInherit
	}
	if s.AlertMode != AlertModeInherit && s.AlertMode != AlertModeIDs {
		return common_errors.NewValidationError("alert_mode must be inherit or ids")
	}
	return nil
}
