package bolt

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	err = d.getObject(0, db.GlobalRunnerProps, intObjectID(runnerID), &runner)
	if err != nil {
		return
	}

	if runner.ProjectID == nil || *runner.ProjectID != projectID {
		err = db.ErrNotFound
	}

	return
}

func validateTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	return nil
}

// GetRunners returns the project's runners, optionally filtered by activity.
// Tag filtering is a SQL-only feature; the Bolt store is a development/test
// stand-in and ignores the tag argument apart from validation.
func (d *BoltDb) GetRunners(projectID int, activeOnly bool, tag *string) (runners []db.Runner, err error) {
	if tag != nil {
		err = validateTag(*tag)
		if err != nil {
			return
		}
	}

	runners = make([]db.Runner, 0)
	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		runner := i.(db.Runner)

		if runner.ProjectID == nil || *runner.ProjectID != projectID {
			return false
		}

		if activeOnly {
			return runner.Active
		}

		return true
	}, &runners)
	return
}

func (d *BoltDb) DeleteRunner(projectID int, runnerID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		runner, err := d.GetRunner(projectID, runnerID)
		if err != nil {
			return err
		}
		if runner.ProjectID == nil || *runner.ProjectID != projectID {
			return db.ErrNotFound
		}
		return d.deleteObject(0, db.GlobalRunnerProps, intObjectID(runnerID), tx)
	})
}

// GetRunnerTags is a stub for the Bolt store. Tag aggregation is implemented
// only in the SQL store; here we return an empty list so callers don't error.
func (d *BoltDb) GetRunnerTags(projectID int) ([]db.RunnerTag, error) {
	return []db.RunnerTag{}, nil
}

func (d *BoltDb) GetRunnerCount() (res int, err error) {
	runners := make([]db.Runner, 0)
	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		runner := i.(db.Runner)
		return runner.ProjectID != nil
	}, &runners)

	if err != nil {
		return 0, err
	}

	return len(runners), nil
}
