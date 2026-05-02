package sql

import (
	"encoding/base64"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/gorilla/securecookie"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) GetRunnerByToken(token string) (runner db.Runner, err error) {

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
	err = d.loadRunnerTagsSingle(&runner)
	return
}

func (d *SqlDb) GetGlobalRunner(runnerID int) (runner db.Runner, err error) {
	err = d.getObject(0, db.GlobalRunnerProps, runnerID, &runner)
	if err != nil {
		return
	}
	err = d.loadRunnerTagsSingle(&runner)
	return
}

func (d *SqlDb) GetAllRunners(activeOnly bool, globalOnly bool, tagFilterMode db.RunnerTagFilterMode, tag *string) (runners []db.Runner, err error) {
	if tag == nil && tagFilterMode == db.RunnerFilterTagCompleteMatch {
		err = fmt.Errorf("tag filter mode is complete match but no tag was provided")
		return
	}

	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {

		if globalOnly {
			builder = builder.Where("project_id is null")
		}

		if activeOnly {
			builder = builder.Where("active=?", activeOnly)
		}

		switch tagFilterMode {
		case db.RunnerFilterHasAnyTag:
			builder = builder.Where(runnerHasAnyTagExpr())
		case db.RunnerFilterIsDefault:
			builder = builder.Where(runnerIsDefaultExpr())
		case db.RunnerFilterIgnoreTags:
			// No tag filtering applied.
		case db.RunnerFilterTagCompleteMatch:
			builder = builder.Where(runnerHasTagExpr(*tag))
		default:
			panic("invalid tag filter mode: " + tagFilterMode)
		}

		return builder
	}, &runners)
	if err != nil {
		return
	}
	err = d.loadRunnerTags(runners)
	return
}

func (d *SqlDb) GetGlobalRunnerTags() (res []db.RunnerTag, err error) {
	query, args, err := squirrel.Select("rt.tag", "count(distinct rt.runner_id) as cnt").
		From("runner__tag rt").
		Join("runner r on r.id = rt.runner_id").
		Where("r.project_id is null").
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

func (d *SqlDb) DeleteGlobalRunner(runnerID int) (err error) {
	err = d.deleteObject(0, db.GlobalRunnerProps, runnerID)
	return
}

func (d *SqlDb) ClearRunnerCache(runner db.Runner) (err error) {
	if runner.ProjectID == nil {
		_, err = d.exec(
			"update `runner` set `cleaning_requested`=? where id=?",
			tz.Now(),
			runner.ID)
		return
	}

	_, err = d.exec(
		"update `runner` set `cleaning_requested`=? where id=? and project_id=?",
		tz.Now(),
		runner.ID,
		runner.ProjectID)

	return
}

func (d *SqlDb) TouchRunner(runner db.Runner) (err error) {
	if runner.ProjectID == nil {
		_, err = d.exec(
			"update `runner` set `touched`=? where id=?",
			tz.Now(),
			runner.ID)
		return
	}

	_, err = d.exec(
		"update `runner` set `touched`=? where id=? and project_id=?",
		tz.Now(),
		runner.ID,
		runner.ProjectID)

	return
}

func (d *SqlDb) UpdateRunner(runner db.Runner) (err error) {
	_, err = d.exec(
		"update `runner` set `name`=?, `active`=?, `is_default`=?, webhook=?, max_parallel_tasks=? where id=?",
		runner.Name,
		runner.Active,
		runner.IsDefault,
		runner.Webhook,
		runner.MaxParallelTasks,
		runner.ID)

	if err != nil {
		return
	}

	err = d.replaceRunnerTags(runner.ID, runner.Tags)
	return
}

func (d *SqlDb) CreateRunner(runner db.Runner) (newRunner db.Runner, err error) {
	token := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))

	insertID, err := d.insert(
		"id",
		"insert into `runner` (project_id, token, webhook, max_parallel_tasks, `name`, `active`, `is_default`, public_key) values (?, ?, ?, ?, ?, ?, ?, ?)",
		runner.ProjectID,
		token,
		runner.Webhook,
		runner.MaxParallelTasks,
		runner.Name,
		runner.Active,
		runner.IsDefault,
		runner.PublicKey)

	if err != nil {
		return
	}

	newRunner = runner
	newRunner.ID = insertID
	newRunner.Token = token
	newRunner.Tags = normalizeTags(runner.Tags)

	if err = d.replaceRunnerTags(newRunner.ID, newRunner.Tags); err != nil {
		return
	}

	return
}
