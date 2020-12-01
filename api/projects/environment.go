package projects

import (
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// EnvironmentMiddleware ensures an environment exists and loads it to the context
func EnvironmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(models.Project)
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

		var env models.Environment
		if err := context.Get(r, "store").(db.Store).Sql().SelectOne(&env, query, args...); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		context.Set(r, "environment", env)
		next.ServeHTTP(w, r)
	})
}

// GetEnvironment retrieves sorted environments from the database
func GetEnvironment(w http.ResponseWriter, r *http.Request) {
	if environment := context.Get(r, "environment"); environment != nil {
		util.WriteJSON(w, http.StatusOK, environment.(models.Environment))
		return
	}

	project := context.Get(r, "project").(models.Project)
	var env []models.Environment

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

	if _, err := context.Get(r, "store").(db.Store).Sql().Select(&env, query, args...); err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusOK, env)
}

// UpdateEnvironment updates an existing environment in the database
func UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	oldEnv := context.Get(r, "environment").(models.Environment)
	var env models.Environment
	if err := util.Bind(w, r, &env); err != nil {
		return
	}

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	if _, err := context.Get(r, "store").(db.Store).Sql().Exec("update project__environment set name=?, json=? where id=?", env.Name, env.JSON, oldEnv.ID); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEnvironment creates an environment in the database
func AddEnvironment(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var env models.Environment

	if err := util.Bind(w, r, &env); err != nil {
		return
	}

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		util.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	res, err := context.Get(r, "store").(db.Store).Sql().Exec("insert into project__environment (project_id, name, json, password) values (?, ?, ?, ?)", project.ID, env.Name, env.JSON, env.Password)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	util.LogWarning(err)
	insertIDInt := int(insertID)
	objType := "environment"

	desc := "Environment " + env.Name + " created"
	_, err = context.Get(r, "store").(db.Store).CreateEvent(models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &insertIDInt,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveEnvironment deletes an environment from the database
func RemoveEnvironment(w http.ResponseWriter, r *http.Request) {
	env := context.Get(r, "environment").(models.Environment)

	templatesC, err := context.Get(r, "store").(db.Store).Sql().SelectInt("select count(1) from project__template where project_id=? and environment_id=?", env.ProjectID, env.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Environment is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := context.Get(r, "store").(db.Store).Sql().Exec("update project__environment set removed=1 where id=?", env.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := context.Get(r, "store").(db.Store).Sql().Exec("delete from project__environment where id=?", env.ID); err != nil {
		panic(err)
	}

	desc := "Environment " + env.Name + " deleted"
	_, err = context.Get(r, "store").(db.Store).CreateEvent(models.Event{
		ProjectID:   &env.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
