package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) CreateSchedule(schedule db.Schedule) (newSchedule db.Schedule, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__schedule (project_id, template_id, cron_format, repository_id)"+
			"values (?, ?, ?, ?)",
		schedule.ProjectID,
		schedule.TemplateID,
		schedule.CronFormat,
		schedule.RepositoryID)

	if err != nil {
		return
	}

	newSchedule = schedule
	newSchedule.ID = insertID

	return
}

func (d *SqlDb) SetScheduleLastCommitHash(projectID int, scheduleID int, lastCommentHash string) error {
	_, err := d.exec("update project__schedule set "+
		"last_commit_hash=? "+
		"where project_id=? and id=?",
		lastCommentHash,
		projectID,
		scheduleID)
	return err
}

func (d *SqlDb) UpdateSchedule(schedule db.Schedule) error {
	_, err := d.exec("update project__schedule set "+
		"cron_format=? "+
		"repository_id=?, "+
		"where project_id=? and id=?",
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.ProjectID,
		schedule.ID)
	return err
}

func (d *SqlDb) GetSchedule(projectID int, scheduleID int) (template db.Schedule, err error) {
	err = d.selectOne(
		&template,
		"select * from project__schedule where project_id=? and id=?",
		projectID,
		scheduleID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) DeleteSchedule(projectID int, scheduleID int) error {
	_, err := d.exec("delete from project__schedule where project_id=? and id=?", projectID, scheduleID)
	return err
}

func (d *SqlDb) GetSchedules() (schedules []db.Schedule, err error) {
	_, err = d.selectAll(&schedules, "select * from project__schedule where cron_format != ''")
	return
}

func (d *SqlDb) GetTemplateSchedules(projectID int, templateID int) (schedules []db.Schedule, err error) {
	_, err = d.selectAll(&schedules,
		"select * from project__schedule where project_id=? and template_id=?",
		projectID,
		templateID)
	return
}
