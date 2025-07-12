//go:build pro

package server_services

import (
	"github.com/semaphoreui/semaphore/db"
	server_services_pro "github.com/semaphoreui/semaphore/pro/services/server_services"
)

func NewVaultAccessKeyDeserializer(
	accessKeyRepo db.AccessKeyManager,
	secretStorageRepo db.SecretStorageRepository,
	encryptionService AccessKeyEncryptionService,
) *server_services_pro.VaultAccessKeyDeserializer {
	return server_services_pro.NewVaultAccessKeyDeserializer(
		accessKeyRepo,
		secretStorageRepo,
		encryptionService,
	)
}
