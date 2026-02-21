package server

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
)

type AccessKeyService interface {
	Update(key db.AccessKey) error
	Create(key db.AccessKey) (newKey db.AccessKey, err error)
	GetAll(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error)
	Delete(projectID int, keyID int) (err error)
}

type AccessKeyServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	encryptionService AccessKeyEncryptionService
	secretStorageRepo db.SecretStorageRepository
}

func NewAccessKeyService(
	accessKeyRepo db.AccessKeyManager,
	encryptionService AccessKeyEncryptionService,
	secretStorageRepo db.SecretStorageRepository,
) AccessKeyService {
	return &AccessKeyServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		encryptionService: encryptionService,
		secretStorageRepo: secretStorageRepo,
	}
}

func (s *AccessKeyServiceImpl) Delete(projectID int, keyID int) (err error) {
	key, err := s.accessKeyRepo.GetAccessKey(projectID, keyID)
	if err != nil {
		return
	}

	if key.SourceStorageID != nil {
		var storage db.SecretStorage
		storage, err = s.secretStorageRepo.GetSecretStorage(projectID, *key.SourceStorageID)
		if err != nil {
			return
		}

		if !storage.ReadOnly {
			err = s.encryptionService.DeleteSecret(&key)
			if err != nil {
				return
			}
		}
	}

	err = s.accessKeyRepo.DeleteAccessKey(projectID, keyID)

	return
}

func (s *AccessKeyServiceImpl) GetAll(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return s.accessKeyRepo.GetAccessKeys(projectID, options, params)
}

func (s *AccessKeyServiceImpl) Create(key db.AccessKey) (newKey db.AccessKey, err error) {

	err = s.encryptionService.SerializeSecret(&key)
	if err != nil && !errors.Is(err, ErrReadOnlyStorage) {
		return
	}

	newKey, err = s.accessKeyRepo.CreateAccessKey(key)
	return
}

func (s *AccessKeyServiceImpl) Update(key db.AccessKey) (err error) {
	if !key.OverrideSecret {
		err = s.accessKeyRepo.UpdateAccessKey(key)
		return
	}

	var oldKey db.AccessKey
	oldKey, err = s.accessKeyRepo.GetAccessKey(*key.ProjectID, key.ID)
	if err != nil {
		return
	}

	if oldKey.SourceStorageType != nil && !oldKey.IsNativelyReadOnly() {
		// validate if it is secure to override secret storage

		var oldSt db.SecretStorage
		oldSt, err = s.secretStorageRepo.GetSecretStorage(*key.ProjectID, *oldKey.SourceStorageID)
		if err != nil {
			return
		}

		if !oldSt.ReadOnly && *oldKey.SourceStorageID != *key.SourceStorageID {
			err = errors.New("cannot override secret storage")
			return
		}
	}

	if !key.IsNativelyReadOnly() {
		err = s.encryptionService.SerializeSecret(&key)
		if err != nil {
			return
		}
	}

	err = s.accessKeyRepo.UpdateAccessKey(key)

	return
}
