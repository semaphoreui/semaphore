package sql

import (
	"strings"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) GetAlert(projectID int, alertID int) (alert db.Alert, err error) {
	err = d.getObject(projectID, db.AlertProps, alertID, &alert)
	return
}

func (d *SqlDb) GetAlerts(projectID int, params db.RetrieveQueryParams) (alerts []db.Alert, err error) {
	alerts = make([]db.Alert, 0)
	err = d.getObjects(projectID, db.AlertProps, params, nil, &alerts)
	if alerts == nil {
		alerts = make([]db.Alert, 0)
	}
	return
}

func (d *SqlDb) GetDefaultAlertIDs(projectID int) (alertIDs []int, err error) {
	alertIDs = make([]int, 0)

	var rows []struct {
		ID int `db:"id"`
	}
	_, err = d.selectAll(
		&rows,
		"select id from project__alert where project_id=? and is_default=? order by id",
		projectID,
		true,
	)
	if err != nil {
		return
	}
	for _, r := range rows {
		alertIDs = append(alertIDs, r.ID)
	}
	return
}

func (d *SqlDb) CreateAlert(alert db.Alert) (newAlert db.Alert, err error) {
	alert.Normalize()
	if err = alert.Validate(); err != nil {
		return
	}

	if err = d.validateAlertNameIsFree(alert.ProjectID, 0, alert.Name); err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__alert "+
			"(project_id, name, `type`, enabled, is_default, chat_id, thread_id, url, token, recipients, key_id, body) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		alert.ProjectID,
		alert.Name,
		alert.Type,
		alert.Enabled,
		alert.IsDefault,
		alert.ChatID,
		alert.ThreadID,
		alert.URL,
		alert.Token,
		alert.Recipients,
		alert.KeyID,
		alert.Body,
	)
	if err != nil {
		return
	}

	newAlert = alert
	newAlert.ID = insertID
	return
}

func (d *SqlDb) UpdateAlert(alert db.Alert) error {
	alert.Normalize()

	old, err := d.GetAlert(alert.ProjectID, alert.ID)
	if err != nil {
		return err
	}
	if alert.Type == db.AlertTypeGotify && alert.Token == nil {
		alert.Token = old.Token
	}

	if err := alert.Validate(); err != nil {
		return err
	}

	if err := d.validateAlertNameIsFree(alert.ProjectID, alert.ID, alert.Name); err != nil {
		return err
	}

	_, err = d.exec(
		"update project__alert set name=?, `type`=?, enabled=?, is_default=?, chat_id=?, thread_id=?, "+
			"url=?, token=?, recipients=?, key_id=?, body=? where project_id=? and id=?",
		alert.Name,
		alert.Type,
		alert.Enabled,
		alert.IsDefault,
		alert.ChatID,
		alert.ThreadID,
		alert.URL,
		alert.Token,
		alert.Recipients,
		alert.KeyID,
		alert.Body,
		alert.ProjectID,
		alert.ID,
	)
	return err
}

func (d *SqlDb) DeleteAlert(projectID int, alertID int) error {
	refs, err := d.GetAlertRefs(projectID, alertID)
	if err != nil {
		return err
	}
	if len(refs.Templates) > 0 || len(refs.Schedules) > 0 {
		return common_errors.NewValidationError("alert is used by templates or schedules")
	}
	used, err := d.alertReferencedByActiveTasks(projectID, alertID)
	if err != nil {
		return err
	}
	if used {
		return common_errors.NewValidationError("alert is used by running tasks")
	}
	return d.deleteObject(projectID, db.AlertProps, alertID)
}

func (d *SqlDb) GetAlertRefs(projectID int, alertID int) (refs db.ObjectReferrers, err error) {
	refs.Templates = make([]db.ObjectReferrer, 0)
	refs.Schedules = make([]db.ObjectReferrer, 0)
	refs.Inventories = make([]db.ObjectReferrer, 0)
	refs.Repositories = make([]db.ObjectReferrer, 0)
	refs.Integrations = make([]db.ObjectReferrer, 0)
	refs.AccessKeys = make([]db.ObjectReferrer, 0)

	_, err = d.selectAll(
		&refs.Templates,
		"select t.id, t.name from project__template t "+
			"join project__template_alert ta on ta.template_id = t.id "+
			"where ta.project_id=? and ta.alert_id=?",
		projectID,
		alertID,
	)
	if err != nil {
		return
	}

	_, err = d.selectAll(
		&refs.Schedules,
		"select s.id, s.name from project__schedule s "+
			"join project__schedule_alert sa on sa.schedule_id = s.id "+
			"where sa.project_id=? and sa.alert_id=?",
		projectID,
		alertID,
	)
	if err != nil {
		return
	}

	alert, alertErr := d.GetAlert(projectID, alertID)
	if alertErr != nil {
		return refs, alertErr
	}
	if !alert.IsDefault {
		return
	}

	var defaults []db.ObjectReferrer
	_, err = d.selectAll(
		&defaults,
		"select id, name from project__template where project_id=? and alert_mode=?",
		projectID,
		db.AlertModeDefault,
	)
	if err != nil {
		return
	}

	seen := make(map[int]bool)
	for _, ref := range refs.Templates {
		seen[ref.ID] = true
	}
	for _, ref := range defaults {
		if seen[ref.ID] {
			continue
		}
		refs.Templates = append(refs.Templates, ref)
	}
	return
}

