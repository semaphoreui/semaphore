package projects

import (
	log "github.com/sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

// EnvironmentMiddleware ensures an environment exists and loads it to the context
func EnvironmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
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

func GetEnvironmentRefs(w http.ResponseWriter, r *http.Request) {
	env := context.Get(r, "environment").(db.Environment)
	refs, err := helpers.Store(r).GetEnvironmentRefs(env.ProjectID, env.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// GetEnvironment retrieves sorted environments from the database
func GetEnvironment(w http.ResponseWriter, r *http.Request) {

	// return single environment if request has environment ID
	if environment := context.Get(r, "environment"); environment != nil {
		helpers.WriteJSON(w, http.StatusOK, environment.(db.Environment))
		return
	}

	project := context.Get(r, "project").(db.Project)

	env, err := helpers.Store(r).GetEnvironments(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, env)
}

// UpdateEnvironment updates an existing environment in the database
func UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	oldEnv := context.Get(r, "environment").(db.Environment)
	var env db.Environment
	if !helpers.Bind(w, r, &env) {
		return
	}

	if env.ID != oldEnv.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Environment ID in body and URL must be the same",
		})
		return
	}

	if env.ProjectID != oldEnv.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
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
	project := context.Get(r, "project").(db.Project)
	var env db.Environment

	if !helpers.Bind(w, r, &env) {
		return
	}

	if project.ID != env.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
	}

	newEnv, err := helpers.Store(r).CreateEnvironment(env)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	objType := db.EventEnvironment

	desc := "Environment " + newEnv.Name + " created"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &newEnv.ID,
		ObjectType:  &objType,
		ObjectID:    &newEnv.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveEnvironment deletes an environment from the database
func RemoveEnvironment(w http.ResponseWriter, r *http.Request) {
	env := context.Get(r, "environment").(db.Environment)

	var err error

	err = helpers.Store(r).DeleteEnvironment(env.ProjectID, env.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Environment is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	desc := "Environment " + env.Name + " deleted"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &env.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
