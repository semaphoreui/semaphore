package server

import "github.com/semaphoreui/semaphore/db"

type AccessKeyService interface {
	UpdateAccessKey(key db.AccessKey) error
	CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error)
	GetAccessKeys(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error)
	DeleteAccessKey(projectID int, keyID int) (err error)
}

type AccessKeyServiceImpl struct {
	accessKeyRepo        db.AccessKeyManager
	encryptionService    AccessKeyEncryptionService
	secretStorageService SecretStorageService
}

func NewAccessKeyService(
	accessKeyRepo db.AccessKeyManager,
	storageService SecretStorageService,
	encryptionService AccessKeyEncryptionService,
) AccessKeyService {
	return &AccessKeyServiceImpl{
		accessKeyRepo:        accessKeyRepo,
		encryptionService:    encryptionService,
		secretStorageService: storageService,
	}
}

func (s *AccessKeyServiceImpl) DeleteAccessKey(projectID int, keyID int) (err error) {
	key, err := s.accessKeyRepo.GetAccessKey(projectID, keyID)
	if err != nil {
		return
	}

	if key.SourceStorageID != nil {
		var storage db.SecretStorage
		storage, err = s.secretStorageService.GetSecretStorage(projectID, *key.SourceStorageID)
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

func (s *AccessKeyServiceImpl) GetAccessKeys(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return s.accessKeyRepo.GetAccessKeys(projectID, options, params)
}

func (s *AccessKeyServiceImpl) CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	err = s.encryptionService.SerializeSecret(&key)
	if err != nil {
		return
	}

	newKey, err = s.accessKeyRepo.CreateAccessKey(key)
	return
}

func (s *AccessKeyServiceImpl) UpdateAccessKey(key db.AccessKey) (err error) {
	if key.OverrideSecret {
		err = s.encryptionService.SerializeSecret(&key)
		if err != nil {
			return
		}
	}

	err = s.accessKeyRepo.UpdateAccessKey(key)
	return
}
