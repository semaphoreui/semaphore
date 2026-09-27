package alerting

import (
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
)

// Payload is the data available to message templates. Field names are part
// of the public template contract (custom bodies reference them), so they
// only ever grow.
type Payload struct {
	// Name is the template name.
	Name   string
	Author string
	// Color is filled per channel from Channel.StatusColor.
	Color string
	Task  PayloadTask
	Chat  PayloadChat
	// Project is the owning project.
	Project PayloadProject
	// Playbook is the template playbook or script.
	Playbook string
	// ScheduleName is set when the task was started by a schedule.
	ScheduleName string
}

type PayloadTask struct {
	ID      string
	URL     string
	Result  string
	Desc    string
	Version string
	// Duration is the human readable run time, empty until the task started.
	Duration string
	// Trigger is manual, schedule, integration or api.
	Trigger string
	Status  task_logger.TaskStatus
}

type PayloadChat struct {
	ID       string
	ThreadID string
}

type PayloadProject struct {
	ID   int
	Name string
}

// PayloadSource is the subset of the store the payload builder needs.
type PayloadSource interface {
	GetUser(userID int) (db.User, error)
	GetTask(projectID int, taskID int) (db.Task, error)
	GetTemplate(projectID int, templateID int) (db.Template, error)
	GetSchedule(projectID int, scheduleID int) (db.Schedule, error)
}

// BuildPayload collects everything templates can reference. Lookups that
// fail degrade to empty strings: a missing user name must never block a
// notification.
func BuildPayload(
	cfg *util.ConfigType,
	source PayloadSource,
	project db.Project,
	template db.Template,
	task db.Task,
	status task_logger.TaskStatus,
) Payload {
	author := "—"
	if task.UserID != nil && source != nil {
		if user, err := source.GetUser(*task.UserID); err == nil {
			author = user.Name
		}
	}

	version := ""
	if task.Version != nil {
		version = *task.Version
	} else if template.Type != db.TemplateTask && source != nil {
		if v := incomingVersion(source, task); v != nil {
			version = "build " + *v
		}
	}

	scheduleName := ""
	if task.ScheduleID != nil && source != nil {
		if schedule, err := source.GetSchedule(task.ProjectID, *task.ScheduleID); err == nil {
			scheduleName = schedule.Name
		}
	}

	webHost := ""
	if cfg != nil {
		webHost = cfg.WebHost
	}

	return Payload{
		Name:   template.Name,
		Author: author,
		Task: PayloadTask{
			ID:       strconv.Itoa(task.ID),
			URL:      fmt.Sprintf("%s/project/%d/templates/%d?t=%d", webHost, template.ProjectID, template.ID, task.ID),
			Result:   status.Format(),
			Desc:     task.Message,
			Version:  version,
			Duration: taskDuration(task),
			Trigger:  taskTrigger(task),
			Status:   status,
		},
		Project: PayloadProject{
			ID:   project.ID,
			Name: project.Name,
		},
		Playbook:     template.Playbook,
		ScheduleName: scheduleName,
	}
}

// incomingVersion mirrors db.Task.GetIncomingVersion without requiring the
// full store: the version of the build task a deploy task was created from.
func incomingVersion(source PayloadSource, task db.Task) *string {
	if task.BuildTaskID == nil {
		return nil
	}
	buildTask, err := source.GetTask(task.ProjectID, *task.BuildTaskID)
	if err != nil {
		return nil
	}
	tpl, err := source.GetTemplate(task.ProjectID, buildTask.TemplateID)
	if err != nil {
		return nil
	}
	if tpl.Type == db.TemplateBuild {
		return buildTask.Version
	}
	return incomingVersion(source, buildTask)
}

func taskDuration(task db.Task) string {
	if task.Start == nil {
		return ""
	}
	end := task.End
	if end == nil {
		now := tz.Now()
		end = &now
	}
	return end.Sub(*task.Start).Round(1e9).String()
}

func taskTrigger(task db.Task) string {
	switch {
	case task.ScheduleID != nil:
		return "schedule"
	case task.IntegrationID != nil:
		return "integration"
	case task.UserID != nil:
		return "manual"
	default:
		return "api"
	}
}
