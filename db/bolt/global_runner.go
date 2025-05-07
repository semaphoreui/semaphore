package bolt

import (
	"encoding/base64"
	"github.com/gorilla/securecookie"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetRunnerByToken(token string) (runner db.Runner, err error) {

	runners := make([]db.Runner, 0)

	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		r := i.(db.Runner)
		return r.Token == token
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

func (d *BoltDb) GetGlobalRunner(runnerID int) (runner db.Runner, err error) {
	err = d.getObject(0, db.GlobalRunnerProps, intObjectID(runnerID), &runner)
	if err != nil {
		return
	}

	if runner.ProjectID != nil {
		err = db.ErrNotFound
	}

	return
}

func (d *BoltDb) GetAllRunners(activeOnly bool, globalOnly bool) (runners []db.Runner, err error) {
	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		runner := i.(db.Runner)

		if globalOnly && runner.ProjectID != nil {
			return false
		}

		if activeOnly {
			return runner.Active
		}

		return true
	}, &runners)
	return
}

func (d *BoltDb) DeleteGlobalRunner(runnerID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {

		var runner db.Runner
		err := d.getObject(0, db.GlobalRunnerProps, intObjectID(runnerID), &runner)

		if err != nil {
			return err
		}

		if runner.ProjectID != nil {
			return db.ErrNotFound
		}

		return d.deleteObject(0, db.GlobalRunnerProps, intObjectID(runnerID), tx)
	})
}

func (d *BoltDb) updateRunner(runner db.Runner, updater func(targetRunner *db.Runner, foundRunner db.Runner)) (err error) {
	return d.db.Update(func(tx *bbolt.Tx) error {
		var origRunner db.Runner

		err = d.getObjectTx(tx, 0, db.GlobalRunnerProps, intObjectID(runner.ID), &origRunner)

		if err != nil {
			return err
		}

		if runner.ProjectID == nil {
			if origRunner.ProjectID != nil {
				return db.ErrNotFound
			}
		} else {
			if *origRunner.ProjectID != *runner.ProjectID {
				return db.ErrNotFound
			}
		}

		updater(&runner, origRunner)

		return d.updateObjectTx(tx, 0, db.GlobalRunnerProps, runner)
	})
}

func (d *BoltDb) ClearRunnerCache(runner db.Runner) (err error) {
	return d.updateRunner(runner, func(targetRunner *db.Runner, foundRunner db.Runner) {
		now := tz.Now()
		targetRunner.CleaningRequested = &now
	})
}

func (d *BoltDb) TouchRunner(runner db.Runner) (err error) {
	return d.updateRunner(runner, func(targetRunner *db.Runner, foundRunner db.Runner) {
		now := tz.Now()
		targetRunner.Touched = &now
	})
}

func (d *BoltDb) UpdateRunner(runner db.Runner) (err error) {
	return d.updateRunner(runner, func(targetRunner *db.Runner, foundRunner db.Runner) {
		targetRunner.PublicKey = foundRunner.PublicKey
		targetRunner.Token = foundRunner.Token
	})
}

func (d *BoltDb) CreateRunner(runner db.Runner) (newRunner db.Runner, err error) {
	runner.Token = base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	res, err := d.createObject(0, db.GlobalRunnerProps, runner)

	if err != nil {
		return
	}
	newRunner = res.(db.Runner)
	return
}
