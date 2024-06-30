package db

type Schedule struct {
	ID         int    `db:"id" json:"id"`
	ProjectID  int    `db:"project_id" json:"project_id"`
	TemplateID int    `db:"template_id" json:"template_id"`
	CronFormat string `db:"cron_format" json:"cron_format"`
	Name       string `db:"name" json:"name"`
	Active     bool   `db:"active" json:"active"`

	LastCommitHash *string `db:"last_commit_hash" json:"-"`
	RepositoryID   *int    `db:"repository_id" json:"repository_id"`
}

type ScheduleWithTpl struct {
	Schedule
	TemplateName string `db:"tpl_name" json:"tpl_name"`
}
