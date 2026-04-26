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
	// Project runners (scoped to this project) plus global runners (project_id IS NULL)
	// both contribute tags here so the template/inventory tag autocomplete sees every
	// runner that could be selected for this project's tasks.
	query, args, err := squirrel.Select("tag").
		From("runner as r").
		Where(squirrel.Or{
			squirrel.Eq{"r.project_id": projectID},
			squirrel.Eq{"r.project_id": nil},
		}).
		Where(squirrel.NotEq{"r.tag": ""}).
		ToSql()

	if err != nil {
		return
	}

	runners := make([]db.Runner, 0)
	_, err = d.selectAll(&runners, query, args...)
	if err != nil {
		return
	}

	tagMap := make(map[string]int)
	for _, r := range runners {
		tagMap[r.Tag]++
	}

	res = make([]db.RunnerTag, 0, len(tagMap))
	for tag, count := range tagMap {
		res = append(res, db.RunnerTag{
			Tag:             tag,
			NumberOfRunners: count,
		})
	}

	return
}
