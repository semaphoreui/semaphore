package projects

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func GetProjects(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	query, args, _ := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("pu.user_id=?", user.ID).
		OrderBy("p.name").
		ToSql()

	var projects []db.Project
	if _, err := db.Mysql.Select(&projects, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, projects)
}

func AddProject(w http.ResponseWriter, r *http.Request) {
	var body db.Project
	user := context.Get(r, "user").(*db.User)

	if err := mulekick.Bind(w, r, &body); err != nil {
		return
	}

	err := body.CreateProject()
	if err != nil {
		panic(err)
	}

	if _, err := db.Mysql.Exec("insert into project__user set project_id=?, user_id=?, admin=1", body.ID, user.ID); err != nil {
		panic(err)
	}

	desc := "Project Created"
	if err := (db.Event{
		ProjectID:   &body.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, body)
}
