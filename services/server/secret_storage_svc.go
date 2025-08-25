package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	pro "github.com/semaphoreui/semaphore/pro/services/server"
)

type SecretStorageService interface {
	GetSecretStorage(projectID int, storageID int) (db.SecretStorage, error)
	Update(storage db.SecretStorage) error
	Delete(projectID int, storageID int) error
	GetSecretStorages(projectID int) ([]db.SecretStorage, error)
	Create(storage db.SecretStorage) (res db.SecretStorage, err error)
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

func (s *SecretStorageServiceImpl) Delete(projectID int, storageID int) (err error) {
	err = s.secretStorageRepo.DeleteSecretStorage(projectID, storageID)
	if err != nil {
		return
	}

	keys, err := s.accessKeyService.GetAll(projectID, db.GetAccessKeyOptions{
		Owner:     db.AccessKeyVault,
		StorageID: &storageID,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	for _, key := range keys {
		err = s.accessKeyService.Delete(projectID, key.ID)
	}

	return
}

func (s *SecretStorageServiceImpl) GetSecretStorage(projectID int, storageID int) (res db.SecretStorage, err error) {
	return s.secretStorageRepo.GetSecretStorage(projectID, storageID)
}

func (s *SecretStorageServiceImpl) Create(storage db.SecretStorage) (res db.SecretStorage, err error) {
	res, err = s.secretStorageRepo.CreateSecretStorage(storage)

	if err != nil {
		return
	}

	key := db.AccessKey{
		Name:      random.String(10),
		Type:      db.AccessKeyString,
		ProjectID: &storage.ProjectID,
		String:    storage.VaultToken,
		Owner:     db.AccessKeyVault,
		StorageID: &res.ID,
	}

	_, err = s.accessKeyService.Create(key)

	return
}

func (s *SecretStorageServiceImpl) Update(storage db.SecretStorage) (err error) {
	err = s.secretStorageRepo.UpdateSecretStorage(storage)
	if err != nil {
		return
	}

	keys, err := s.accessKeyService.GetAll(storage.ProjectID, db.GetAccessKeyOptions{
		Owner:     db.AccessKeyVault,
		StorageID: &storage.ID,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		if storage.VaultToken != "" {
			_, err = s.accessKeyService.Create(db.AccessKey{
				Name:      random.String(10),
				Type:      db.AccessKeyString,
				ProjectID: &storage.ProjectID,
				String:    storage.VaultToken,
				Owner:     db.AccessKeyVault,
				StorageID: &storage.ID,
			})
		} else {
			// empty vault token means the user didn't set a new token,
			// so we don't create a new access key.
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
			err = s.accessKeyService.Update(vault)
		}
	}

	return
}

func (s *SecretStorageServiceImpl) GetSecretStorages(projectID int) (storages []db.SecretStorage, err error) {
	return pro.GetSecretStorages(s.secretStorageRepo, projectID)
}
