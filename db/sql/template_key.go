package sql

func (d *SqlDb) GetTemplateKeys(projectID int, templateID int) (keyIDs []int, err error) {
	keyIDs = make([]int, 0)

	var rows []struct {
		KeyID int `db:"key_id"`
	}

	_, err = d.selectAll(
		&rows,
		"select key_id from project__template_key "+
			"where project_id=? and template_id=? order by key_id",
		projectID,
		templateID,
	)

	if err != nil {
		return
	}

	for _, r := range rows {
		keyIDs = append(keyIDs, r.KeyID)
	}

	return
}

// UpdateTemplateKeys replaces the keys a template installs Galaxy requirements
// with.
//
// The keys are resolved in the project of the template first: the foreign key of
// key_id only points at access_key, so a key of another project would be stored
// happily and then fail to load, and the credentials of one project would be
// reachable from another.
//
// The replacement runs in one transaction so a failing insert can not leave the
// template with the old keys deleted and the new ones missing.
func (d *SqlDb) UpdateTemplateKeys(projectID int, templateID int, keyIDs []int) (err error) {
	seen := make(map[int]bool, len(keyIDs))
	unique := make([]int, 0, len(keyIDs))

	for _, keyID := range keyIDs {
		if seen[keyID] {
			continue
		}
		seen[keyID] = true

		if _, err = d.GetAccessKey(projectID, keyID); err != nil {
			return
		}

		unique = append(unique, keyID)
	}

	tx, err := d.Sql().Begin()

	if err != nil {
		return
	}

	_, err = tx.Exec(
		d.PrepareQuery("delete from project__template_key where project_id=? and template_id=?"),
		projectID,
		templateID,
	)

	if err != nil {
		_ = tx.Rollback()
		return
	}

	for _, keyID := range unique {
		_, err = tx.Exec(
			d.PrepareQuery("insert into project__template_key (project_id, template_id, key_id) values (?, ?, ?)"),
			projectID,
			templateID,
			keyID,
		)

		if err != nil {
			_ = tx.Rollback()
			return
		}
	}

	return tx.Commit()
}
