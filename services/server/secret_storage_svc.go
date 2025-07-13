package server

import (
	"github.com/semaphoreui/semaphore/db"
	pro "github.com/semaphoreui/semaphore/pro/services/server"
)

type SecretStorageService interface {
	GetSecretStorage(projectID int, storageID int) (db.SecretStorage, error)
	UpdateSecretStorage(storage db.SecretStorage) error
	DeleteSecretStorage(projectID int, storageID int) error
	GetSecretStorages(projectID int) ([]db.SecretStorage, error)
}

func NewSecretStorageService(
	secretStorageRepo db.SecretStorageRepository,
	accessKeyService AccessKeyService,
) SecretStorageService {
	return &SecretStorageServiceImpl{
		secretStorageRepo: secretStorageRepo,
		accessKeyService:  accessKeyService,
	}
}

type SecretStorageServiceImpl struct {
	secretStorageRepo db.SecretStorageRepository
	accessKeyService  AccessKeyService
}

func (s *SecretStorageServiceImpl) DeleteSecretStorage(projectID int, storageID int) error {
	return s.secretStorageRepo.DeleteSecretStorage(projectID, storageID)
}

func (s *SecretStorageServiceImpl) GetSecretStorage(projectID int, storageID int) (res db.SecretStorage, err error) {
	return s.secretStorageRepo.GetSecretStorage(projectID, storageID)
}

func (s *SecretStorageServiceImpl) UpdateSecretStorage(storage db.SecretStorage) (err error) {
	err = s.secretStorageRepo.UpdateSecretStorage(storage)
	if err != nil {
		return
	}

	keys, err := s.accessKeyService.GetAccessKeys(storage.ProjectID, db.GetAccessKeyOptions{
		Owner:     db.AccessKeyVault,
		StorageID: &storage.ID,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		if storage.VaultToken != "" {
			_, err = s.accessKeyService.CreateAccessKey(db.AccessKey{
				Type:      db.AccessKeyString,
				ProjectID: &storage.ProjectID,
				Secret:    nil,
				String:    storage.VaultToken,
				Owner:     db.AccessKeyVault,
				Plain:     nil,
				StorageID: &storage.ID,
			})
		}
	} else {
		vault := keys[0]
		if storage.VaultToken == "" {
			// Do nothing if the vault token is empty,
			// as it means the user haven't set a new token.

			//err = s.keyRepo.DeleteAccessKey(storage.ProjectID, vault.ID)
		} else {
			vault.OverrideSecret = true
			vault.String = storage.VaultToken
			err = s.accessKeyService.UpdateAccessKey(vault)
		}
	}

	return
}

func (s *SecretStorageServiceImpl) GetSecretStorages(projectID int) (storages []db.SecretStorage, err error) {
	return pro.GetSecretStorages(s.secretStorageRepo, projectID)
}
