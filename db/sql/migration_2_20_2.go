package sql

import (
	"strconv"

	"github.com/go-gorp/gorp/v3"
)

type migration_2_20_2 struct {
	db *SqlDb
}

// PreApply renames templates which share a name inside a project, so that
// v2.20.2.sql can put a unique index on (project_id, name). Template names were
// never unique before, so any installation may hold duplicates and the index
// would otherwise fail the upgrade.
//
// Which names collide is decided by the database rather than by comparing them
// here: the unique index uses the collation of the column, and on the default
// MySQL and MariaDB collations "Build", "build" and "Build " are all the same
// key while Go sees three different strings. Comparing in Go would leave those
// rows in place and the index would still fail.
//
// The rename is a loop rather than a single statement because a generated name
// can itself be taken ("Build" twice next to a real "Build (2)").
func (m migration_2_20_2) PreApply(tx *gorp.Transaction) error {
	type templateName struct {
		ID        int    `db:"id"`
		ProjectID int    `db:"project_id"`
		Name      string `db:"name"`
	}

	var duplicates []templateName

	// Every template which an older one of the project already shadows. Ordered
	// by id so the oldest keeps its name and the result does not depend on the
	// order rows come back in.
	_, err := tx.Select(&duplicates, m.db.PrepareQuery(
		"select `id`, `project_id`, `name` from `project__template` t "+
			"where exists (select 1 from (select `id`, `project_id`, `name` from `project__template`) o "+
			"where o.`project_id` = t.`project_id` and o.`name` = t.`name` and o.`id` < t.`id`) "+
			"order by t.`id`"))

	if err != nil {
		return err
	}

	for _, template := range duplicates {
		name, err := m.freeTemplateName(tx, template.ProjectID, template.Name)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			m.db.PrepareQuery("update `project__template` set `name`=? where `id`=?"),
			name, template.ID)

		if err != nil {
			return err
		}
	}

	return nil
}

// freeTemplateName returns a name based on base which no template of the project
// uses. Availability is asked of the database so that the answer follows the
// same collation as the unique index.
func (m migration_2_20_2) freeTemplateName(tx *gorp.Transaction, projectID int, base string) (string, error) {
	for i := 2; ; i++ {
		name := base + " (" + strconv.Itoa(i) + ")"

		count, err := tx.SelectInt(m.db.PrepareQuery(
			"select count(*) from `project__template` where `project_id`=? and `name`=?"),
			projectID, name)

		if err != nil {
			return "", err
		}

		if count == 0 {
			return name, nil
		}
	}
}
