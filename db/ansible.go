package db

import "time"

type AnsibleTaskHost struct {
	ID          int       `json:"id" db:"id"`
	TaskID      int       `json:"task_id" db:"task_id"`
	ProjectID   int       `json:"project_id" db:"project_id"`
	Host        string    `json:"host" db:"host"`
	Changed     int       `json:"changed" db:"changed"`
	Failed      int       `json:"failed" db:"failed"`
	Ignored     int       `json:"ignored" db:"ignored"`
	Ok          int       `json:"ok" db:"ok"`
	Rescued     int       `json:"rescued" db:"rescued"`
	Skipped     int       `json:"skipped" db:"skipped"`
	Unreachable int       `json:"unreachable" db:"unreachable"`
	Created     time.Time `db:"created" json:"created"`
}

type AnsibleTaskError struct {
	ID        int       `json:"id" db:"id"`
	TaskID    int       `json:"task_id" db:"task_id"`
	ProjectID int       `json:"project_id" db:"project_id"`
	Host      string    `json:"host" db:"host"`
	Task      string    `json:"task" db:"task"`
	Error     string    `json:"error" db:"error"`
	Created   time.Time `db:"created" json:"created"`
}
