package db

type RunnerState string

//const (
//	RunnerOffline RunnerState = "offline"
//	RunnerActive  RunnerState = "active"
//)

type Runner struct {
	ID        int    `db:"id" json:"-"`
	Token     string `db:"token" json:"-"`
	ProjectID *int   `db:"project_id" json:"project_id"`
	//State            RunnerState `db:"state" json:"state"`
	Integration      string `db:"integration" json:"integration"`
	MaxParallelTasks int    `db:"max_parallel_tasks" json:"max_parallel_tasks"`
}
