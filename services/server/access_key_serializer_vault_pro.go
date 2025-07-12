//go:build pro

package server

import (
	"github.com/semaphoreui/semaphore/db"
	pro "github.com/semaphoreui/semaphore/pro/services/server"
)

func NewVaultAccessKeyDeserializer(
	accessKeyRepo db.AccessKeyManager,
	secretStorageRepo db.SecretStorageRepository,
	encryptionService AccessKeyEncryptionService,
) *pro.VaultAccessKeyDeserializer {
	return pro.NewVaultAccessKeyDeserializer(
		accessKeyRepo,
		secretStorageRepo,
		encryptionService,
	)
}
