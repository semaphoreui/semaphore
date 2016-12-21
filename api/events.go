package api

import (
	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func getEvents(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	q := squirrel.Select("event.*, p.name as project_name").
		From("event").
		LeftJoin("project as p on event.project_id=p.id").
		OrderBy("created desc")

	projectObj, exists := c.Get("project")
	if exists == true {
		// limit query to project
		project := projectObj.(models.Project)
		q = q.Where("event.project_id=?", project.ID)
	} else {
		q = q.LeftJoin("project__user as pu on pu.project_id=p.id").
			Where("p.id IS NULL or pu.user_id=?", user.ID)
	}

	var events []models.Event

	query, args, _ := q.ToSql()
	if _, err := database.Mysql.Select(&events, query, args...); err != nil {
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

		query, args, _ := q.ToSql()
		name, err := database.Mysql.SelectNullStr(query, args...)
		if err != nil {
			panic(err)
		}

		if name.Valid == true {
			events[i].ObjectName = name.String
		}
	}

	c.JSON(200, events)
}
