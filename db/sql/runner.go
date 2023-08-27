package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	return
}

func (d *SqlDb) GetRunners(projectID int) (runners []db.Runner, err error) {
	return
}

func (d *SqlDb) DeleteRunner(projectID int, runnerID int) (err error) {
	return
}

func (d *SqlDb) GetGlobalRunner(runnerID int) (runner db.Runner, err error) {
	return
}

func (d *SqlDb) GetGlobalRunners() (runners []db.Runner, err error) {
	return
}

func (d *SqlDb) DeleteGlobalRunner(runnerID int) (err error) {
	return
}

func (d *SqlDb) UpdateRunner(runner db.Runner) (err error) {
	return
}

func (d *SqlDb) CreateRunner(runner db.Runner) (newRunner db.Runner, err error) {
	return
}
