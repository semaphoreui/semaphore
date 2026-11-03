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
	ID        int `db:"id" json:"id" backup:"-"`
	ProjectID int `db:"project_id" json:"project_id" backup:"-"`

	// TemplateID is the task template the schedule starts. Exactly one of
	// TemplateID and WorkflowID is set; the store refuses anything else.
	TemplateID *int `db:"template_id" json:"template_id" backup:"-"`
	// WorkflowID is the workflow (a project__workflow_template row) the
	// schedule starts instead of a template. The run gets the workflow's
	// survey defaults; TaskParams do not apply and are dropped by the store.
	WorkflowID *int `db:"workflow_id" json:"workflow_id" backup:"-"`

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

// ScheduleWithTpl is a schedule as the project list returns it: with the name
// of its target. TemplateName is empty for a workflow schedule and
// WorkflowName for a template one.
type ScheduleWithTpl struct {
	Schedule
	TemplateName string `db:"tpl_name" json:"tpl_name"`
	WorkflowName string `db:"wf_name" json:"wf_name"`
}

func (s *Schedule) Validate() error {
	if s.TemplateID == nil && s.WorkflowID == nil {
		return common_errors.NewValidationError("Schedule must run a template or a workflow")
	}

	if s.TemplateID != nil && s.WorkflowID != nil {
		return common_errors.NewValidationError("Schedule can run either a template or a workflow, not both")
	}

	// A commit checker polls the repository of its template; a workflow has
	// no single repository to poll.
	if s.WorkflowID != nil && s.RepositoryID != nil {
		return common_errors.NewValidationError("Schedule that runs a workflow cannot check a repository for commits")
	}

	return nil
}

// RunsWorkflow reports whether the schedule starts a workflow rather than a
// task template.
func (s *Schedule) RunsWorkflow() bool {
	return s.WorkflowID != nil
}
