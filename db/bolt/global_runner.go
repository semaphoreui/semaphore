package bolt

import (
	"encoding/base64"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetGlobalRunnerByToken(token string) (runner db.Runner, err error) {

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

func (d *BoltDb) DeleteGlobalRunner(runnerID int) (err error) {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteObject(0, db.GlobalRunnerProps, intObjectID(runnerID), tx)
	})
}

func (d *BoltDb) updateRunner(runner db.Runner, updater func(targetRunner *db.Runner, foundRunner db.Runner)) (err error) {
	var origRunner db.Runner

	if runner.ProjectID == nil {
		origRunner, err = d.GetGlobalRunner(runner.ID)
	} else {
		err = d.getObject(*runner.ProjectID, db.RunnerProps, intObjectID(runner.ID), &origRunner)
	}

	if err != nil {
		return
	}

	updater(&runner, origRunner)

	if runner.ProjectID == nil {
		return d.updateObject(0, db.GlobalRunnerProps, runner)
	} else {
		return d.updateObject(*runner.ProjectID, db.RunnerProps, runner)
	}
}

func (d *BoltDb) ClearRunnerCache(runner db.Runner) (err error) {
	return d.updateRunner(runner, func(targetRunner *db.Runner, foundRunner db.Runner) {
		now := time.Now().UTC()
		targetRunner.CleaningRequested = &now
	})
}

func (d *BoltDb) TouchRunner(runner db.Runner) (err error) {
	return d.updateRunner(runner, func(targetRunner *db.Runner, foundRunner db.Runner) {
		now := time.Now().UTC()
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

	bucketID := 0
	props := db.GlobalRunnerProps

	if runner.ProjectID != nil {
		bucketID = *runner.ProjectID
		props = db.RunnerProps
	}

	res, err := d.createObject(bucketID, props, runner)

	if err != nil {
		return
	}
	newRunner = res.(db.Runner)
	return
}
