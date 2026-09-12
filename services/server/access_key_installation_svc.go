package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

type AccessKeyInstallationService interface {
	Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error)
	InstallAll(keys []db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error)
}

func NewAccessKeyInstallationService(encryptionService AccessKeyEncryptionService) AccessKeyInstallationService {
	return &AccessKeyInstallationServiceImpl{
		encryptionService: encryptionService,
	}
}

type AccessKeyInstallationServiceImpl struct {
	encryptionService AccessKeyEncryptionService
}

// InstallAll installs several keys into a single agent, decrypting each of them
// first, as Install does for a single key.
func (s *AccessKeyInstallationServiceImpl) InstallAll(keys []db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error) {
	decrypted := make([]db.AccessKey, 0, len(keys))

	for _, key := range keys {
		if key.Type == db.AccessKeyNone {
			continue
		}

		if err = s.encryptionService.DeserializeSecret(&key); err != nil {
			return
		}

		decrypted = append(decrypted, key)
	}

	if len(decrypted) == 0 {
		return
	}

	return ssh.KeyInstaller{}.InstallAll(decrypted, usage, logger)
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
