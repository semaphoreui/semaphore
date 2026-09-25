package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// HostConfigMiddleware ensures a mapping exists and loads it to the context
func HostConfigMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		hostConfigID, ok := helpers.GetIntParamOrAbort("host_config_id", w, r)
		if !ok {
			return
		}

		hostConfig, err := helpers.Store(r).GetHostConfig(project.ID, hostConfigID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "host_config", hostConfig)
		next.ServeHTTP(w, r)
	})
}

// GetHostConfigs returns a single mapping, or all mappings of the project
func GetHostConfigs(w http.ResponseWriter, r *http.Request) {
	if hostConfig := helpers.GetFromContext(r, "host_config"); hostConfig != nil {
		helpers.WriteJSON(w, http.StatusOK, hostConfig.(db.HostConfig))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	hostConfigs, err := helpers.Store(r).GetHostConfigs(
		project.ID, helpers.QueryParamsForProps(r.URL, db.HostConfigProps))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, hostConfigs)
}

// AddHostConfig creates a mapping in the database
func AddHostConfig(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var hostConfig db.HostConfig

	if !helpers.Bind(w, r, &hostConfig) {
		return
	}

	if hostConfig.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := hostConfig.Validate(); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := db.ValidateHostConfig(helpers.Store(r), &hostConfig); err != nil {
		helpers.WriteError(w, err)
		return
	}

	newHostConfig, err := helpers.Store(r).CreateHostConfig(hostConfig)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newHostConfig.ProjectID,
		ObjectType:  db.EventHostConfig,
		ObjectID:    newHostConfig.ID,
		Description: fmt.Sprintf("Host config for %s created", newHostConfig.Name),
	})

	helpers.WriteJSON(w, http.StatusCreated, newHostConfig)
}

// UpdateHostConfig writes a mapping to the database
func UpdateHostConfig(w http.ResponseWriter, r *http.Request) {
	oldHostConfig := helpers.GetFromContext(r, "host_config").(db.HostConfig)

	var hostConfig db.HostConfig

	if !helpers.Bind(w, r, &hostConfig) {
		return
	}

	if hostConfig.ID != oldHostConfig.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Host config ID in body and URL must be the same",
		})
		return
	}

	if hostConfig.ProjectID != oldHostConfig.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := hostConfig.Validate(); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := db.ValidateHostConfig(helpers.Store(r), &hostConfig); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).UpdateHostConfig(hostConfig); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   hostConfig.ProjectID,
		ObjectType:  db.EventHostConfig,
		ObjectID:    hostConfig.ID,
		Description: fmt.Sprintf("Host config for %s updated", hostConfig.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveHostConfig deletes a mapping from the database
func RemoveHostConfig(w http.ResponseWriter, r *http.Request) {
	hostConfig := helpers.GetFromContext(r, "host_config").(db.HostConfig)

	if err := helpers.Store(r).DeleteHostConfig(hostConfig.ProjectID, hostConfig.ID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   hostConfig.ProjectID,
		ObjectType:  db.EventHostConfig,
		ObjectID:    hostConfig.ID,
		Description: fmt.Sprintf("Host config for %s deleted", hostConfig.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}
