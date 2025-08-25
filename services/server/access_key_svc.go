package server

import "github.com/semaphoreui/semaphore/db"

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

	err = key.Validate(true)
	if err != nil {
		return
	}

	err = s.encryptionService.SerializeSecret(&key)
	if err != nil {
		return
	}

	newKey, err = s.accessKeyRepo.CreateAccessKey(key)
	return
}

func (s *AccessKeyServiceImpl) Update(key db.AccessKey) (err error) {
	if key.OverrideSecret {
		err = s.encryptionService.SerializeSecret(&key)
		if err != nil {
			return
		}
	}

	err = s.accessKeyRepo.UpdateAccessKey(key)
	return
}
