package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) getEvents(q squirrel.SelectBuilder, params db.RetrieveQueryParams) (events []db.Event, err error) {

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&events, query, args...)

	if err != nil {
		return
	}

	err = db.FillEvents(d, events)

	return
}

func (d *SqlDb) CreateEvent(evt db.Event) (newEvent db.Event, err error) {
	var created = tz.Now()

	_, err = d.exec(
		"insert into event(user_id, project_id, integration_id, object_id, object_type, description, created, action, ip, user_agent) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		evt.UserID,
		evt.ProjectID,
		evt.IntegrationID,
		evt.ObjectID,
		evt.ObjectType,
		evt.Description,
		created,
		evt.Action,
		evt.IP,
		evt.UserAgent)

	if err != nil {
		return
	}

	newEvent = evt
	newEvent.Created = created
	return
}

func (d *SqlDb) GetUserEvents(userID int, params db.RetrieveQueryParams) ([]db.Event, error) {
	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("id desc").
		LeftJoin("project__user as pu on pu.project_id=p.id").
		Where("event.project_id IS NOT NULL AND pu.user_id=?", userID)

	return d.getEvents(q, params)
}

func (d *SqlDb) GetEvents(projectID int, params db.RetrieveQueryParams) ([]db.Event, error) {
	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("id desc").
		Where("event.project_id=?", projectID)

	return d.getEvents(q, params)
}

func (d *SqlDb) GetAllEvents(params db.RetrieveQueryParams) ([]db.Event, error) {
	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("id desc")

	return d.getEvents(q, params)
}
