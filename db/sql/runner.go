//go:build !pro

package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	err = db.ErrNotFound
	return
}

func (d *SqlDb) GetRunners(projectID int, activeOnly bool, tag *string) (runners []db.Runner, err error) {
	runners = make([]db.Runner, 0)
	return
}

func (d *SqlDb) DeleteRunner(projectID int, runnerID int) (err error) {
	err = db.ErrNotFound
	return
}

func (d *SqlDb) GetRunnerTags(projectID int) (res []db.RunnerTag, err error) {
	query, args, err := squirrel.Select("tag").
		From("runner as r").
		Where(squirrel.Eq{"r.project_id": projectID}).
		Where(squirrel.NotEq{"r.tag": ""}).
		ToSql()

	if err != nil {
		return
	}

	runners := make([]db.Runner, 0)
	_, err = d.selectAll(&runners, query, args...)

	res = make([]db.RunnerTag, 0)
	for _, r := range runners {
		res = append(res, db.RunnerTag{
			Tag: r.Tag,
		})
	}

	return
}
