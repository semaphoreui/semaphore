package models

import "time"

type Task struct {
	ID          int    `db:"id" json:"id"`
	TemplateID  int    `db:"template_id" json:"template_id"`
	Status      string `db:"status" json:"status"`
	Playbook    string `db:"playbook" json:"playbook"`
	Environment string `db:"environment" json:"environment"`
}

type TaskOutput struct {
	TaskID int       `db:"task_id" json:"task_id"`
	Time   time.Time `db:"time" json:"time"`
	Output string    `db:"output" json:"output"`
}
