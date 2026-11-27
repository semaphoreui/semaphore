package server

import "github.com/semaphoreui/semaphore/db"

type IntegrationService interface {
	FillIntegration(integration *db.Integration) error
}

type IntegrationServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	encryptionService AccessKeyEncryptionService
}

func NewIntegrationService(
	accessKeyRepo db.AccessKeyManager,
	encryptionService AccessKeyEncryptionService,
) IntegrationService {
	return &IntegrationServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		encryptionService: encryptionService,
	}
}

func (s *IntegrationServiceImpl) FillIntegration(integration *db.Integration) (err error) {
	if integration.AuthSecretID != nil {
		integration.AuthSecret, err = s.accessKeyRepo.GetAccessKey(integration.ProjectID, *integration.AuthSecretID)
	}

	if err != nil {
		return
	}

	err = s.encryptionService.DeserializeSecret(&integration.AuthSecret)

	return
}
