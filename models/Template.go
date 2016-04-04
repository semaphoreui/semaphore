package models

type Template struct {
	ID            int    `db:"id" json:"id"`
	SshKeyID      int    `db:"ssh_key_id" json:"ssh_key_id"`
	ProjectID     int    `db:"project_id" json:"project_id"`
	InventoryID   int    `db:"inventory_id" json:"inventory_id"`
	RepositoryID  int    `db:"repository_id" json:"repository_id"`
	EnvironmentID *int   `db:"environment_id" json:"environment_id"`
	Playbook      string `db:"playbook" json:"playbook"`
}

type TemplateSchedule struct {
	TemplateID int    `db:"template_id" json:"template_id"`
	CronFormat string `db:"cron_format" json:"cron_format"`
}
