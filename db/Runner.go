package db

import "time"

type RunnerState string

type Runner struct {
	ID                int        `db:"id" json:"id"`
	Token             string     `db:"token" json:"-"`
	ProjectID         *int       `db:"project_id" json:"project_id"`
	Webhook           string     `db:"webhook" json:"webhook"`
	MaxParallelTasks  int        `db:"max_parallel_tasks" json:"max_parallel_tasks"`
	Active            bool       `db:"active" json:"active"`
	Name              string     `db:"name" json:"name"`
	Tag               string     `db:"tag" json:"tag"`
	Touched           *time.Time `db:"touched" json:"touched"`
	CleaningRequested *time.Time `db:"cleaning_requested" json:"cleaning_requested"`

	PublicKey *string `db:"public_key" json:"-"`
}
