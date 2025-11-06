package sql

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func validateTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("Tag cannot be empty")
	}

	return nil
}

func makePropsNonGlobal(props db.ObjectProps) (res db.ObjectProps) {
	res = props
	res.IsGlobal = false
	return
}

var runnerProps = makePropsNonGlobal(db.GlobalRunnerProps)

func (d *SqlDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	err = d.getObject(projectID, runnerProps, runnerID, &runner)
	return
}

func (d *SqlDb) GetRunners(projectID int, activeOnly bool, tag *string) (runners []db.Runner, err error) {
	if tag != nil {
		err = validateTag(*tag)
		if err != nil {
			return
		}
	}

	err = d.getObjects(projectID, runnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		if tag != nil {
			builder = builder.Where("tag=?", *tag)
		}

		if activeOnly {
			builder = builder.Where("active=?", activeOnly)
		}

		return builder
	}, &runners)
	return
}

func (d *SqlDb) DeleteRunner(projectID int, runnerID int) (err error) {
	err = d.deleteObject(projectID, runnerProps, runnerID)
	return
}

func (d *SqlDb) GetRunnerCount() (res int, err error) {
	query, args, err := squirrel.Select("count(*)").
		From("runner").
		Where(squirrel.NotEq{"project_id": nil}).
		ToSql()

	if err != nil {
		return
	}

	cnt, err := d.Sql().SelectInt(query, args...)

	res = int(cnt)

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
