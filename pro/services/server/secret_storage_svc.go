package server

import "github.com/semaphoreui/semaphore/db"

func GetSecretStorages(repo db.SecretStorageRepository, projectID int) (storages []db.SecretStorage, err error) {
	storages = make([]db.SecretStorage, 0)
	return
}

func SyncDvlsSecrets(
	storage db.SecretStorage,
	accessKeyRepo db.AccessKeyManager,
	decryptor DvlsStorageTokenDeserializer,
) error {
	return nil
}
