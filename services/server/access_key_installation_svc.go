package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

type AccessKeyInstallationService interface {
	Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error)
}

func NewAccessKeyInstallationService(encryptionService AccessKeyEncryptionService) AccessKeyInstallationService {
	return &AccessKeyInstallationServiceImpl{
		encryptionService: encryptionService,
	}
}

type AccessKeyInstallationServiceImpl struct {
	encryptionService AccessKeyEncryptionService
}

func (s *AccessKeyInstallationServiceImpl) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error) {

	if key.Type == db.AccessKeyNone {
		return
	}

	err = s.encryptionService.DeserializeSecret(&key)

	if err != nil {
		return
	}

	installation, err = ssh.KeyInstaller{}.Install(key, usage, logger)

	return
}
