package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetSchedules() (schedules []db.Schedule, err error) {
	var allProjects []db.Project

	err = d.getObjects(0, db.ProjectProps, db.RetrieveQueryParams{}, nil, &allProjects)

	if err != nil {
		return
	}

	for _, proj := range allProjects {
		var projSchedules []db.Schedule
		projSchedules, err = d.getProjectSchedules(proj.ID, nil)
		if err != nil {
			return
		}
		schedules = append(schedules, projSchedules...)
	}

	return
}

func (d *BoltDb) getProjectSchedules(projectID int, filter func(referringObj db.Schedule) bool) (schedules []db.Schedule, err error) {
	schedules = []db.Schedule{}
	err = d.getObjects(projectID, db.ScheduleProps, db.RetrieveQueryParams{}, func(referringObj interface{}) bool {
		return filter == nil || filter(referringObj.(db.Schedule))
	}, &schedules)
	return
}

func (d *BoltDb) GetProjectSchedules(projectID int) (schedules []db.ScheduleWithTpl, err error) {
	schedules = []db.ScheduleWithTpl{}

	orig, err := d.getProjectSchedules(projectID, func(s db.Schedule) bool {
		return s.RepositoryID == nil
	})

	if err != nil {
		return
	}

	for _, s := range orig {
		var tpl db.Template
		tpl, err = d.GetTemplate(projectID, s.TemplateID)
		if err != nil {
			return
		}
		schedules = append(schedules, db.ScheduleWithTpl{
			Schedule:     s,
			TemplateName: tpl.Name,
		})
	}

	return
}

func (d *BoltDb) GetTemplateSchedules(projectID int, templateID int) (schedules []db.Schedule, err error) {
	schedules, err = d.getProjectSchedules(projectID, func(s db.Schedule) bool {
		return s.TemplateID == templateID
	})

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

func (d *BoltDb) deleteSchedule(projectID int, scheduleID int, tx *bbolt.Tx) error {
	return d.deleteObject(projectID, db.ScheduleProps, intObjectID(scheduleID), tx)
}

func (d *BoltDb) DeleteSchedule(projectID int, scheduleID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteSchedule(projectID, scheduleID, tx)
	})
}

func (d *BoltDb) SetScheduleActive(projectID int, scheduleID int, active bool) error {
	schedule, err := d.GetSchedule(projectID, scheduleID)
	if err != nil {
		return err
	}
	schedule.Active = active
	return d.updateObject(projectID, db.ScheduleProps, schedule)
}

func (d *BoltDb) SetScheduleCommitHash(projectID int, scheduleID int, hash string) error {
	schedule, err := d.GetSchedule(projectID, scheduleID)
	if err != nil {
		return err
	}
	schedule.LastCommitHash = &hash
	return d.updateObject(projectID, db.ScheduleProps, schedule)
}
