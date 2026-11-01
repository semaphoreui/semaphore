package server

import "github.com/semaphoreui/semaphore/db"

// IntegrationService loads what the webhook receiver needs beyond the rows
// themselves: the decrypted key an endpoint or an integration group verifies
// requests with.
type IntegrationService interface {
	// FillAuthSecret loads auth.AuthSecret from the project's keys by
	// auth.AuthSecretID and decrypts it. Without an AuthSecretID the secret is
	// left empty, which VerificationSecret then refuses for the methods that
	// need one.
	FillAuthSecret(projectID int, auth *db.IntegrationAuth) error
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

func (s *IntegrationServiceImpl) FillAuthSecret(projectID int, auth *db.IntegrationAuth) (err error) {
	if auth.AuthSecretID == nil {
		return nil
	}

	// Scoped by project: a key id of another project is not found.
	auth.AuthSecret, err = s.accessKeyRepo.GetAccessKey(projectID, *auth.AuthSecretID)
	if err != nil {
		return
	}

	return s.encryptionService.DeserializeSecret(&auth.AuthSecret)
}
