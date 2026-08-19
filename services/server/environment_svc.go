package server

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

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

	var errors []error

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
					errors = append(errors, err)
				}
			}
		}
	}

	if len(errors) > 0 {
		err = fmt.Errorf("failed to delete some secrets: %v", errors)
		return
	}

	return
}
