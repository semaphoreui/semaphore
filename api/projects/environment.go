package projects

import (
	"database/sql"
	"net/http"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func EnvironmentMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	envID, err := util.GetIntParam("environment_id", w, r)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("*").
		From("project__environment").
		Where("project_id=?", project.ID).
		Where("id=?", envID).
		ToSql()

	var env models.Environment
	if err := database.Mysql.SelectOne(&env, query, args...); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "environment", env)
}

func GetEnvironment(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var env []models.Environment

	q := squirrel.Select("*").
		From("project__environment").
		Where("project_id=?", project.ID)

	query, args, _ := q.ToSql()

	if _, err := database.Mysql.Select(&env, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, env)
}

func UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	oldEnv := context.Get(r, "environment").(models.Environment)
	var env models.Environment
	if err := mulekick.Bind(w, r, &env); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update project__environment set name=?, json=? where id=?", env.Name, env.JSON, oldEnv.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddEnvironment(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var env models.Environment

	if err := mulekick.Bind(w, r, &env); err != nil {
		return
	}

	res, err := database.Mysql.Exec("insert into project__environment set project_id=?, name=?, json=?, password=?", project.ID, env.Name, env.JSON, env.Password)
	if err != nil {
		panic(err)
	}

	insertID, _ := res.LastInsertId()
	insertIDInt := int(insertID)
	objType := "environment"

	desc := "Environment " + env.Name + " created"
	if err := (models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &insertIDInt,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveEnvironment(w http.ResponseWriter, r *http.Request) {
	env := context.Get(r, "environment").(models.Environment)

	templatesC, err := database.Mysql.SelectInt("select count(1) from project__template where project_id=? and environment_id=?", env.ProjectID, env.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Environment is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := database.Mysql.Exec("update project__environment set removed=1 where id=?", env.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := database.Mysql.Exec("delete from project__environment where id=?", env.ID); err != nil {
		panic(err)
	}

	desc := "Environment " + env.Name + " deleted"
	if err := (models.Event{
		ProjectID:   &env.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
