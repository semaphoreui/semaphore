package server_services

import "github.com/semaphoreui/semaphore/db"

type AccessKeyService interface {
	UpdateAccessKey(key db.AccessKey) error
	CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error)
	GetAccessKeys(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error)
}

type AccessKeyServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	encryptionService AccessKeyEncryptionService
}

func NewAccessKeyService(
	accessKeyRepo db.AccessKeyManager,
	encryptionService AccessKeyEncryptionService,
) AccessKeyService {
	return &AccessKeyServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		encryptionService: encryptionService,
	}
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
