package models

import "time"
import "github.com/ansible-semaphore/semaphore/database"

type Task struct {
	ID         int `db:"id" json:"id"`
	TemplateID int `db:"template_id" json:"template_id" binding:"required"`

	Status string `db:"status" json:"status"`
	Debug  bool   `db:"debug" json:"debug"`

	// override variables
	Playbook    string `db:"playbook" json:"playbook"`
	Environment string `db:"environment" json:"environment"`

	Created time.Time  `db:"created" json:"created"`
	Start   *time.Time `db:"start" json:"start"`
	End     *time.Time `db:"end" json:"end"`
}

type TaskOutput struct {
	TaskID int       `db:"task_id" json:"task_id"`
	Time   time.Time `db:"time" json:"time"`
	Output string    `db:"output" json:"output"`
}

func init() {
	database.Mysql.AddTableWithName(Task{}, "task").SetKeys(true, "id")
	database.Mysql.AddTableWithName(TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
}
