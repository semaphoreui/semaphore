package bolt

type migration_2_10_33 struct {
	migration
}

func (d migration_2_10_33) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return
	}

	vaults := make(map[string]map[string]map[string]interface{})

	for _, projectID := range projectIDs {
		var err2 error
		vaults[projectID], err2 = d.getObjects(projectID, "template_vault")
		if err2 != nil {
			return err2
		}
	}

	for projectID, projectVaults := range vaults {
		for repoID, vault := range projectVaults {
			if vault["type"] != nil && vault["type"] != "" {
				continue
			}
			vault["type"] = "password"
			err = d.setObject(projectID, "template_vault", repoID, vault)
			if err != nil {
				return err
			}
		}
	}

	return
}
