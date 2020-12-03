package api

import (
	util2 "github.com/ansible-semaphore/semaphore/api/util"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

//nolint: gocyclo
func getEvents(w http.ResponseWriter, r *http.Request, limit uint64) {
	user := context.Get(r, "user").(*models.User)

	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("created desc")

	if limit > 0 {
		q = q.Limit(limit)
	}

	projectObj, exists := context.GetOk(r, "project")
	if exists {
		// limit query to project
		project := projectObj.(models.Project)
		q = q.Where("event.project_id=?", project.ID)
	} else {
		q = q.LeftJoin("project__user as pu on pu.project_id=p.id").
			Where("p.id IS NULL or pu.user_id=?", user.ID)
	}

	var events []models.Event

	query, args, err := q.ToSql()
	util.LogWarning(err)
	if _, err := util2.GetStore(r).Sql().Select(&events, query, args...); err != nil {
		panic(err)
	}

	for i, evt := range events {
		if evt.ObjectID == nil || evt.ObjectType == nil {
			continue
		}

		var q squirrel.SelectBuilder

		switch *evt.ObjectType {
		case "task":
			q = squirrel.Select("case when length(task.playbook) > 0 then task.playbook else tpl.playbook end").
				From("task").
				Join("project__template as tpl on task.template_id=tpl.id").
				Where("task.id=?", evt.ObjectID)
		default:
			continue
		}

		query, args, err := q.ToSql()
		util.LogWarning(err)
		name, err := util2.GetStore(r).Sql().SelectNullStr(query, args...)
		if err != nil {
			panic(err)
		}

		if name.Valid {
			events[i].ObjectName = name.String
		}
	}

	util2.WriteJSON(w, http.StatusOK, events)
}

func getLastEvents(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, 200)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, 0)
}
