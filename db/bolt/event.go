package bolt

import (
	"database/sql"
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
	"go.etcd.io/bbolt"
	"time"
)

func (d *BoltDb) getEventObjectName(evt db.Event) (string, error) {
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

func (d *BoltDb) getEvents(q squirrel.SelectBuilder, params db.RetrieveQueryParams) (events []db.Event, err error) {

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

func (d *BoltDb) CreateEvent(evt db.Event) (newEvent db.Event, err error) {
	newEvent = evt
	newEvent.Created = time.Now()

	err = d.db.Update(func(tx *bbolt.Tx) error {
		b, err2 := tx.CreateBucketIfNotExists([]byte("events"))
		if err2 != nil {
			return err2
		}

		str, err2 := json.Marshal(newEvent)
		if err2 != nil {
			return err2
		}

		id, err2 := b.NextSequence()
		if err2 != nil {
			return err2
		}

		return b.Put(makeObjectId(int(id)), str)
	})

	return
}

func (d *BoltDb) GetUserEvents(userID int, params db.RetrieveQueryParams) ([]db.Event, error) {
	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("created desc").
		LeftJoin("project__user as pu on pu.project_id=p.id").
		Where("p.id IS NULL or pu.user_id=?", userID)

	return d.getEvents(q, params)
}

func (d *BoltDb) GetEvents(projectID int, params db.RetrieveQueryParams) ([]db.Event, error) {
	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("created desc").
		Where("event.project_id=?", projectID)

	return d.getEvents(q, params)
}
