package server

import "github.com/Digital-Data-Co/semaphore/db"

func GetSecretStorages(repo db.SecretStorageRepository, projectID int) (storages []db.SecretStorage, err error) {
	storages = make([]db.SecretStorage, 0)
	return
}
