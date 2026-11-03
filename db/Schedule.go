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
	TemplateID     *int   `db:"template_id" json:"template_id" backup:"-"`
	WorkflowID     *int   `db:"workflow_id" json:"workflow_id" backup:"-"`
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

func (s Schedule) Validate() error {
	if (s.TemplateID == nil) == (s.WorkflowID == nil) {
		return common_errors.NewValidationError("Schedule must run either a template or a workflow")
	}
	if s.WorkflowID != nil && s.RepositoryID != nil {
		return common_errors.NewValidationError("Workflow schedules cannot check repository commits")
	}
	return nil
}
