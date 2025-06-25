//go:build !pro

package bolt

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *BoltDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	err = db.ErrNotFound
	return
}

func (d *BoltDb) GetRunners(projectID int, activeOnly bool, tag *string) (runners []db.Runner, err error) {
	runners = make([]db.Runner, 0)
	return
}

func (d *BoltDb) DeleteRunner(projectID int, runnerID int) (err error) {
	err = db.ErrNotFound
	return
}

func (d *BoltDb) GetRunnerTags(projectID int) ([]db.RunnerTag, error) {
	return []db.RunnerTag{
		{
			Tag:             "tag1",
			NumberOfRunners: 1,
		},
	}, nil
}
