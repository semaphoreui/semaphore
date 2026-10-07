package projects

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
)

type EnvironmentController struct {
	accessKeyRepo        db.AccessKeyManager
	accessKeyService     server.AccessKeyService
	encryptionService    server.AccessKeyEncryptionService
	environmentService   server.EnvironmentService
	secretStorageService server.SecretStorageService
}

func NewEnvironmentController(
	accessKeyRepo db.AccessKeyManager,
	encryptionService server.AccessKeyEncryptionService,
	accessKeyService server.AccessKeyService,
	environmentService server.EnvironmentService,
	secretStorageService server.SecretStorageService,
) *EnvironmentController {
	return &EnvironmentController{
		accessKeyRepo:        accessKeyRepo,
		accessKeyService:     accessKeyService,
		encryptionService:    encryptionService,
		environmentService:   environmentService,
		secretStorageService: secretStorageService,
	}
}

func (c *EnvironmentController) updateEnvironmentSecrets(env db.Environment) (audit.EnvironmentMetadata, error) {
	var applied audit.EnvironmentMetadata
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
			var storageType *db.AccessKeySourceStorageType

			if env.SecretStorageID != nil {
				keyPrefix := ""
				if env.SecretStorageKeyPrefix != nil {
					keyPrefix = *env.SecretStorageKeyPrefix
				}
				keyPath := keyPrefix + random.String(10)
				sourceStorageKey = &keyPath

				keyType := db.AccessKeySourceStorageVault
				storageType = &keyType
			}

			key, err = c.accessKeyService.Create(db.AccessKey{
				Name:              secret.Name,
				String:            secret.Secret,
				EnvironmentID:     &env.ID,
				ProjectID:         &env.ProjectID,
				Type:              db.AccessKeyString,
				Owner:             secret.Type.GetAccessKeyOwner(),
				SourceStorageID:   env.SecretStorageID,
				SourceStorageKey:  sourceStorageKey,
				SourceStorageType: storageType,
			})

			if err != nil {
				errors = append(errors, err)
				continue
			}
			applied.SecretsCreated++
		case db.EnvironmentSecretDelete:
			key, err = c.accessKeyRepo.GetAccessKey(env.ProjectID, secret.ID)

			if err != nil {
				errors = append(errors, err)
				continue
			}

			if key.EnvironmentID == nil || *key.EnvironmentID != env.ID {
				errors = append(errors, fmt.Errorf("secret does not belong to this environment"))
				continue
			}

			err = c.accessKeyService.Delete(env.ProjectID, secret.ID)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			applied.SecretsDeleted++
		case db.EnvironmentSecretUpdate:
			key, err = c.accessKeyRepo.GetAccessKey(env.ProjectID, secret.ID)

			if err != nil {
				errors = append(errors, err)
				continue
			}

			if key.EnvironmentID == nil || *key.EnvironmentID != env.ID {
				errors = append(errors, fmt.Errorf("secret does not belong to this environment"))
				continue
			}

			updateKey := db.AccessKey{
				ID:                key.ID,
				ProjectID:         key.ProjectID,
				Name:              secret.Name,
				Type:              db.AccessKeyString,
				Owner:             key.Owner,
				SourceStorageID:   env.SecretStorageID,
				SourceStorageType: key.SourceStorageType,
				SourceStorageKey:  key.SourceStorageKey,
			}
			if secret.Secret != "" {
				updateKey.String = secret.Secret
				updateKey.OverrideSecret = true
			}

			err = c.accessKeyService.Update(updateKey)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			applied.SecretsUpdated++
		}
	}

	if len(errors) > 0 {
		return applied, errors[0]
	}

	return applied, nil
}

// EnvironmentMiddleware ensures an environment exists and loads it to the context
func (c *EnvironmentController) EnvironmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		envID, ok := helpers.GetIntParamOrAbort("environment_id", w, r)
		if !ok {
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

	secrets, err := c.updateEnvironmentSecrets(env)
	recordEnvironmentSave(r, audit.ResourceEnvironmentUpdate, env, secrets, err)
	if err != nil {
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
		return
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

	secrets, err := c.updateEnvironmentSecrets(newEnv)
	recordEnvironmentSave(r, audit.ResourceEnvironmentCreate, newEnv, secrets, err)
	if err != nil {
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

	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Environment is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	// The environment row is deleted before its secrets, so a failed secret is a partial success.
	if err != nil && !errors.Is(err, server.ErrSecretsLeftBehind) {
		helpers.WriteError(w, err)
		return
	}

	event := audit.Event{
		Kind:      audit.ResourceEnvironmentDelete,
		Target:    audit.ResourceTarget(audit.TargetEnvironment, env.ID, env.Name),
		ProjectID: env.ProjectID,
		Metadata:  audit.DeleteMetadata{},
	}
	if err != nil {
		event.Reason = audit.ReasonSecretFailed
		event.Metadata = audit.DeleteMetadata{Partial: true}
	}
	helpers.Audit(r).Record(r.Context(), event)

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

// SyncEnvironment triggers a sync of secrets for the environment
func (c *EnvironmentController) SyncEnvironment(w http.ResponseWriter, r *http.Request) {
	env := helpers.GetFromContext(r, "environment").(db.Environment)

	sync, err := helpers.Store(r).GetEnvironmentSecretSync(env.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = c.secretStorageService.SyncSecrets(sync)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   env.ProjectID,
		ObjectType:  db.EventEnvironment,
		ObjectID:    env.ID,
		Description: fmt.Sprintf("Environment %s secrets synced", env.Name),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceEnvironmentSync,
		Target:    audit.ResourceTarget(audit.TargetEnvironment, env.ID, env.Name),
		ProjectID: env.ProjectID,
	})

	w.WriteHeader(http.StatusNoContent)
}

// The environment row is stored before its secrets, so a failed secret is a partial success.
func recordEnvironmentSave(r *http.Request, kind audit.Kind, env db.Environment, secrets audit.EnvironmentMetadata, err error) {
	event := audit.Event{
		Kind:      kind,
		Target:    audit.ResourceTarget(audit.TargetEnvironment, env.ID, env.Name),
		ProjectID: env.ProjectID,
	}
	if err != nil {
		secrets.Partial = true
		event.Reason = audit.ReasonSecretFailed
	}
	event.Metadata = secrets
	helpers.Audit(r).Record(r.Context(), event)
}
