package sql

import (
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
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
	return d.selectAlertIDs(
		"select id as alert_id from project__alert where project_id=? and is_default=? and enabled=? order by id",
		projectID,
		true,
		true,
	)
}

func (d *SqlDb) CreateAlert(alert db.Alert) (newAlert db.Alert, err error) {
	alert.Normalize()
	if err = alert.Validate(); err != nil {
		return
	}
	if err = d.validateAlertNameIsFree(alert.ProjectID, 0, alert.Name); err != nil {
		return
	}
	if err = d.validateAlertKey(alert.ProjectID, alert.KeyID); err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__alert "+
			"(project_id, name, `type`, enabled, is_default, events, chat_id, thread_id, url, recipients, key_id, params, body) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		alert.ProjectID,
		alert.Name,
		alert.Type,
		alert.Enabled,
		alert.IsDefault,
		alert.Events,
		alert.ChatID,
		alert.ThreadID,
		alert.URL,
		alert.Recipients,
		alert.KeyID,
		alert.Params,
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
	if err := alert.Validate(); err != nil {
		return err
	}
	if _, err := d.GetAlert(alert.ProjectID, alert.ID); err != nil {
		return err
	}
	if err := d.validateAlertNameIsFree(alert.ProjectID, alert.ID, alert.Name); err != nil {
		return err
	}
	if err := d.validateAlertKey(alert.ProjectID, alert.KeyID); err != nil {
		return err
	}

	_, err := d.exec(
		"update project__alert set name=?, `type`=?, enabled=?, is_default=?, events=?, chat_id=?, thread_id=?, "+
			"url=?, recipients=?, key_id=?, params=?, body=? where project_id=? and id=?",
		alert.Name,
		alert.Type,
		alert.Enabled,
		alert.IsDefault,
		alert.Events,
		alert.ChatID,
		alert.ThreadID,
		alert.URL,
		alert.Recipients,
		alert.KeyID,
		alert.Params,
		alert.Body,
		alert.ProjectID,
		alert.ID,
	)
	return err
}

// validateAlertKey makes sure the secret key belongs to the same project, so
// a request body can not point an alert at another project's credentials.
func (d *SqlDb) validateAlertKey(projectID int, keyID *int) error {
	if keyID == nil {
		return nil
	}
	if _, err := d.GetAccessKey(projectID, *keyID); err != nil {
		return common_errors.NewValidationError("access key does not belong to this project")
	}
	return nil
}

func (d *SqlDb) SetAlertActive(projectID int, alertID int, active bool) error {
	res, err := d.exec(
		"update project__alert set enabled=? where project_id=? and id=?",
		active,
		projectID,
		alertID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (d *SqlDb) DeleteAlert(projectID int, alertID int) error {
	refs, err := d.GetAlertRefs(projectID, alertID)
	if err != nil {
		return err
	}
	if len(refs.Templates) > 0 || len(refs.Schedules) > 0 {
		return common_errors.NewValidationError("alert is used by templates or schedules")
	}
	return d.deleteObject(projectID, db.AlertProps, alertID)
}

// GetAlertRefs lists templates and schedules explicitly linked to the alert.
// Templates in default mode are not listed: they follow whatever is marked as
// default at send time and do not block deletion.
func (d *SqlDb) GetAlertRefs(projectID int, alertID int) (refs db.ObjectReferrers, err error) {
	refs.Templates = make([]db.ObjectReferrer, 0)
	refs.Schedules = make([]db.ObjectReferrer, 0)
	refs.Inventories = make([]db.ObjectReferrer, 0)
	refs.Repositories = make([]db.ObjectReferrer, 0)
	refs.Integrations = make([]db.ObjectReferrer, 0)
	refs.AccessKeys = make([]db.ObjectReferrer, 0)
	refs.Alerts = make([]db.ObjectReferrer, 0)

	_, err = d.selectAll(
		&refs.Templates,
		"select t.id, t.name from project__template t "+
			"join project__template_alert ta on ta.template_id = t.id "+
			"where ta.project_id=? and ta.alert_id=? order by t.name",
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
			"where sa.project_id=? and sa.alert_id=? order by s.name",
		projectID,
		alertID,
	)
	return
}

func (d *SqlDb) GetTemplateAlerts(projectID int, templateID int) ([]int, error) {
	return d.selectAlertIDs(
		"select alert_id from project__template_alert where project_id=? and template_id=? order by alert_id",
		projectID,
		templateID,
	)
}

func (d *SqlDb) UpdateTemplateAlerts(projectID int, templateID int, alertIDs []int) error {
	return d.replaceAlertLinks(
		"delete from project__template_alert where project_id=? and template_id=?",
		"insert into project__template_alert (project_id, template_id, alert_id) values (?, ?, ?)",
		projectID,
		templateID,
		alertIDs,
	)
}

func (d *SqlDb) GetScheduleAlerts(projectID int, scheduleID int) ([]int, error) {
	return d.selectAlertIDs(
		"select alert_id from project__schedule_alert where project_id=? and schedule_id=? order by alert_id",
		projectID,
		scheduleID,
	)
}

func (d *SqlDb) UpdateScheduleAlerts(projectID int, scheduleID int, alertIDs []int) error {
	return d.replaceAlertLinks(
		"delete from project__schedule_alert where project_id=? and schedule_id=?",
		"insert into project__schedule_alert (project_id, schedule_id, alert_id) values (?, ?, ?)",
		projectID,
		scheduleID,
		alertIDs,
	)
}

func (d *SqlDb) ClaimAlertSend(taskID int, destination string, event db.AlertEvent) (bool, error) {
	_, err := d.exec(
		"insert into task__alert_send (task_id, destination, event, created) values (?, ?, ?, ?)",
		taskID,
		destination,
		string(event),
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

func (d *SqlDb) UnclaimAlertSend(taskID int, destination string, event db.AlertEvent) error {
	_, err := d.exec(
		"delete from task__alert_send where task_id=? and destination=? and event=?",
		taskID,
		destination,
		string(event),
	)
	return err
}

// isUniqueViolation matches the primary-key conflict raised by SQLite
// ("UNIQUE constraint failed"), MySQL ("Duplicate entry") and PostgreSQL
// ("duplicate key value violates unique constraint").
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

func (d *SqlDb) selectAlertIDs(query string, args ...any) (alertIDs []int, err error) {
	alertIDs = make([]int, 0)

	var rows []struct {
		AlertID int `db:"alert_id"`
	}
	_, err = d.selectAll(&rows, query, args...)
	if err != nil {
		return
	}
	for _, r := range rows {
		alertIDs = append(alertIDs, r.AlertID)
	}
	return
}

func (d *SqlDb) replaceAlertLinks(deleteQuery, insertQuery string, projectID int, ownerID int, alertIDs []int) error {
	if err := d.validateAlertIDs(projectID, alertIDs); err != nil {
		return err
	}

	if _, err := d.exec(deleteQuery, projectID, ownerID); err != nil {
		return err
	}

	seen := make(map[int]bool)
	for _, alertID := range alertIDs {
		if seen[alertID] {
			continue
		}
		seen[alertID] = true
		if _, err := d.exec(insertQuery, projectID, ownerID, alertID); err != nil {
			return err
		}
	}
	return nil
}

// validateAlertIDs makes sure every linked alert belongs to the project, so a
// request body cannot bind a template to another project's destination.
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
