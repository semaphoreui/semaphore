package server

import (
	"errors"
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

// ErrSecretsLeftBehind means the object row is deleted but some of its secrets were not.
var ErrSecretsLeftBehind = errors.New("deleted, but some secrets were not removed")

type EnvironmentService interface {
	Delete(projectID int, environmentID int) error

	// ApplySecrets runs the create / update / delete operations in env.Secrets against the group
	// env.ID. A secret with a value becomes an Access Key owned by the group plus a link; a secret
	// with a key id becomes a link to that key only. The counters tell what was applied; the first
	// failure is returned after every operation has been tried.
	ApplySecrets(env db.Environment) (audit.EnvironmentMetadata, error)
}

func NewEnvironmentService(
	environmentRepo db.EnvironmentManager,
	encryptionService AccessKeyEncryptionService,
	secretStorageRepo db.SecretStorageRepository,
	accessKeyRepo db.AccessKeyManager,
	accessKeyService AccessKeyService,
) EnvironmentService {
	return &EnvironmentServiceImpl{
		environmentRepo:   environmentRepo,
		encryptionService: encryptionService,
		secretStorageRepo: secretStorageRepo,
		accessKeyRepo:     accessKeyRepo,
		accessKeyService:  accessKeyService,
	}
}

type EnvironmentServiceImpl struct {
	environmentRepo   db.EnvironmentManager
	encryptionService AccessKeyEncryptionService
	secretStorageRepo db.SecretStorageRepository
	accessKeyRepo     db.AccessKeyManager
	accessKeyService  AccessKeyService
}

func (s *EnvironmentServiceImpl) Delete(projectID int, environmentID int) (err error) {
	if projectID <= 0 || environmentID <= 0 {
		return fmt.Errorf("invalid project or environment ID")
	}

	env, err := s.environmentRepo.GetEnvironment(projectID, environmentID)
	if err != nil {
		return
	}

	secrets, err := s.environmentRepo.GetEnvironmentOwnedKeys(projectID, environmentID)
	if err != nil {
		return
	}

	// The group's own keys go with it by cascade; a key another group still links to would stop that cascade
	// at the foreign key half-way through. Refuse up front and name the group that has to unlink first.
	for _, key := range secrets {
		var users []db.ObjectReferrer
		users, err = s.environmentRepo.GetAccessKeyEnvironmentRefs(projectID, key.ID)
		if err != nil {
			return
		}
		for _, user := range users {
			if user.ID == environmentID {
				continue
			}
			return common_errors.NewValidationError(fmt.Sprintf(
				"secret %q is used by variable group %q; unlink it there first", key.Name, user.Name))
		}
	}

	// Read before the delete, so a failed lookup leaves the environment in place.
	var storage *db.SecretStorage
	if env.SecretStorageID != nil {
		var res db.SecretStorage
		res, err = s.secretStorageRepo.GetSecretStorage(projectID, *env.SecretStorageID)
		if err != nil {
			return
		}
		storage = &res
	}

	err = s.environmentRepo.DeleteEnvironment(projectID, environmentID)

	if err != nil {
		return
	}

	var errs []error

	// The group's own keys cascade with it on SQLite and Postgres; MySQL ignores the inline cascade of
	// access_key.environment_id and would leave them behind as orphans. Their links are gone with the group.
	for _, key := range secrets {
		if delErr := s.accessKeyRepo.DeleteAccessKey(projectID, key.ID); delErr != nil && !errors.Is(delErr, db.ErrNotFound) {
			errs = append(errs, delErr)
		}
	}

	if storage != nil {
		if !storage.ReadOnly {
			for _, secret := range secrets {
				if secret.Synchronized {
					continue
				}
				err = s.encryptionService.DeleteSecret(&secret)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf("%w: failed to delete some secrets: %v", ErrSecretsLeftBehind, errs)
		return
	}

	return
}

func (s *EnvironmentServiceImpl) ApplySecrets(env db.Environment) (applied audit.EnvironmentMetadata, err error) {
	var errs []error

	for _, secret := range env.Secrets {
		secret.ProjectID = env.ProjectID
		secret.EnvironmentID = env.ID

		opErr := secret.Validate()
		if opErr == nil {
			switch secret.Operation {
			case db.EnvironmentSecretCreate:
				if opErr = s.createSecret(env, secret); opErr == nil {
					applied.SecretsCreated++
				}
			case db.EnvironmentSecretUpdate:
				if opErr = s.updateSecret(env, secret); opErr == nil {
					applied.SecretsUpdated++
				}
			case db.EnvironmentSecretDelete:
				if opErr = s.deleteSecret(env, secret); opErr == nil {
					applied.SecretsDeleted++
				}
			}
		}

		if opErr != nil {
			errs = append(errs, opErr)
		}
	}

	if len(errs) > 0 {
		err = errs[0]
	}

	return
}

// linkableKey loads a key of the project and tells whether a group may point at it.
func (s *EnvironmentServiceImpl) linkableKey(projectID int, keyID int) (key db.AccessKey, err error) {
	key, err = s.accessKeyRepo.GetAccessKey(projectID, keyID)
	if errors.Is(err, db.ErrNotFound) {
		err = common_errors.NewValidationError(fmt.Sprintf("key %d does not exist in this project", keyID))
	}
	if err != nil {
		return
	}

	err = db.ValidateEnvironmentSecretKey(key)
	return
}

func ownedBy(key db.AccessKey, env db.Environment) bool {
	return key.EnvironmentID != nil && *key.EnvironmentID == env.ID
}

// linkRow is the persisted part of a secret: the value and the operation stay out of the store call.
func linkRow(env db.Environment, secret db.EnvironmentSecret, keyID int) db.EnvironmentSecret {
	return db.EnvironmentSecret{
		ID:            secret.ID,
		ProjectID:     env.ProjectID,
		EnvironmentID: env.ID,
		Type:          secret.Type,
		Name:          secret.Name,
		KeyID:         keyID,
	}
}

func (s *EnvironmentServiceImpl) createSecret(env db.Environment, secret db.EnvironmentSecret) error {
	if secret.KeyID != 0 {
		if _, err := s.linkableKey(env.ProjectID, secret.KeyID); err != nil {
			return err
		}
		_, err := s.environmentRepo.CreateEnvironmentSecret(linkRow(env, secret, secret.KeyID))
		return err
	}

	key, err := s.accessKeyService.Create(s.ownedKey(env, secret))
	if err != nil {
		return err
	}

	if _, err = s.environmentRepo.CreateEnvironmentSecret(linkRow(env, secret, key.ID)); err != nil {
		// Without the link the key is invisible: not in the Key Store, not in the group.
		if delErr := s.accessKeyService.Delete(env.ProjectID, key.ID); delErr != nil {
			log.WithError(delErr).WithFields(log.Fields{
				"project_id": env.ProjectID, "environment_id": env.ID, "key_id": key.ID,
			}).Error("failed to remove the key of a secret that could not be linked")
		}
		return err
	}

	return nil
}

// ownedKey builds the Access Key that holds the value of a secret created from the group form.
func (s *EnvironmentServiceImpl) ownedKey(env db.Environment, secret db.EnvironmentSecret) db.AccessKey {
	envID := env.ID
	projectID := env.ProjectID
	key := db.AccessKey{
		Name:            secret.Name,
		String:          secret.Secret,
		EnvironmentID:   &envID,
		ProjectID:       &projectID,
		Type:            db.AccessKeyString,
		Owner:           secret.Type.GetAccessKeyOwner(),
		SourceStorageID: env.SecretStorageID,
	}

	if env.SecretStorageID != nil {
		keyPrefix := ""
		if env.SecretStorageKeyPrefix != nil {
			keyPrefix = *env.SecretStorageKeyPrefix
		}
		keyPath := keyPrefix + random.String(10)
		storageType := db.AccessKeySourceStorageVault
		key.SourceStorageKey = &keyPath
		key.SourceStorageType = &storageType
	}

	return key
}

func (s *EnvironmentServiceImpl) updateSecret(env db.Environment, secret db.EnvironmentSecret) error {
	current, err := s.environmentRepo.GetEnvironmentSecret(env.ProjectID, env.ID, secret.ID)
	if errors.Is(err, db.ErrNotFound) {
		return common_errors.NewValidationError("secret does not belong to this environment")
	}
	if err != nil {
		return err
	}

	key, err := s.accessKeyRepo.GetAccessKey(env.ProjectID, current.KeyID)
	if err != nil {
		return err
	}
	owned := ownedBy(key, env)

	if secret.Secret != "" {
		if !owned {
			return common_errors.NewValidationError(
				"secret " + secret.Name + " is linked to key " + key.Name + "; change its value in the Key Store or unlink it")
		}
		key.Name = secret.Name
		key.Type = db.AccessKeyString
		key.String = secret.Secret
		key.OverrideSecret = true
		if err = s.accessKeyService.Update(key); err != nil {
			return err
		}
	} else if owned && key.Name != secret.Name && secret.KeyID == 0 {
		// The owned key is named after the secret, so refs and the audit log read naturally.
		key.Name = secret.Name
		if err = s.accessKeyService.Update(key); err != nil {
			return err
		}
	}

	newKeyID := current.KeyID
	if secret.KeyID != 0 && secret.KeyID != current.KeyID {
		if _, err = s.linkableKey(env.ProjectID, secret.KeyID); err != nil {
			return err
		}
		newKeyID = secret.KeyID
	}

	if err = s.environmentRepo.UpdateEnvironmentSecret(linkRow(env, secret, newKeyID)); err != nil {
		return err
	}

	if newKeyID != current.KeyID && owned {
		return s.dropOwnedKeyIfUnused(env, key)
	}

	return nil
}

func (s *EnvironmentServiceImpl) deleteSecret(env db.Environment, secret db.EnvironmentSecret) error {
	current, err := s.environmentRepo.GetEnvironmentSecret(env.ProjectID, env.ID, secret.ID)
	if errors.Is(err, db.ErrNotFound) {
		return common_errors.NewValidationError("secret does not belong to this environment")
	}
	if err != nil {
		return err
	}

	key, err := s.accessKeyRepo.GetAccessKey(env.ProjectID, current.KeyID)
	if err != nil {
		return err
	}

	if err = s.environmentRepo.DeleteEnvironmentSecret(env.ProjectID, env.ID, current.ID); err != nil {
		return err
	}

	if ownedBy(key, env) {
		return s.dropOwnedKeyIfUnused(env, key)
	}

	return nil
}

// dropOwnedKeyIfUnused removes the group's own key once no secret of any group points at it. A key
// another group still links to stays, with this group as its owner.
func (s *EnvironmentServiceImpl) dropOwnedKeyIfUnused(env db.Environment, key db.AccessKey) error {
	n, err := s.environmentRepo.CountEnvironmentSecretsByKey(env.ProjectID, key.ID)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.accessKeyService.Delete(env.ProjectID, key.ID)
}
