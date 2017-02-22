package projects

import (
	"database/sql"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func ProjectMiddleware(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*models.User)

	projectID, err := util.GetIntParam("project_id", c)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("p.id=?", projectID).
		Where("pu.user_id=?", user.ID).
		ToSql()

	var project models.Project
	if err := database.Mysql.SelectOne(&project, query, args...); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	c.Set("project", project)
	c.Next()
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	c.JSON(200, context.Get(r, "project"))
}

func MustBeAdmin(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	user := context.Get(r, "user").(*models.User)

	userC, err := database.Mysql.SelectInt("select count(1) from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=? and pu.admin=1", user.ID, project.ID)
	if err != nil {
		panic(err)
	}

	if userC == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var body struct {
		Name string `json:"name"`
	}

	if err := mulekick.Bind(w, r, &body); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update project set name=? where id=?", body.Name, project.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)

	tx, err := database.Mysql.Begin()
	if err != nil {
		panic(err)
	}

	statements := []string{
		"delete tao from task__output as tao join task as t on t.id=tao.task_id join project__template as pt on pt.id=t.template_id where pt.project_id=?",
		"delete t from task as t join project__template as pt on pt.id=t.template_id where pt.project_id=?",
		"delete from project__template where project_id=?",
		"delete from project__user where project_id=?",
		"delete from project__repository where project_id=?",
		"delete from project__inventory where project_id=?",
		"delete from access_key where project_id=?",
		"delete from project where id=?",
	}

	for _, statement := range statements {
		_, err := tx.Exec(statement, project.ID)

		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
