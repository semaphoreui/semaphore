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

// runnerHasTagExpr renders a parameterised EXISTS clause that checks
// whether the runner row identified by `pe.id` has the given tag.
// runner__tag.tag is indexed, so this is O(log N) per row.
func runnerHasTagExpr(tag string) squirrel.Sqlizer {
	return squirrel.Expr(
		"exists (select 1 from runner__tag rt where rt.runner_id = pe.id and rt.tag = ?)",
		tag,
	)
}

// runnerHasNoTagsExpr matches runners with an empty tag set.
func runnerHasNoTagsExpr() squirrel.Sqlizer {
	return squirrel.Expr(
		"not exists (select 1 from runner__tag rt where rt.runner_id = pe.id)",
	)
}

// runnerHasAnyTagExpr matches runners whose tag set is non-empty.
func runnerHasAnyTagExpr() squirrel.Sqlizer {
	return squirrel.Expr(
		"exists (select 1 from runner__tag rt where rt.runner_id = pe.id)",
	)
}

func (d *SqlDb) GetRunner(projectID int, runnerID int) (runner db.Runner, err error) {
	err = d.getObject(projectID, runnerProps, runnerID, &runner)
	if err != nil {
		return
	}
	err = d.loadRunnerTagsSingle(&runner)
	return
}

func (d *SqlDb) GetRunners(projectID int, activeOnly bool, tagFilterMode db.RunnerTagFilterMode, tag *string) (runners []db.Runner, err error) {
	if tag == nil && tagFilterMode == db.RunnerFilterTagCompleteMatch {
		err = fmt.Errorf("tag filter mode is complete match but no tag was provided")
		return
	}

	err = d.getObjects(projectID, runnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		switch tagFilterMode {
		case db.RunnerFilterTagCompleteMatch:
			builder = builder.Where(runnerHasTagExpr(*tag))
		case db.RunnerFilterHasNoTags:
			builder = builder.Where(runnerHasNoTagsExpr())
		case db.RunnerFilterIgnoreTags:
			// No tag filtering applied.
		default:
			panic("invalid tag filter mode for GetRunners: " + tagFilterMode)
		}

		if activeOnly {
			builder = builder.Where("active=?", activeOnly)
		}

		return builder
	}, &runners)
	if err != nil {
		return
	}
	err = d.loadRunnerTags(runners)
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
	query, args, err := squirrel.Select("rt.tag", "count(distinct rt.runner_id) as cnt").
		From("runner__tag rt").
		Join("runner r on r.id = rt.runner_id").
		Where(squirrel.Or{
			squirrel.Eq{"r.project_id": projectID},
			squirrel.Eq{"r.project_id": nil},
		}).
		GroupBy("rt.tag").
		ToSql()

	if err != nil {
		return
	}

	type row struct {
		Tag string `db:"tag"`
		Cnt int    `db:"cnt"`
	}

	rows := make([]row, 0)
	_, err = d.selectAll(&rows, query, args...)
	if err != nil {
		return
	}

	res = make([]db.RunnerTag, 0, len(rows))
	for _, r := range rows {
		res = append(res, db.RunnerTag{
			Tag:             r.Tag,
			NumberOfRunners: r.Cnt,
		})
	}

	return
}
