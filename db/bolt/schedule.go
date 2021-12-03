package bolt

import "github.com/neo1908/semaphore/db"

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

func (d *BoltDb) GetTemplateSchedules(projectID int, templateID int) (schedules []db.Schedule, err error) {
	schedules = make([]db.Schedule, 0)

	projSchedules, err := d.GetProjectSchedules(projectID)
	if err != nil {
		return
	}

	for _, s := range projSchedules {
		if s.TemplateID == templateID {
			schedules = append(schedules, s)
		}
	}

	return
}

func (d *BoltDb) CreateSchedule(schedule db.Schedule) (newSchedule db.Schedule, err error) {
	newTpl, err := d.createObject(schedule.ProjectID, db.ScheduleProps, schedule)
	if err != nil {
		return
	}
	newSchedule = newTpl.(db.Schedule)
	return
}

func (d *BoltDb) UpdateSchedule(schedule db.Schedule) error {
	return d.updateObject(schedule.ProjectID, db.ScheduleProps, schedule)
}

func (d *BoltDb) GetSchedule(projectID int, scheduleID int) (schedule db.Schedule, err error) {
	err = d.getObject(projectID, db.ScheduleProps, intObjectID(scheduleID), &schedule)
	return
}

func (d *BoltDb) DeleteSchedule(projectID int, scheduleID int) error {
	return d.deleteObject(projectID, db.ScheduleProps, intObjectID(scheduleID))
}
