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
		OrderBy("id desc").
		Where(`(
			p.id IS NULL
			or exists (
				select 1 from project__user_role pur join `+"`role`"+` r on r.id = pur.role_id
				where pur.project_id = p.id and pur.user_id = ? and r.permissions & 128 <> 0
			)
			or exists (
				select 1 from project__user_mapped_role mr
				join group_role_mapping m on m.id = mr.mapping_id
				join `+"`role`"+` r on r.id = m.role_id
				where mr.project_id = p.id and mr.user_id = ? and r.permissions & 128 <> 0
			)
		)`, userID, userID)

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
