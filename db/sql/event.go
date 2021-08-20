package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
	"time"
)

func (d *SqlDb) getEventObjectName(evt db.Event) (string, error) {
	if evt.ObjectID == nil || evt.ObjectType == nil {
		return "", nil
	}

	var q squirrel.SelectBuilder

	switch *evt.ObjectType {
	case "task":
		q = squirrel.Select("case when length(task.playbook) > 0 then task.playbook else tpl.playbook end").
			From("task").
			Join("project__template as tpl on task.template_id=tpl.id").
			Where("task.id=?", evt.ObjectID)
	default:
		return "", nil
	}

	query, args, err := q.ToSql()

	if err != nil {
		return "", err
	}

	var name sql.NullString
	name, err = d.sql.SelectNullStr(query, args...)

	if err != nil {
		return "", err
	}

	if name.Valid {
		return name.String, nil
	} else {
		return "", nil
	}
}

func (d *SqlDb) getEvents(q squirrel.SelectBuilder, params db.RetrieveQueryParams) (events []db.Event, err error) {

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(&events, query, args...)

	if err != nil {
		return
	}

	for i, evt := range events {
		var objName string
		objName, err = d.getEventObjectName(evt)
		if objName == "" {
			continue
		}

		if err != nil {
			return
		}

		events[i].ObjectName = objName
	}

	return
}

func (d *SqlDb) CreateEvent(evt db.Event) (newEvent db.Event, err error) {
	var created = time.Now()

	_, err = d.sql.Exec(
		"insert into event(user_id, project_id, object_id, object_type, description, created) values (?, ?, ?, ?, ?, ?)",
		evt.UserID,
		evt.ProjectID,
		evt.ObjectID,
		evt.ObjectType,
		evt.Description,
		created)

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
		OrderBy("created desc").
		LeftJoin("project__user as pu on pu.project_id=p.id").
		Where("p.id IS NULL or pu.user_id=?", userID)

	return d.getEvents(q, params)
}

func (d *SqlDb) GetEvents(projectID int, params db.RetrieveQueryParams) ([]db.Event, error) {
	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("created desc").
		Where("event.project_id=?", projectID)

	return d.getEvents(q, params)
}
