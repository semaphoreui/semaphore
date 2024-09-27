package db

import (
	"fmt"
	"github.com/go-gorp/gorp/v3"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// Task is a model of a task which will be executed by the runner
type Task struct {
	ID         int `db:"id" json:"id"`
	TemplateID int `db:"template_id" json:"template_id" binding:"required"`
	ProjectID  int `db:"project_id" json:"project_id"`

	Status task_logger.TaskStatus `db:"status" json:"status"`

	Debug  bool `db:"debug" json:"debug"`
	DryRun bool `db:"dry_run" json:"dry_run"`
	Diff   bool `db:"diff" json:"diff"`

	// override variables
	Playbook    string  `db:"playbook" json:"playbook"`
	Environment string  `db:"environment" json:"environment"`
	Limit       string  `db:"hosts_limit" json:"limit"`
	Secret      string  `db:"-" json:"secret"`
	Arguments   *string `db:"arguments" json:"arguments"`

	UserID        *int `db:"user_id" json:"user_id"`
	IntegrationID *int `db:"integration_id" json:"integration_id"`
	ScheduleID    *int `db:"schedule_id" json:"schedule_id"`

	Created time.Time  `db:"created" json:"created"`
	Start   *time.Time `db:"start" json:"start"`
	End     *time.Time `db:"end" json:"end"`

	Message string `db:"message" json:"message"`

	// CommitMessage is a git commit hash of playbook repository which
	// was active when task was created.
	CommitHash *string `db:"commit_hash" json:"commit_hash"`
	// CommitMessage contains message retrieved from git repository after checkout to CommitHash.
	// It is readonly by API.
	CommitMessage string `db:"commit_message" json:"commit_message"`
	BuildTaskID   *int   `db:"build_task_id" json:"build_task_id"`
	// Version is a build version.
	// This field available only for Build tasks.
	Version *string `db:"version" json:"version"`

	InventoryID *int `db:"inventory_id" json:"inventory_id"`
}

func (task *Task) PreInsert(gorp.SqlExecutor) error {
	task.Created = task.Created.UTC()
	return nil
}

func (task *Task) PreUpdate(gorp.SqlExecutor) error {
	if task.Start != nil {
		start := task.Start.UTC()
		task.Start = &start
	}

	if task.End != nil {
		end := task.End.UTC()
		task.End = &end
	}
	return nil
}

func (task *Task) GetIncomingVersion(d Store) *string {
	if task.BuildTaskID == nil {
		return nil
	}

	buildTask, err := d.GetTask(task.ProjectID, *task.BuildTaskID)

	if err != nil {
		return nil
	}

	tpl, err := d.GetTemplate(task.ProjectID, buildTask.TemplateID)
	if err != nil {
		return nil
	}

	if tpl.Type == TemplateBuild {
		return buildTask.Version
	}

	return buildTask.GetIncomingVersion(d)
}

func (task *Task) GetUrl() *string {
	if util.Config.WebHost != "" {
		taskUrl := fmt.Sprintf("%s/project/%d/history?t=%d", util.Config.WebHost, task.ProjectID, task.ID)
		return &taskUrl
	}

	return nil
}

func (task *Task) ValidateNewTask(template Template) error {
	return nil
}

func (task *TaskWithTpl) Fill(d Store) error {
	if task.BuildTaskID != nil {
		build, err := d.GetTask(task.ProjectID, *task.BuildTaskID)
		if err == ErrNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		task.BuildTask = &build
	}
	return nil
}

// TaskWithTpl is the task data with additional fields
type TaskWithTpl struct {
	Task
	TemplatePlaybook string       `db:"tpl_playbook" json:"tpl_playbook"`
	TemplateAlias    string       `db:"tpl_alias" json:"tpl_alias"`
	TemplateType     TemplateType `db:"tpl_type" json:"tpl_type"`
	TemplateApp      TemplateApp  `db:"tpl_app" json:"tpl_app"`
	UserName         *string      `db:"user_name" json:"user_name"`
	BuildTask        *Task        `db:"-" json:"build_task"`
}

// TaskOutput is the ansible log output from the task
type TaskOutput struct {
	TaskID int       `db:"task_id" json:"task_id"`
	Task   string    `db:"task" json:"task"`
	Time   time.Time `db:"time" json:"time"`
	Output string    `db:"output" json:"output"`
}

type TaskStageType string

const (
	TaskStageRepositoryClone TaskStageType = "repository_clone"
	TaskStageTerraformPlan   TaskStageType = "terraform_plan"
	TaskStageTerraformApply  TaskStageType = "terraform_apply"
)

type TaskStage struct {
	TaskID        int           `db:"task_id" json:"task_id"`
	Start         *time.Time    `db:"start" json:"start"`
	End           *time.Time    `db:"end" json:"end"`
	StartOutputID *int          `db:"start_output_id" json:"start_output_id"`
	EndOutputID   *int          `db:"end_output_id" json:"end_output_id"`
	Type          TaskStageType `db:"type" json:"type"`
}
