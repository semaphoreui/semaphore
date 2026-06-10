package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetNotifications(projectID int, params db.RetrieveQueryParams) (notifications []db.Notification, err error) {
	q := squirrel.Select("*").
		From("project__notification").
		Where("project_id = ?", projectID).
		OrderBy("name")

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&notifications, query, args...)
	return
}

func (d *SqlDb) GetNotification(projectID int, notificationID int) (notification db.Notification, err error) {
	q := squirrel.Select("*").
		From("project__notification").
		Where("project_id = ? AND id = ?", projectID, notificationID)

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	err = d.selectOne(&notification, query, args...)
	return
}

func (d *SqlDb) CreateNotification(notification db.Notification) (db.Notification, error) {
	insertID, err := d.insert(
		"id",
		"insert into project__notification (project_id, name, type, config) values (?, ?, ?, ?)",
		notification.ProjectID,
		notification.Name,
		notification.Type,
		notification.Config,
	)

	if err != nil {
		return db.Notification{}, err
	}

	notification.ID = insertID
	return notification, nil
}

func (d *SqlDb) UpdateNotification(notification db.Notification) error {
	_, err := d.exec(
		"update project__notification set name=?, type=?, config=? where project_id=? and id=?",
		notification.Name,
		notification.Type,
		notification.Config,
		notification.ProjectID,
		notification.ID,
	)
	return err
}

func (d *SqlDb) DeleteNotification(projectID int, notificationID int) (err error) {
	refs, err := d.GetNotificationRefs(projectID, notificationID)
	if err != nil {
		return err
	}
	if len(refs.Templates) > 0 {
		return db.ErrInvalidOperation
	}

	_, err = d.exec("DELETE FROM project__notification WHERE project_id = ? AND id = ?", projectID, notificationID)
	return
}

func (d *SqlDb) GetNotificationRefs(projectID int, notificationID int) (refs db.ObjectReferrers, err error) {
	var templates []db.Template
	templates, err = d.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	// Initialize slices to avoid null values in JSON responses
	refs.Templates = make([]db.ObjectReferrer, 0)
	refs.Inventories = make([]db.ObjectReferrer, 0)
	refs.Repositories = make([]db.ObjectReferrer, 0)
	refs.Integrations = make([]db.ObjectReferrer, 0)
	refs.Schedules = make([]db.ObjectReferrer, 0)
	refs.AccessKeys = make([]db.ObjectReferrer, 0)

	for _, tpl := range templates {
		for _, id := range tpl.NotificationIDs {
			if id == notificationID {
				refs.Templates = append(refs.Templates, db.ObjectReferrer{
					ID:   tpl.ID,
					Name: tpl.Name,
				})
				break
			}
		}
	}

	return
}
