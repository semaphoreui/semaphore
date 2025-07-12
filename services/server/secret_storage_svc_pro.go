//go:build pro

package server

import (
	"github.com/semaphoreui/semaphore/db"
	pro "github.com/semaphoreui/semaphore/pro/services/server"
)

func (s *SecretStorageServiceImpl) GetSecretStorages(projectID int) (storages []db.SecretStorage, err error) {
	return pro.GetSecretStorages(s.secretStorageRepo, projectID)
}
