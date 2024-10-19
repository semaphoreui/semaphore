package bolt

type migration_2_10_12 struct {
	migration
}

func (d migration_2_10_12) Apply() error {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return err
	}

	for _, projectID := range projectIDs {
		schedules, err := d.getObjects(projectID, "schedule")
		if err != nil {
			return err
		}

		for scheduleID, schedule := range schedules {
			schedule["active"] = true
			err = d.setObject(projectID, "schedule", scheduleID, schedule)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
