package db

import "time"

const (
	HostStatusOk          = "ok"
	HostStatusChanged     = "changed"
	HostStatusFailed      = "failed"
	HostStatusUnreachable = "unreachable"
)

// Host is a row of the hosts list and the identity of the card, within a project the counters are those of the project.
type Host struct {
	ID                int               `db:"id" json:"id"`
	Name              string            `db:"name" json:"name"`
	DisplayName       string            `db:"display_name" json:"display_name"`
	FirstSeen         time.Time         `db:"first_seen" json:"first_seen"`
	LastSeen          time.Time         `db:"last_seen" json:"last_seen"`
	LastAutomated     *time.Time        `db:"last_automated" json:"last_automated"`
	Runs              int               `db:"runs" json:"runs"`
	LastStatus        string            `db:"last_status" json:"last_status"`
	LastTaskID        *int              `db:"last_task_id" json:"last_task_id"`
	LastProjectID     *int              `db:"last_project_id" json:"last_project_id"`
	Decommissioned    *time.Time        `db:"decommissioned" json:"decommissioned"`
	DecommissionCount int               `db:"decommission_count" json:"-"`
	LastTemplate      *HostTemplateRef  `db:"-" json:"last_template"`
	Templates         []HostTemplateRef `db:"-" json:"templates"`
	Projects          []HostProjectRef  `db:"-" json:"projects"`
}

type HostTemplateRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type HostProjectRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type HostDetails struct {
	Host
	FailedInLast10 int `json:"failed_in_last_10"`
}

// HostTask is a row of the host card Tasks tab: a task and its result on this host.
type HostTask struct {
	TaskID       int        `db:"task_id" json:"task_id"`
	ProjectID    int        `db:"project_id" json:"project_id"`
	ProjectName  string     `db:"project_name" json:"project_name"`
	TemplateID   int        `db:"template_id" json:"template_id"`
	TemplateName string     `db:"template_name" json:"template_name"`
	Status       string     `db:"status" json:"status"`
	Start        *time.Time `db:"start" json:"start"`
	UserName     *string    `db:"user_name" json:"user_name"`
	Ok           int        `db:"ok" json:"ok"`
	Changed      int        `db:"changed" json:"changed"`
	Failed       int        `db:"failed" json:"failed"`
	Unreachable  int        `db:"unreachable" json:"unreachable"`
	Skipped      int        `db:"skipped" json:"skipped"`
	Rescued      int        `db:"rescued" json:"rescued"`
	Ignored      int        `db:"ignored" json:"ignored"`
}

type HostListParams struct {
	Search      string
	Status      string
	NotSeenDays int
	TemplateID  int
	Offset      int
	Count       int
}

// HostRepository reads the registry of hosts seen in Ansible runs.
type HostRepository interface {
	// GetHosts lists one page, a nil projectID means all projects, Templates are filled within a project only.
	GetHosts(projectID *int, params HostListParams) (hosts []Host, total int, err error)
	// GetHost returns ErrNotFound if the host does not exist or never touched the project.
	GetHost(hostID int, projectID *int) (HostDetails, error)
	// GetHostTasks lists tasks newest first, a nil projectID means all projects.
	GetHostTasks(hostID int, projectID *int, count int, beforeTaskID int) ([]HostTask, error)
}
