package db

import (
	"fmt"
	"time"
)

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

	UserID *int `db:"user_id" json:"user_id"`

	Created time.Time  `db:"created" json:"created"`
	Start   *time.Time `db:"start" json:"start"`
	End     *time.Time `db:"end" json:"end"`

	Message string `db:"message" json:"message"`

	CommitHash *string `db:"commit_hash" json:"commit_hash"`
	// CommitMessage contains message retrieved from git repository after checkout to CommitHash.
	// It is readonly by API.
	CommitMessage string `db:"commit_message" json:"commit_message"`

	BuildTaskID *int `db:"build_task_id" json:"build_task_id"`

	// Version is a build version.
	// This field available only for Build tasks.
	Version *string `db:"version" json:"version"`
}

func (task *Task) GetVersion(d Store) (string, error) {
	tpl, err := d.GetTemplate(task.ProjectID, task.TemplateID)
	if err != nil {
		return "", err
	}

	switch tpl.Type {
	case TemplateTask:
		return "", fmt.Errorf("only build and deploy tasks has versions")
	case TemplateBuild:
		if task.Version == nil {
			return "", fmt.Errorf("build task must have version")
		}
		return *task.Version, nil
	case TemplateDeploy:
		var buildTask Task
		buildTask, err = d.GetTask(task.ProjectID, *task.BuildTaskID)
		if err != nil {
			return "", err
		}
		return buildTask.GetVersion(d)
	default:
		return "", fmt.Errorf("unknown task type")
	}
}

func (task *Task) ValidateNewTask(template Template) error {
	switch template.Type {
	case TemplateBuild:
	case TemplateDeploy:
	case TemplateTask:
	}
	return nil
}

func (task *TaskWithTpl) Fill(d Store) error {
	if task.BuildTaskID != nil {
		build, err := d.GetTask(task.ProjectID, *task.BuildTaskID)
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
