package server

import (
	"errors"

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
		Owner:     db.AccessKeySecretStorage,
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
	sourceStorageType := storage.SourceStorageType
	sourceStorageKey := ""

	if storage.Secret == "" {
		err = errors.New("secret must be set")
		return
	}

	if sourceStorageType != nil {
		switch *sourceStorageType {
		case db.AccessKeySourceStorageEnv:
			sourceStorageKey = storage.Secret
		case db.AccessKeySourceStorageFile:
			sourceStorageKey = storage.Secret
		default:
			err = errors.New("unsupported source storage type")
			return
		}
	}

	res, err = s.secretStorageRepo.CreateSecretStorage(storage)

	if err != nil {
		return
	}

	key := db.AccessKey{
		Name:              random.String(10),
		Type:              db.AccessKeyString,
		ProjectID:         &storage.ProjectID,
		Owner:             db.AccessKeySecretStorage,
		StorageID:         &res.ID,
		SourceStorageType: sourceStorageType,
	}

	if sourceStorageKey != "" {
		key.SourceStorageKey = &sourceStorageKey
	} else {
		key.String = storage.Secret
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
		Owner:     db.AccessKeySecretStorage,
		StorageID: &storage.ID,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		if storage.Secret == "" {
			// empty vault token means the user didn't set a new token,
			// so we don't create a new access key.
			return
		}

		sourceStorageType := storage.SourceStorageType
		sourceStorageKey := ""

		if sourceStorageType != nil {
			switch *sourceStorageType {
			case db.AccessKeySourceStorageEnv, db.AccessKeySourceStorageFile:
				sourceStorageKey = storage.Secret
			default:
				err = errors.New("unsupported source storage type")
				return
			}
		}

		newKey := db.AccessKey{
			Name:              random.String(10),
			Type:              db.AccessKeyString,
			ProjectID:         &storage.ProjectID,
			Owner:             db.AccessKeySecretStorage,
			StorageID:         &storage.ID,
			SourceStorageType: sourceStorageType,
		}

		if sourceStorageKey != "" {
			newKey.SourceStorageKey = &sourceStorageKey
		} else {
			newKey.String = storage.Secret
		}

		_, err = s.accessKeyService.Create(newKey)

	} else {
		vault := keys[0]
		if storage.Secret == "" {
			// Do nothing if the vault token is empty,
			// as it means the user haven't set a new token.

			//err = s.keyRepo.DeleteAccessKey(storage.ProjectID, vault.ID)
			return
		}

		sourceStorageType := storage.SourceStorageType
		sourceStorageKey := ""

		if sourceStorageType != nil {
			switch *sourceStorageType {
			case db.AccessKeySourceStorageEnv, db.AccessKeySourceStorageFile:
				sourceStorageKey = storage.Secret
			default:
				err = errors.New("unsupported source storage type")
				return
			}
		}

		vault.OverrideSecret = true
		vault.SourceStorageType = sourceStorageType
		if sourceStorageKey != "" {
			vault.SourceStorageKey = &sourceStorageKey
			vault.String = ""
			// Clear previously persisted encrypted secret when switching to env/file source.
			vault.Secret = nil
		} else {
			vault.SourceStorageKey = nil
			vault.String = storage.Secret
		}

		err = s.accessKeyService.Update(vault)
	}

	return
}

func (s *SecretStorageServiceImpl) GetSecretStorages(projectID int) (storages []db.SecretStorage, err error) {
	return pro.GetSecretStorages(s.secretStorageRepo, projectID)
}
