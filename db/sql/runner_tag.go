package sql

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

// loadRunnerTags fills the Tags slice on each runner using a single
// `WHERE runner_id IN (...)` query, so listing N runners costs 2 queries
// regardless of how many tags any of them has.
func (d *SqlDb) loadRunnerTags(runners []db.Runner) error {
	if len(runners) == 0 {
		return nil
	}

	ids := make([]int, 0, len(runners))
	indexByID := make(map[int]int, len(runners))
	for i := range runners {
		runners[i].Tags = nil
		ids = append(ids, runners[i].ID)
		indexByID[runners[i].ID] = i
	}

	query, args, err := squirrel.Select("runner_id", "tag").
		From("runner__tag").
		Where(squirrel.Eq{"runner_id": ids}).
		ToSql()
	if err != nil {
		return err
	}

	type row struct {
		RunnerID int    `db:"runner_id"`
		Tag      string `db:"tag"`
	}

	var rows []row
	if _, err = d.selectAll(&rows, query, args...); err != nil {
		return err
	}

	for _, r := range rows {
		i, ok := indexByID[r.RunnerID]
		if !ok {
			continue
		}
		runners[i].Tags = append(runners[i].Tags, r.Tag)
	}

	return nil
}

// loadRunnerTagsSingle fills Tags on one runner.
func (d *SqlDb) loadRunnerTagsSingle(runner *db.Runner) error {
	rs := []db.Runner{*runner}
	if err := d.loadRunnerTags(rs); err != nil {
		return err
	}
	runner.Tags = rs[0].Tags
	return nil
}

// replaceRunnerTags deletes-then-reinserts the tag set for a runner.
// Caller is responsible for transactional context (we run two cheap
// statements; concurrent writes to the same runner row are not expected
// — runner edits are admin-driven and rare).
func (d *SqlDb) replaceRunnerTags(runnerID int, tags []string) error {
	if _, err := d.exec("delete from runner__tag where runner_id=?", runnerID); err != nil {
		return err
	}

	tags = normalizeTags(tags)
	if len(tags) == 0 {
		return nil
	}

	// Single multi-row insert keeps the round-trips down for runners
	// with many tags.
	values := make([]string, 0, len(tags))
	args := make([]any, 0, len(tags)*2)
	for _, t := range tags {
		values = append(values, "(?, ?)")
		args = append(args, runnerID, t)
	}

	q := "insert into runner__tag (runner_id, tag) values " + strings.Join(values, ", ")
	_, err := d.exec(q, args...)
	return err
}

// normalizeTags trims, drops empty, and dedupes (preserving order).
func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
