package server

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

func GetSecretStorages(repo db.SecretStorageRepository, projectID int) (storages []db.SecretStorage, err error) {
	storages = make([]db.SecretStorage, 0)
	return
}

func SyncSecrets(
	storage db.SecretStorage,
	accessKeyRepo db.AccessKeyManager,
	decryptor DvlsStorageTokenDeserializer,
) error {
	switch storage.Type {
	case db.SecretStorageTypeDvls:
		return nil
	case db.SecretStorageTypeAwsSm:
		return nil
	case db.SecretStorageTypeAzureKv:
		return nil
	default:
		return fmt.Errorf("sync is not supported for storage type %q", storage.Type)
	}
}
