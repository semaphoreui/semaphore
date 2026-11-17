package sql

import (
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) validateScheduleTarget(schedule db.Schedule) error {
	return d.validateRunTarget(schedule.ProjectID, schedule.TemplateID, schedule.WorkflowID, "Schedule")
}

func (d *SqlDb) CreateSchedule(schedule db.Schedule) (newSchedule db.Schedule, err error) {
	if err = schedule.Validate(); err != nil {
		return
	}

	if err = d.validateScheduleTarget(schedule); err != nil {
		return
	}

	if schedule.RunsWorkflow() {
		// A workflow's root node carries its own task params.
		schedule.TaskParams = nil
	}

	if schedule.TaskParams != nil {
		params := schedule.TaskParams
		params.ProjectID = schedule.ProjectID
		err = d.Sql().Insert(params)
		if err != nil {
			return
		}
		schedule.TaskParamsID = &params.ID
	}

	if schedule.Type == "" {
		schedule.Type = db.ScheduleTypeCron
	}

	insertID, err := d.insert(
		"id",
		"insert into project__schedule (project_id, template_id, workflow_id, cron_format, repository_id, `name`, `active`, run_at, `type`, task_params_id, delete_after_run)"+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		schedule.ProjectID,
		schedule.TemplateID,
		schedule.WorkflowID,
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.Name,
		schedule.Active,
		schedule.RunAt,
		schedule.Type,
		schedule.TaskParamsID,
		schedule.DeleteAfterRun)

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
	if err = schedule.Validate(); err != nil {
		return
	}

	if err = d.validateScheduleTarget(schedule); err != nil {
		return
	}

	var curr db.Schedule
	err = d.getObject(schedule.ProjectID, db.ScheduleProps, schedule.ID, &curr)
	if err != nil {
		return
	}

	if schedule.RunsWorkflow() {
		// Switching to a workflow leaves the old task params behind: the row
		// is unlinked below and removed after the update succeeds.
		schedule.TaskParams = nil
		schedule.TaskParamsID = nil
	} else if schedule.TaskParams == nil {
		schedule.TaskParamsID = curr.TaskParamsID
	}

	if schedule.TaskParams != nil {
		params := schedule.TaskParams
		params.ProjectID = schedule.ProjectID

		if curr.TaskParamsID == nil {
			err = d.Sql().Insert(params)
		} else {
			params.ID = *curr.TaskParamsID
			_, err = d.Sql().Update(params)
		}

		if err != nil {
			return
		}

		schedule.TaskParamsID = &params.ID
	}

	if schedule.Type == "" {
		schedule.Type = db.ScheduleTypeCron
	}

	_, err = d.exec("update project__schedule set "+
		"cron_format=?, "+
		"repository_id=?, "+
		"template_id=?, "+
		"workflow_id=?, "+
		"`name`=?, "+
		"`active`=?, "+
		"run_at=?, "+
		"`type`=?, "+
		"last_commit_hash = NULL, "+
		"task_params_id=?, "+
		"delete_after_run=? "+
		"where project_id=? and id=?",
		schedule.CronFormat,
		schedule.RepositoryID,
		schedule.TemplateID,
		schedule.WorkflowID,
		schedule.Name,
		schedule.Active,
		schedule.RunAt,
		schedule.Type,
		schedule.TaskParamsID,
		schedule.DeleteAfterRun,
		schedule.ProjectID,
		schedule.ID)

	if err != nil {
		return
	}

	if schedule.RunsWorkflow() && curr.TaskParamsID != nil {
		err = d.deleteObject(schedule.ProjectID, db.TaskParamsProps, *curr.TaskParamsID)
		if errors.Is(err, db.ErrNotFound) {
			err = nil
		}
	}

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

		schedule.TaskParams = &taskParams
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
		if errors.Is(err, db.ErrNotFound) {
			err = nil
		}
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

	// A schedule runs a template or a workflow, so one of the two joins
	// misses; the missing name comes back empty rather than null.
	_, err = d.selectAll(&schedules,
		"SELECT ps.*, COALESCE(pt.name, '') as tpl_name, COALESCE(pw.name, '') as wf_name FROM project__schedule ps "+
			"LEFT JOIN project__template pt ON pt.id = ps.template_id "+
			"LEFT JOIN project__workflow_template pw ON pw.id = ps.workflow_id "+
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
			schedules[i].TaskParams = &taskParams
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
