package server

import (
	"github.com/Digital-Data-Co/semaphore/db"
	"github.com/Digital-Data-Co/semaphore/pkg/ssh"
	"github.com/Digital-Data-Co/semaphore/pkg/task_logger"
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
