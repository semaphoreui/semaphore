package bolt

import "github.com/semaphoreui/semaphore/db"

func (d *BoltDb) GetSecretStorages(projectID int) ([]db.SecretStorage, error) {
	return []db.SecretStorage{}, nil
}

func (d *BoltDb) CreateSecretStorage(storage db.SecretStorage) (db.SecretStorage, error) {
	//TODO implement me
	panic("implement me")
}

func (d *BoltDb) GetSecretStorage(projectID int, storageID int) (db.SecretStorage, error) {
	//TODO implement me
	panic("implement me")
}

func (d *BoltDb) DeleteSecretStorage(projectID int, storageID int) error {
	panic("implement me")
}

func (d *BoltDb) UpdateSecretStorage(storage db.SecretStorage) error {
	//TODO implement me
	panic("implement me")
}

func (d *BoltDb) GetSecretStorageRefs(projectID int, storageID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.SecretStorageProps, storageID)
}
