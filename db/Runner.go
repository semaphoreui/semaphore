package db

type RunnerState string

const (
	RunnerOffline RunnerState = "offline"
	RunnerActive  RunnerState = "active"
)

type Runner struct {
	ID        int         `db:"id" json:"-"`
	Token     string      `db:"token" json:"-"`
	ProjectID *int        `db:"project_id" json:"project_id"`
	State     RunnerState `db:"state" json:"state"`
}
