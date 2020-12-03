package projects

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
)

// EnvironmentMiddleware ensures an environment exists and loads it to the context
func EnvironmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(models.Project)
		envID, err := helpers.GetIntParam("environment_id", w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		env, err := helpers.Store(r).GetEnvironment(project.ID, envID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "environment", env)
		next.ServeHTTP(w, r)
	})
}

// GetEnvironment retrieves sorted environments from the database
func GetEnvironment(w http.ResponseWriter, r *http.Request) {

	// return single environment if request has environment ID
	if environment := context.Get(r, "environment"); environment != nil {
		helpers.WriteJSON(w, http.StatusOK, environment.(models.Environment))
		return
	}

	project := context.Get(r, "project").(models.Project)

	params := db.RetrieveQueryParams{
		SortBy: r.URL.Query().Get("sort"),
		SortInverted: r.URL.Query().Get("order") == desc,
	}

	env, err := helpers.Store(r).GetEnvironments(project.ID, params)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, env)
}

// UpdateEnvironment updates an existing environment in the database
func UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	oldEnv := context.Get(r, "environment").(models.Environment)
	var env models.Environment
	if !helpers.Bind(w, r, &env) {
		return
	}

	if env.ID != oldEnv.ID || env.ProjectID != oldEnv.ProjectID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	if err := helpers.Store(r).UpdateEnvironment(env); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEnvironment creates an environment in the database
func AddEnvironment(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(models.Project)
	var env models.Environment

	if !helpers.Bind(w, r, &env) {
		return
	}

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	res, err := helpers.Store(r).Sql().Exec("insert into project__environment (project_id, name, json, password) values (?, ?, ?, ?)", project.ID, env.Name, env.JSON, env.Password)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	util.LogWarning(err)
	insertIDInt := int(insertID)
	objType := "environment"

	desc := "Environment " + env.Name + " created"
	_, err = helpers.Store(r).CreateEvent(models.Event{
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

	templatesC, err := helpers.Store(r).Sql().SelectInt(
		"select count(1) from project__template where project_id=? and environment_id=?",
		env.ProjectID,
		env.ID)

	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Environment is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := helpers.Store(r).Sql().Exec("update project__environment set removed=1 where id=?", env.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := helpers.Store(r).Sql().Exec("delete from project__environment where id=?", env.ID); err != nil {
		panic(err)
	}

	desc := "Environment " + env.Name + " deleted"
	_, err = helpers.Store(r).CreateEvent(models.Event{
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
