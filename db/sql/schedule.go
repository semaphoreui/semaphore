package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateSchedule(schedule db.Schedule) (newSchedule db.Schedule, err error) {
	if schedule.Type == "" {
		schedule.Type = db.ScheduleTypeCron
	}
	if err = schedule.NormalizeAlerts(); err != nil {
		return
	}
	if err = d.validateAlertIDs(schedule.ProjectID, schedule.AlertIDs); err != nil {
		return
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	if schedule.TaskParams != nil {
		params := schedule.TaskParams
		params.ProjectID = schedule.ProjectID
		err = tx.Insert(params)
		if err != nil {
			_ = tx.Rollback()
			return
		}
		schedule.TaskParamsID = &params.ID
	}

	insertID, err := d.insertTx(
		tx,
		"id",
		"insert into project__schedule (project_id, template_id, cron_format, repository_id, `name`, `active`, run_at, `type`, task_params_id, delete_after_run, alert_mode, alert_on_success, alert_on_error)"+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		schedule.ProjectID,
		schedule.TemplateID,
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.Name,
		schedule.Active,
		schedule.RunAt,
		schedule.Type,
		schedule.TaskParamsID,
		schedule.DeleteAfterRun,
		schedule.AlertMode,
		schedule.AlertOnSuccess,
		schedule.AlertOnError)

	if err != nil {
		_ = tx.Rollback()
		return
	}

	if err = d.updateScheduleAlertsInTx(tx, schedule.ProjectID, insertID, schedule.AlertIDs); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	newSchedule = schedule
	newSchedule.ID = insertID
	err = d.fillScheduleAlerts(newSchedule.ProjectID, &newSchedule)
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
	if schedule.Type == "" {
		schedule.Type = db.ScheduleTypeCron
	}
	if schedule.AlertMode == "" {
		var curr db.Schedule
		if err = d.getObject(schedule.ProjectID, db.ScheduleProps, schedule.ID, &curr); err != nil {
			return
		}
		schedule.AlertMode = curr.AlertMode
		if schedule.AlertOnSuccess == nil {
			schedule.AlertOnSuccess = curr.AlertOnSuccess
		}
		if schedule.AlertOnError == nil {
			schedule.AlertOnError = curr.AlertOnError
		}
	}
	if err = schedule.NormalizeAlerts(); err != nil {
		return
	}
	if schedule.AlertIDs != nil {
		if err = d.validateAlertIDs(schedule.ProjectID, schedule.AlertIDs); err != nil {
			return
		}
	}

	var curr db.Schedule
	if schedule.TaskParams != nil {
		err = d.getObject(schedule.ProjectID, db.ScheduleProps, schedule.ID, &curr)
		if err != nil {
			return
		}
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	if schedule.TaskParams != nil {
		params := schedule.TaskParams
		params.ProjectID = schedule.ProjectID

		if curr.TaskParamsID == nil {
			err = tx.Insert(params)
		} else {
			params.ID = *curr.TaskParamsID
			_, err = tx.Update(params)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		schedule.TaskParamsID = &params.ID
	}

	_, err = d.execTx(tx, "update project__schedule set "+
		"cron_format=?, "+
		"repository_id=?, "+
		"template_id=?, "+
		"`name`=?, "+
		"`active`=?, "+
		"run_at=?, "+
		"`type`=?, "+
		"last_commit_hash = NULL, "+
		"task_params_id=?, "+
		"delete_after_run=?, "+
		"alert_mode=?, "+
		"alert_on_success=?, "+
		"alert_on_error=? "+
		"where project_id=? and id=?",
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.TemplateID,
		schedule.Name,
		schedule.Active,
		schedule.RunAt,
		schedule.Type,
		schedule.TaskParamsID,
		schedule.DeleteAfterRun,
		schedule.AlertMode,
		schedule.AlertOnSuccess,
		schedule.AlertOnError,
		schedule.ProjectID,
		schedule.ID)
	if err != nil {
		_ = tx.Rollback()
		return
	}

	// inherit never keeps custom bindings, so delete refs stay accurate.
	// nil alert_ids on ids mode means the client omitted them (legacy PUT).
	if schedule.AlertMode != db.AlertModeIDs {
		if err = d.updateScheduleAlertsInTx(tx, schedule.ProjectID, schedule.ID, nil); err != nil {
			_ = tx.Rollback()
			return
		}
	} else if schedule.AlertIDs != nil {
		if err = d.updateScheduleAlertsInTx(tx, schedule.ProjectID, schedule.ID, schedule.AlertIDs); err != nil {
			_ = tx.Rollback()
			return
		}
	}

	return tx.Commit()
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

		schedule.TaskParams = &taskParams
	}

	err = d.fillScheduleAlerts(projectID, &schedule)
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
	_, err = d.selectAll(&schedules, "select * from project__schedule where cron_format != '' or run_at is not null")
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
	if err != nil {
		return
	}

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
			schedules[i].TaskParams = &taskParams
		}
	}

	for i := range schedules {
		err = d.fillScheduleAlerts(projectID, &schedules[i].Schedule)
		if err != nil {
			return nil, err
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
	if err != nil {
		return
	}

	for i := range schedules {
		err = d.fillScheduleAlerts(projectID, &schedules[i])
		if err != nil {
			return
		}
	}
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
