package bolt

type migration_2_8_40 struct {
	migration
}

func (d migration_2_8_40) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return
	}

	templates := make(map[string]map[string]map[string]interface{})

	for _, projectID := range projectIDs {
		var err2 error
		templates[projectID], err2 = d.getObjects(projectID, "template")
		if err2 != nil {
			return err2
		}
	}

	for projectID, projectTemplates := range templates {
		for repoID, tpl := range projectTemplates {
			tpl["name"] = tpl["alias"]
			delete(tpl, "alias")
			err = d.setObject(projectID, "template", repoID, tpl)
			if err != nil {
				return err
			}
		}
	}

	return
}
