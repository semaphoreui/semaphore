package server

import (
	"errors"
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

// ErrSecretsLeftBehind means the object row is deleted but some of its secrets were not.
var ErrSecretsLeftBehind = errors.New("deleted, but some secrets were not removed")

type EnvironmentService interface {
	Delete(projectID int, environmentID int) error
}

func NewEnvironmentService(
	environmentRepo db.EnvironmentManager,
	encryptionService AccessKeyEncryptionService,
	secretStorageRepo db.SecretStorageRepository,
) EnvironmentService {
	return &EnvironmentServiceImpl{
		environmentRepo:   environmentRepo,
		encryptionService: encryptionService,
		secretStorageRepo: secretStorageRepo,
	}
}

type EnvironmentServiceImpl struct {
	environmentRepo   db.EnvironmentManager
	encryptionService AccessKeyEncryptionService
	secretStorageRepo db.SecretStorageRepository
}

func (s *EnvironmentServiceImpl) Delete(projectID int, environmentID int) (err error) {
	if projectID <= 0 || environmentID <= 0 {
		return fmt.Errorf("invalid project or environment ID")
	}

	env, err := s.environmentRepo.GetEnvironment(projectID, environmentID)
	if err != nil {
		return
	}

	secrets, err := s.environmentRepo.GetEnvironmentSecrets(projectID, environmentID)
	if err != nil {
		return
	}

	err = s.environmentRepo.DeleteEnvironment(projectID, environmentID)

	if err != nil {
		return
	}

	var errs []error

	if env.SecretStorageID != nil {
		var storage db.SecretStorage
		storage, err = s.secretStorageRepo.GetSecretStorage(projectID, *env.SecretStorageID)
		if err != nil {
			return
		}

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
