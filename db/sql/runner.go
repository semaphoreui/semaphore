package sql

import (
	"encoding/base64"
	"github.com/Masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/securecookie"
)

func (d *SqlDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	return
}

func (d *SqlDb) GetRunners(projectID int) (runners []db.Runner, err error) {
	return
}

func (d *SqlDb) DeleteRunner(projectID int, runnerID int) (err error) {
	return
}

func (d *SqlDb) GetGlobalRunnerByToken(token string) (runner db.Runner, err error) {

	runners := make([]db.Runner, 0)

	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		return builder.Where("token=?", token)
	}, &runners)

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = db.ErrNotFound
		return
	}

	runner = runners[0]
	return
}

func (d *SqlDb) GetGlobalRunner(runnerID int) (runner db.Runner, err error) {
	err = d.getObject(0, db.GlobalRunnerProps, runnerID, &runner)
	return
}

func (d *SqlDb) GetGlobalRunners(activeOnly bool) (runners []db.Runner, err error) {
	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		if activeOnly {
			builder = builder.Where("active=?", activeOnly)
		}

		return builder
	}, &runners)
	return
}

func (d *SqlDb) DeleteGlobalRunner(runnerID int) (err error) {
	err = d.deleteObject(0, db.GlobalRunnerProps, runnerID)
	return
}

func (d *SqlDb) UpdateRunner(runner db.Runner) (err error) {
	_, err = d.exec(
		"update runner set name=?, active=?, webhook=?, max_parallel_tasks=? where id=?",
		runner.Name,
		runner.Active,
		runner.Webhook,
		runner.MaxParallelTasks,
		runner.ID)

	return
}

func (d *SqlDb) CreateRunner(runner db.Runner) (newRunner db.Runner, err error) {
	token := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	insertID, err := d.insert(
		"id",
		"insert into runner (project_id, token, webhook, max_parallel_tasks, name, active) values (?, ?, ?, ?, ?, ?)",
		runner.ProjectID,
		token,
		runner.Webhook,
		runner.MaxParallelTasks,
		runner.Name,
		runner.Active)

	if err != nil {
		return
	}

	newRunner = runner
	newRunner.ID = insertID
	newRunner.Token = token
	return
}
