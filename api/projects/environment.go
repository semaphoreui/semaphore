package projects

import (
	"errors"
	"fmt"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/services/server"
	"net/http"
)

type EnvironmentController struct {
	accessKeyRepo      db.AccessKeyManager
	accessKeyService   server.AccessKeyService
	encryptionService  server.AccessKeyEncryptionService
	environmentService server.EnvironmentService
}

func NewEnvironmentController(
	accessKeyRepo db.AccessKeyManager,
	encryptionService server.AccessKeyEncryptionService,
	accessKeyService server.AccessKeyService,
	environmentService server.EnvironmentService,
) *EnvironmentController {
	return &EnvironmentController{
		accessKeyRepo:      accessKeyRepo,
		accessKeyService:   accessKeyService,
		encryptionService:  encryptionService,
		environmentService: environmentService,
	}
}

func (c *EnvironmentController) updateEnvironmentSecrets(env db.Environment) error {

	errors := make([]error, 0)

	for _, secret := range env.Secrets {
		err := secret.Validate()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		var key db.AccessKey

		switch secret.Operation {
		case db.EnvironmentSecretCreate:
			var sourceStorageKey *string
			if env.SecretStorageKeyPrefix != nil {
				tmp := *env.SecretStorageKeyPrefix + random.String(10)
				sourceStorageKey = &tmp
			}

			key, err = c.accessKeyService.Create(db.AccessKey{
				Name:             secret.Name,
				String:           secret.Secret,
				EnvironmentID:    &env.ID,
				ProjectID:        &env.ProjectID,
				Type:             db.AccessKeyString,
				Owner:            secret.Type.GetAccessKeyOwner(),
				SourceStorageID:  env.SecretStorageID,
				SourceStorageKey: sourceStorageKey,
			})

			if err != nil {
				errors = append(errors, err)
				continue
			}
		case db.EnvironmentSecretDelete:
			key, err = c.accessKeyRepo.GetAccessKey(env.ProjectID, secret.ID)

			if err != nil {
				errors = append(errors, err)
				continue
			}

			if key.EnvironmentID == nil && *key.EnvironmentID == env.ID {
				errors = append(errors, err)
				continue
			}

			err = c.accessKeyService.Delete(env.ProjectID, secret.ID)
		case db.EnvironmentSecretUpdate:
			key, err = c.accessKeyRepo.GetAccessKey(env.ProjectID, secret.ID)

			if err != nil {
				errors = append(errors, err)
				continue
			}

			if key.EnvironmentID == nil && *key.EnvironmentID == env.ID {
				errors = append(errors, err)
				continue
			}

			updateKey := db.AccessKey{
				ID:        key.ID,
				ProjectID: key.ProjectID,
				Name:      secret.Name,
				Type:      db.AccessKeyString,
				Owner:     key.Owner,
			}
			if secret.Secret != "" {
				updateKey.String = secret.Secret
				updateKey.OverrideSecret = true
			}

			err = c.accessKeyService.Update(updateKey)
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// EnvironmentMiddleware ensures an environment exists and loads it to the context
func (c *EnvironmentController) EnvironmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
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

		if err = c.encryptionService.FillEnvironmentSecrets(&env, false); err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "environment", env)
		next.ServeHTTP(w, r)
	})
}

func GetEnvironmentRefs(w http.ResponseWriter, r *http.Request) {
	env := helpers.GetFromContext(r, "environment").(db.Environment)
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
	if environment := helpers.GetFromContext(r, "environment"); environment != nil {
		helpers.WriteJSON(w, http.StatusOK, environment.(db.Environment))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	env, err := helpers.Store(r).GetEnvironments(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, env)
}

// UpdateEnvironment updates an existing environment in the database
func (c *EnvironmentController) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	oldEnv := helpers.GetFromContext(r, "environment").(db.Environment)
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

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldEnv.ProjectID,
		ObjectType:  db.EventEnvironment,
		ObjectID:    oldEnv.ID,
		Description: fmt.Sprintf("Environment %s updated", env.Name),
	})

	if err := c.updateEnvironmentSecrets(env); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEnvironment creates an environment in the database
func (c *EnvironmentController) AddEnvironment(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
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

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newEnv.ProjectID,
		ObjectType:  db.EventEnvironment,
		ObjectID:    newEnv.ID,
		Description: fmt.Sprintf("Environment %s created", newEnv.Name),
	})

	if err = c.updateEnvironmentSecrets(newEnv); err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Reload env
	env, err = helpers.Store(r).GetEnvironment(newEnv.ProjectID, newEnv.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	// Use empty array to avoid null in JSON
	env.Secrets = []db.EnvironmentSecret{}

	helpers.WriteJSON(w, http.StatusCreated, env)
}

// RemoveEnvironment deletes an environment from the database
func (c *EnvironmentController) RemoveEnvironment(w http.ResponseWriter, r *http.Request) {
	env := helpers.GetFromContext(r, "environment").(db.Environment)

	err := c.environmentService.Delete(env.ProjectID, env.ID)
	//err := helpers.Store(r).DeleteEnvironment(env.ProjectID, env.ID)

	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Environment is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   env.ProjectID,
		ObjectType:  db.EventEnvironment,
		ObjectID:    env.ID,
		Description: fmt.Sprintf("Environment %s deleted", env.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}
