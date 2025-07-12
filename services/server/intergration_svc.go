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

func (s *IntegrationServiceImpl) FillIntegration(inventory *db.Integration) (err error) {
	if inventory.AuthSecretID != nil {
		inventory.AuthSecret, err = s.accessKeyRepo.GetAccessKey(inventory.ProjectID, *inventory.AuthSecretID)
	}

	if err != nil {
		return
	}

	err = s.encryptionService.DeserializeSecret(&inventory.AuthSecret)

	return
}
