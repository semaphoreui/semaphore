package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateSchedule(schedule db.Schedule) (newSchedule db.Schedule, err error) {

	if schedule.TaskParamsID != nil {
		params := schedule.TaskParams
		params.ProjectID = schedule.ProjectID
		err = d.Sql().Insert(&params)
		if err != nil {
			return
		}
		schedule.TaskParamsID = &params.ID
	}

	insertID, err := d.insert(
		"id",
		"insert into project__schedule (project_id, template_id, cron_format, repository_id, `name`, `active`, task_params_id)"+
			"values (?, ?, ?, ?, ?, ?, ?)",
		schedule.ProjectID,
		schedule.TemplateID,
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.Name,
		schedule.Active,
		schedule.TaskParamsID)

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

func (d *SqlDb) UpdateSchedule(schedule db.Schedule) (err error) {

	if schedule.TaskParamsID != nil {
		var curr db.Schedule
		err = d.getObject(schedule.ProjectID, db.ScheduleProps, schedule.ID, &curr)
		if err != nil {
			return
		}

		params := schedule.TaskParams
		params.ProjectID = schedule.ProjectID

		if curr.TaskParamsID == nil {
			err = d.Sql().Insert(&params)
		} else {
			params.ID = *curr.TaskParamsID
			_, err = d.Sql().Update(&params)
		}

		if err != nil {
			return
		}

		schedule.TaskParamsID = &params.ID
	}

	_, err = d.exec("update project__schedule set "+
		"cron_format=?, "+
		"repository_id=?, "+
		"template_id=?, "+
		"`name`=?, "+
		"`active`=?, "+
		"last_commit_hash = NULL, "+
		"task_params_id=? "+
		"where project_id=? and id=?",
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.TemplateID,
		schedule.Name,
		schedule.Active,
		schedule.TaskParamsID,
		schedule.ProjectID,
		schedule.ID)

	return
}

func (d *SqlDb) GetSchedule(projectID int, scheduleID int) (schedule db.Schedule, err error) {
	err = d.selectOne(
		&schedule,
		"select * from project__schedule where project_id=? and id=?",
		projectID,
		scheduleID)

	if err != nil {
		return
	}

	if schedule.TaskParamsID != nil {
		var taskParams db.TaskParams
		err = d.getObject(projectID, db.TaskParamsProps, *schedule.TaskParamsID, &taskParams)
		if err != nil {
			return
		}

		schedule.TaskParams = taskParams
	}

	return
}

func (d *SqlDb) DeleteSchedule(projectID int, scheduleID int) (err error) {
	var schedule db.Schedule
	err = d.getObject(projectID, db.ScheduleProps, scheduleID, &schedule)
	if err != nil {
		return
	}

	err = d.deleteObject(projectID, db.ScheduleProps, scheduleID)
	if err != nil {
		return
	}

	if schedule.TaskParamsID != nil {
		err = d.deleteObject(projectID, db.TaskParamsProps, *schedule.TaskParamsID)
	}

	return err
}

func (d *SqlDb) GetSchedules() (schedules []db.Schedule, err error) {
	_, err = d.selectAll(&schedules, "select * from project__schedule where cron_format != ''")
	return
}

func (d *SqlDb) GetProjectSchedules(projectID int, includeTaskParams bool, includeCommitCheckers bool) (schedules []db.ScheduleWithTpl, err error) {

	repoFilter := ""
	if !includeCommitCheckers {
		repoFilter = "ps.repository_id IS NULL AND "
	}

	_, err = d.selectAll(&schedules,
		"SELECT ps.*, pt.name as tpl_name FROM project__schedule ps "+
			"JOIN project__template pt ON pt.id = ps.template_id "+
			"WHERE "+
			repoFilter+
			"ps.project_id=?",
		projectID)

	if includeTaskParams {
		for i := range schedules {
			if schedules[i].TaskParamsID == nil {
				continue
			}

			var taskParams db.TaskParams
			err = d.getObject(projectID, db.TaskParamsProps, *schedules[i].TaskParamsID, &taskParams)
			if err != nil {
				return nil, err
			}
			schedules[i].TaskParams = taskParams
		}
	}

	return
}

func (d *SqlDb) GetTemplateSchedules(projectID int, templateID int, onlyCommitCheckers bool) (schedules []db.Schedule, err error) {

	q := squirrel.Select("*").
		From("project__schedule").
		Where("project_id=?", projectID).
		Where("template_id=?", templateID)

	if onlyCommitCheckers {
		q = q.Where("repository_id IS NOT NULL")
	}

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&schedules, query, args...)
	return
}

func (d *SqlDb) SetScheduleActive(projectID int, scheduleID int, active bool) error {
	_, err := d.exec("update project__schedule set `active`=? where project_id=? and id=?",
		active,
		projectID,
		scheduleID)
	return err
}

func (d *SqlDb) SetScheduleCommitHash(projectID int, scheduleID int, hash string) error {
	_, err := d.exec("update project__schedule set last_commit_hash=? where project_id=? and id=?",
		hash,
		projectID,
		scheduleID)
	return err
}