func (d *SqlDb) alertReferencedByActiveTasks(projectID int, alertID int) (bool, error) {
	unfinished := task_logger.UnfinishedTaskStatuses()
	placeholders := make([]string, len(unfinished))
	args := make([]any, 0, 1+len(unfinished))
	args = append(args, projectID)
	for i, status := range unfinished {
		placeholders[i] = "?"
		args = append(args, status)
	}

	var tasks []db.Task
	_, err := d.selectAll(
		&tasks,
		"select id, alert_snapshot from task where project_id=? and status in ("+strings.Join(placeholders, ",")+")",
		args...,
	)
	if err != nil {
		return false, err
	}
	for _, task := range tasks {
		if task.AlertSnapshot == nil {
			continue
		}
		for _, id := range task.AlertSnapshot.AlertIDs {
			if id == alertID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (d *SqlDb) GetTemplateAlerts(projectID int, templateID int) (alertIDs []int, err error) {
	return d.getLinkedAlertIDs(
		"select alert_id from project__template_alert where project_id=? and template_id=? order by alert_id",
		projectID,
		templateID,
	)
}

func (d *SqlDb) UpdateTemplateAlerts(projectID int, templateID int, alertIDs []int) error {
	return d.replaceLinkedAlerts(
		"delete from project__template_alert where project_id=? and template_id=?",
		"insert into project__template_alert (project_id, template_id, alert_id) values (?, ?, ?)",
		projectID,
		templateID,
		alertIDs,
	)
}

func (d *SqlDb) GetScheduleAlerts(projectID int, scheduleID int) (alertIDs []int, err error) {
	return d.getLinkedAlertIDs(
		"select alert_id from project__schedule_alert where project_id=? and schedule_id=? order by alert_id",
		projectID,
		scheduleID,
	)
}

func (d *SqlDb) UpdateScheduleAlerts(projectID int, scheduleID int, alertIDs []int) error {
	return d.replaceLinkedAlerts(
		"delete from project__schedule_alert where project_id=? and schedule_id=?",
		"insert into project__schedule_alert (project_id, schedule_id, alert_id) values (?, ?, ?)",
		projectID,
		scheduleID,
		alertIDs,
	)
}

func (d *SqlDb) getLinkedAlertIDs(query string, projectID int, ownerID int) (alertIDs []int, err error) {
	alertIDs = make([]int, 0)

	var rows []struct {
		AlertID int `db:"alert_id"`
	}

	_, err = d.selectAll(&rows, query, projectID, ownerID)
	if err != nil {
		return
	}

	for _, r := range rows {
		alertIDs = append(alertIDs, r.AlertID)
	}
	return
}

func (d *SqlDb) replaceLinkedAlerts(deleteQuery, insertQuery string, projectID int, ownerID int, alertIDs []int) (err error) {
	if err = d.validateAlertIDs(projectID, alertIDs); err != nil {
		return
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	if err = d.replaceLinkedAlertsInTx(tx, deleteQuery, insertQuery, projectID, ownerID, alertIDs); err != nil {
		_ = tx.Rollback()
		return
	}

	return tx.Commit()
}

func (d *SqlDb) updateTemplateAlertsInTx(tx *gorp.Transaction, projectID int, templateID int, alertIDs []int) error {
	return d.replaceLinkedAlertsInTx(
		tx,
		"delete from project__template_alert where project_id=? and template_id=?",
		"insert into project__template_alert (project_id, template_id, alert_id) values (?, ?, ?)",
		projectID,
		templateID,
		alertIDs,
	)
}

func (d *SqlDb) updateScheduleAlertsInTx(tx *gorp.Transaction, projectID int, scheduleID int, alertIDs []int) error {
	return d.replaceLinkedAlertsInTx(
		tx,
		"delete from project__schedule_alert where project_id=? and schedule_id=?",
		"insert into project__schedule_alert (project_id, schedule_id, alert_id) values (?, ?, ?)",
		projectID,
		scheduleID,
		alertIDs,
	)
}

func (d *SqlDb) replaceLinkedAlertsInTx(
	tx *gorp.Transaction,
	deleteQuery string,
	insertQuery string,
	projectID int,
	ownerID int,
	alertIDs []int,
) error {
	_, err := tx.Exec(d.PrepareQuery(deleteQuery), projectID, ownerID)
	if err != nil {
		return err
	}

	seen := make(map[int]bool)
	for _, alertID := range alertIDs {
		if seen[alertID] {
			continue
		}
		seen[alertID] = true

		_, err = tx.Exec(d.PrepareQuery(insertQuery), projectID, ownerID, alertID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *SqlDb) ClaimAlertSend(taskID int, alertID int, event string) (bool, error) {
	_, err := d.exec(
		"insert into task__alert_send (task_id, alert_id, event, created) values (?, ?, ?, ?)",
		taskID,
		alertID,
		event,
		tz.Now(),
	)
	if err == nil {
		return true, nil
	}
	if isUniqueViolation(err) {
		return false, nil
	}
	return false, err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") ||
		strings.Contains(msg, "duplicate")
}

func (d *SqlDb) validateAlertIDs(projectID int, alertIDs []int) error {
	for _, alertID := range alertIDs {
		if _, err := d.GetAlert(projectID, alertID); err != nil {
			return common_errors.NewValidationError("alert does not belong to this project")
		}
	}
	return nil
}

func (d *SqlDb) validateAlertNameIsFree(projectID int, alertID int, name string) error {
	var count int
	err := d.selectOne(
		&count,
		"select count(*) from project__alert where project_id=? and name=? and id<>?",
		projectID,
		name,
		alertID,
	)
	if err != nil {
		return err
	}
	if count > 0 {
		return common_errors.NewValidationError("alert with name " + name + " already exists")
	}
	return nil
}

func (d *SqlDb) fillScheduleAlerts(projectID int, schedule *db.Schedule) error {
	ids, err := d.GetScheduleAlerts(projectID, schedule.ID)
	if err != nil {
		return err
	}
	schedule.AlertIDs = ids
	if schedule.AlertMode == "" {
		schedule.AlertMode = db.AlertModeInherit
	}
	return nil
}
