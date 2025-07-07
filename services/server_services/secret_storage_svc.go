package server_services

import "github.com/semaphoreui/semaphore/db"

type SecretStorageService interface {
	GetSecretStorage(projectID int, storageID int) (db.SecretStorage, error)
	UpdateSecretStorage(storage db.SecretStorage) error
}

func NewSecretStorageService(
	secretStorageRepo db.SecretStorageRepository,
	keyRepo db.AccessKeyManager,
) SecretStorageService {
	return &SecretStorageServiceImpl{
		secretStorageRepo: secretStorageRepo,
		keyRepo:           keyRepo,
	}
}

type SecretStorageServiceImpl struct {
	secretStorageRepo db.SecretStorageRepository
	keyRepo           db.AccessKeyManager
}

func (s *SecretStorageServiceImpl) GetSecretStorage(projectID int, storageID int) (res db.SecretStorage, err error) {
	return s.secretStorageRepo.GetSecretStorage(projectID, storageID)
}

func (s *SecretStorageServiceImpl) UpdateSecretStorage(storage db.SecretStorage) (err error) {
	err = s.secretStorageRepo.UpdateSecretStorage(storage)
	if err != nil {
		return
	}

	keys, err := s.keyRepo.GetAccessKeys(storage.ProjectID, db.GetAccessKeyOptions{
		Owner:     db.AccessKeyVault,
		StorageID: &storage.ID,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		if storage.VaultToken != "" {
			_, err = s.keyRepo.CreateAccessKey(db.AccessKey{
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
			err = s.keyRepo.DeleteAccessKey(storage.ProjectID, vault.ID)
		} else {
			vault.OverrideSecret = true
			vault.String = storage.VaultToken
			err = s.keyRepo.UpdateAccessKey(vault)
		}
	}

	return
}
