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

func (d *SqlDb) UpdateTemplateKeys(projectID int, templateID int, keyIDs []int) (err error) {
	_, err = d.exec(
		"delete from project__template_key where project_id=? and template_id=?",
		projectID,
		templateID,
	)
	if err != nil {
		return
	}

	seen := make(map[int]bool)
	for _, keyID := range keyIDs {
		if seen[keyID] {
			continue
		}
		seen[keyID] = true

		_, err = d.exec(
			"insert into project__template_key (project_id, template_id, key_id) values (?, ?, ?)",
			projectID,
			templateID,
			keyID,
		)
		if err != nil {
			return
		}
	}

	return
}
