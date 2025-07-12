//go:build !pro

package server

import "github.com/semaphoreui/semaphore/db"

func (s *SecretStorageServiceImpl) GetSecretStorages(projectID int) (storages []db.SecretStorage, err error) {
	storages = []db.SecretStorage{}
	return
}
