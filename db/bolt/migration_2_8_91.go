package bolt

type migration_2_8_91 struct {
	migration
}

func (d migration_2_8_91) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return
	}

	usersByProjectMap := make(map[string]map[string]map[string]interface{})

	for _, projectID := range projectIDs {
		usersByProjectMap[projectID], err = d.getObjects(projectID, "user")
		if err != nil {
			return
		}
	}

	for projectID, projectUsers := range usersByProjectMap {
		for userId, userData := range projectUsers {
			if userData["admin"] == true {
				userData["role"] = "owner"
			} else {
				userData["role"] = "manager"
			}
			delete(userData, "admin")
			err = d.setObject(projectID, "user", userId, userData)
			if err != nil {
				return
			}
		}
	}

	return
}
