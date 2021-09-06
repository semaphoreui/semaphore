package bolt

import "github.com/ansible-semaphore/semaphore/db"

func (d *BoltDb) GetSchedules() (schedules []db.Schedule, err error) {
	var allProjects []db.Project

	err = d.getObjects(0, db.ProjectProps, db.RetrieveQueryParams{}, nil, &allProjects)

	if err != nil {
		return
	}

	for _, proj := range allProjects {
		var projSchedules []db.Schedule
		projSchedules, err = d.GetProjectSchedules(proj.ID)
		if err != nil {
			return
		}
		schedules = append(schedules, projSchedules...)
	}

	return
}

func (d *BoltDb) GetProjectSchedules(projectID int) (schedules []db.Schedule, err error) {
	err = d.getObjects(projectID, db.ScheduleProps, db.RetrieveQueryParams{}, nil, &schedules)
	return
}


func (d *BoltDb) GetTemplateSchedules(projectID int, templateID int) (schedule db.Schedule, err error) {
	projSchedules, err := d.GetProjectSchedules(projectID)
	if err != nil {
		return
	}

	for _, s := range projSchedules {
		if s.TemplateID == templateID {
			schedule = s
			return
		}
	}

	err = db.ErrNotFound
	return
}

