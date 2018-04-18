package projects

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// EnvironmentMiddleware ensures an environment exists and loads it to the context
func EnvironmentMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	envID, err := util.GetIntParam("environment_id", w, r)
	if err != nil {
		return
	}

	query, args, err := squirrel.Select("*").
		From("project__environment").
		Where("project_id=?", project.ID).
		Where("id=?", envID).
		ToSql()
	util.LogWarning(err)

	var env db.Environment
	if err := db.Mysql.SelectOne(&env, query, args...); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "environment", env)
}

// GetEnvironment retrieves sorted environments from the database
func GetEnvironment(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var env []db.Environment

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != "asc" && order != "desc" {
		order = "asc"
	}

	q := squirrel.Select("*").
		From("project__environment pe").
		Where("project_id=?", project.ID)

	switch sort {
	case "name":
		q = q.Where("pe.project_id=?", project.ID).
			OrderBy("pe." + sort + " " + order)
	default:
		q = q.Where("pe.project_id=?", project.ID).
			OrderBy("pe.name " + order)
	}

	query, args, err := q.ToSql()
	util.LogWarning(err)

	if _, err := db.Mysql.Select(&env, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, env)
}

// UpdateEnvironment updates an existing environment in the database
func UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	oldEnv := context.Get(r, "environment").(db.Environment)
	var env db.Environment
	if err := mulekick.Bind(w, r, &env); err != nil {
		return
	}

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	if _, err := db.Mysql.Exec("update project__environment set name=?, json=? where id=?", env.Name, env.JSON, oldEnv.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEnvironment creates an environment in the database
func AddEnvironment(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var env db.Environment

	if err := mulekick.Bind(w, r, &env); err != nil {
		return
	}

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		mulekick.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	res, err := db.Mysql.Exec("insert into project__environment set project_id=?, name=?, json=?, password=?", project.ID, env.Name, env.JSON, env.Password)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	util.LogWarning(err)
	insertIDInt := int(insertID)
	objType := "environment"

	desc := "Environment " + env.Name + " created"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &insertIDInt,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveEnvironment deletes an environment from the database
func RemoveEnvironment(w http.ResponseWriter, r *http.Request) {
	env := context.Get(r, "environment").(db.Environment)

	templatesC, err := db.Mysql.SelectInt("select count(1) from project__template where project_id=? and environment_id=?", env.ProjectID, env.ID)
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

		if _, err := db.Mysql.Exec("update project__environment set removed=1 where id=?", env.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.Mysql.Exec("delete from project__environment where id=?", env.ID); err != nil {
		panic(err)
	}

	desc := "Environment " + env.Name + " deleted"
	if err := (db.Event{
		ProjectID:   &env.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
