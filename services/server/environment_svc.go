package server

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	pro "github.com/semaphoreui/semaphore/pro/services/server"
)

type EnvironmentService interface {
	Delete(projectID int, environmentID int) error
	SyncSecrets(env db.Environment) error
}

func NewEnvironmentService(
	environmentRepo db.EnvironmentManager,
	secretStorageRepo db.SecretStorageRepository,
	accessKeyRepo db.AccessKeyManager,
	accessKeyService AccessKeyService,
	encryptionService AccessKeyEncryptionService,
) EnvironmentService {
	return &EnvironmentServiceImpl{
		environmentRepo:   environmentRepo,
		secretStorageRepo: secretStorageRepo,
		accessKeyRepo:     accessKeyRepo,
		accessKeyService:  accessKeyService,
		encryptionService: encryptionService,
	}
}

type EnvironmentServiceImpl struct {
	environmentRepo   db.EnvironmentManager
	secretStorageRepo db.SecretStorageRepository
	accessKeyRepo     db.AccessKeyManager
	accessKeyService  AccessKeyService
	encryptionService AccessKeyEncryptionService
}

func (s *EnvironmentServiceImpl) SyncSecrets(env db.Environment) error {
	return pro.SyncEnvironmentSecrets(env, s.secretStorageRepo, s.accessKeyRepo, s.encryptionService)
}

func (s *EnvironmentServiceImpl) Delete(projectID int, environmentID int) (err error) {
	// Implement the logic to delete an environment
	// This is a placeholder implementation
	if projectID <= 0 || environmentID <= 0 {
		return fmt.Errorf("invalid project or environment ID")
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

	for _, secret := range secrets {
		err = s.encryptionService.DeleteSecret(&secret)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		err = fmt.Errorf("failed to delete some secrets: %v", errors)
		return
	}

	return
}
