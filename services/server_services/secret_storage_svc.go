package server_services

import "github.com/semaphoreui/semaphore/db"

type SecretStorageService interface {
}

type SecretStorageServiceImpl struct {
	secretStorageRepo db.SecretStorageRepository
}

func NewSecretStorageService(secretStorageRepo db.SecretStorageRepository) SecretStorageService {
	return &SecretStorageServiceImpl{
		secretStorageRepo: secretStorageRepo,
	}
}
