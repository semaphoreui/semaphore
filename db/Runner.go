package db

type RunnerState string

//const (
//	RunnerOffline RunnerState = "offline"
//	RunnerActive  RunnerState = "active"
//)

type Runner struct {
	ID        int    `db:"id" json:"id"`
	Token     string `db:"token" json:"-"`
	ProjectID *int   `db:"project_id" json:"project_id"`
	//State            RunnerState `db:"state" json:"state"`
	Webhook          string `db:"webhook" json:"webhook"`
	MaxParallelTasks int    `db:"max_parallel_tasks" json:"max_parallel_tasks"`
	Active           bool   `db:"active" json:"active"`
	Name             string `db:"name" json:"name"`
}
