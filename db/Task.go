package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"

	"github.com/go-gorp/gorp/v3"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type DefaultTaskParams struct {
}

type TerraformTaskParams struct {
	Plan        bool `json:"plan"`
	Destroy     bool `json:"destroy"`
	AutoApprove bool `json:"auto_approve"`
	Upgrade     bool `json:"upgrade"`
	Reconfigure bool `json:"reconfigure"`
}

type AnsibleTaskParams struct {
	Debug      bool     `json:"debug"`
	DebugLevel int      `json:"debug_level"`
	DryRun     bool     `json:"dry_run"`
	Diff       bool     `json:"diff"`
	Limit      []string `json:"limit"`
	Tags       []string `json:"tags"`
	SkipTags   []string `json:"skip_tags"`
}

// Task is a model of a task which will be executed by the runner
type Task struct {
	ID         int `db:"id" json:"id"`
	TemplateID int `db:"template_id" json:"template_id" binding:"required"`
	ProjectID  int `db:"project_id" json:"project_id"`

	Status task_logger.TaskStatus `db:"status" json:"status"`

	// override variables
	Playbook    string  `db:"playbook" json:"playbook"`
	Environment string  `db:"environment" json:"environment,omitempty"`
	Secret      string  `db:"-" json:"secret,omitempty"`
	Arguments   *string `db:"arguments" json:"arguments,omitempty"`
	GitBranch   *string `db:"git_branch" json:"git_branch,omitempty"`

	UserID        *int `db:"user_id" json:"user_id,omitempty"`
	IntegrationID *int `db:"integration_id" json:"integration_id,omitempty"`
	ScheduleID    *int `db:"schedule_id" json:"schedule_id,omitempty"`

	Created time.Time  `db:"created" json:"created"`
	Start   *time.Time `db:"start" json:"start,omitempty"`
	End     *time.Time `db:"end" json:"end,omitempty"`

	Message string `db:"message" json:"message,omitempty"`

	// CommitHash is a git commit hash of playbook repository which
	// was active when task was created.
	CommitHash *string `db:"commit_hash" json:"commit_hash,omitempty"`
	// CommitMessage contains message retrieved from git repository after checkout to CommitHash.
	// It is readonly by API.
	CommitMessage string `db:"commit_message" json:"commit_message,omitempty"`
	BuildTaskID   *int   `db:"build_task_id" json:"build_task_id,omitempty"`
	// Version is a build version.
	// This field available only for Build tasks.
	Version *string `db:"version" json:"version,omitempty"`

	InventoryID *int `db:"inventory_id" json:"inventory_id,omitempty"`

	Params MapStringAnyField `db:"params" json:"params,omitempty"`

	// Limit is deprecated, use Params.Limit instead
	Limit string `db:"-" json:"limit"`
}

func (task *Task) ExtractParams(target any) (err error) {
	content, err := json.Marshal(task.Params)
	if err != nil {
		return
	}
	err = json.Unmarshal(content, target)
	return
}

// PreInsert is a hook which is called before inserting task into database.
// Called directly in BoltDB implementation.
func (task *Task) PreInsert(gorp.SqlExecutor) error {
	task.Created = tz.In(task.Created)

	if _, ok := task.Params["limit"]; !ok {
		if task.Params == nil {
			task.Params = make(MapStringAnyField)
		}

		if task.Limit != "" {
			limits := strings.Split(task.Limit, ",")

			for i := range limits {
				limits[i] = strings.TrimSpace(limits[i])
			}

			task.Params["limit"] = limits
		}
	}

	return nil
}

func (task *Task) PreUpdate(gorp.SqlExecutor) error {
	if task.Start != nil {
		start := tz.In(*task.Start)
		task.Start = &start
	}

	if task.End != nil {
		end := tz.In(*task.End)
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

	var params any
	switch template.App {
	case AppAnsible:
		params = &AnsibleTaskParams{}
	case AppTerraform, AppTofu, AppTerragrunt:
		params = &TerraformTaskParams{}
	default:
		params = &DefaultTaskParams{}
	}

	return task.ExtractParams(params)
}

func (task *TaskWithTpl) Fill(d Store) error {
	if task.BuildTaskID != nil {
		build, err := d.GetTask(task.ProjectID, *task.BuildTaskID)
		if errors.Is(err, ErrNotFound) {
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
	TemplateType     TemplateType `db:"tpl_type" json:"tpl_type,omitempty"`
	TemplateApp      TemplateApp  `db:"tpl_app" json:"tpl_app,omitempty"`
	UserName         *string      `db:"user_name" json:"user_name,omitempty"`
	BuildTask        *Task        `db:"-" json:"build_task,omitempty"`
}

// TaskOutput is the ansible log output from the task
type TaskOutput struct {
	ID      int       `db:"id" json:"id"`
	TaskID  int       `db:"task_id" json:"task_id"`
	Time    time.Time `db:"time" json:"time"`
	Output  string    `db:"output" json:"output"`
	StageID *int      `db:"stage_id" json:"stage_id"`
}

type TaskStageType string

const (
	TaskStageInit          TaskStageType = "init"
	TaskStageTerraformPlan TaskStageType = "terraform_plan"
	TaskStageRunning       TaskStageType = "running"
	TaskStagePrintResult   TaskStageType = "print_result"
)

type TaskStage struct {
	ID     int           `db:"id" json:"id"`
	TaskID int           `db:"task_id" json:"task_id"`
	Start  *time.Time    `db:"start" json:"start"`
	End    *time.Time    `db:"end" json:"end"`
	Type   TaskStageType `db:"type" json:"type"`
}

type TaskStageWithResult struct {
	ID            int           `db:"id" json:"id"`
	TaskID        int           `db:"task_id" json:"task_id"`
	Start         *time.Time    `db:"start" json:"start"`
	End           *time.Time    `db:"end" json:"end"`
	StartOutputID *int          `db:"start_output_id" json:"start_output_id"`
	EndOutputID   *int          `db:"end_output_id" json:"end_output_id"`
	Type          TaskStageType `db:"type" json:"type"`
	JSON          string        `db:"json" json:"-"`
	Result        any           `db:"-" json:"result"`
}

type TaskStageResult struct {
	ID      int    `db:"id" json:"id"`
	TaskID  int    `db:"task_id" json:"task_id"`
	StageID int    `db:"stage_id" json:"stage_id"`
	JSON    string `db:"json" json:"json"`
}
