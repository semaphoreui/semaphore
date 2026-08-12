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
// The rename is done here rather than in SQL because a generated name can clash
// with a name which is already taken ("Build" twice next to a real "Build (2)"),
// which needs a retry that portable SQL cannot express.
func (m migration_2_20_2) PreApply(tx *gorp.Transaction) error {
	type templateName struct {
		ID        int    `db:"id"`
		ProjectID int    `db:"project_id"`
		Name      string `db:"name"`
	}

	var templates []templateName

	// Ordered by id so the oldest template of each name keeps it, and so the
	// result does not depend on the order rows come back in.
	_, err := tx.Select(&templates,
		m.db.PrepareQuery("select `id`, `project_id`, `name` from `project__template` order by `id`"))

	if err != nil {
		return err
	}

	taken := make(map[string]bool, len(templates))
	key := func(projectID int, name string) string {
		return strconv.Itoa(projectID) + "\x00" + name
	}

	// Every name in use is reserved before anything is renamed, so that a
	// generated name cannot take the name of a template which already has it:
	// "Build" twice next to a real "Build (2)" must not turn the latter into
	// "Build (2) (2)".
	var duplicates []templateName

	for _, template := range templates {
		if taken[key(template.ProjectID, template.Name)] {
			duplicates = append(duplicates, template)
			continue
		}

		taken[key(template.ProjectID, template.Name)] = true
	}

	for _, template := range duplicates {
		name := template.Name
		for i := 2; taken[key(template.ProjectID, name)]; i++ {
			name = template.Name + " (" + strconv.Itoa(i) + ")"
		}

		_, err = tx.Exec(
			m.db.PrepareQuery("update `project__template` set `name`=? where `id`=?"),
			name, template.ID)

		if err != nil {
			return err
		}

		taken[key(template.ProjectID, name)] = true
	}

	return nil
}
