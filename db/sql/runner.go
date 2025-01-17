//go:build !pro

package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	err = db.ErrNotFound
	return
}

func (d *SqlDb) GetRunners(projectID int, activeOnly bool) (runners []db.Runner, err error) {
	runners = make([]db.Runner, 0)
	return
}

func (d *SqlDb) DeleteRunner(projectID int, runnerID int) (err error) {
	err = db.ErrNotFound
	return
}
