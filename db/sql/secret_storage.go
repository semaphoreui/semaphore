package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetSecretStorages(projectID int) (storages []db.SecretStorage, err error) {
	storages = make([]db.SecretStorage, 0)

	q, err := d.makeObjectsQuery(projectID, db.SecretStorageProps, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&storages, query, args...)

	if err != nil {
		return
	}

	for i := range storages {
		if err = d.fillStorageSync(&storages[i]); err != nil {
			return
		}
	}

	return
}

func (d *SqlDb) CreateSecretStorage(storage db.SecretStorage) (newStorage db.SecretStorage, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__secret_storage (name, type, project_id, params, readonly) values (?, ?, ?, ?, ?)",
		storage.Name,
		storage.Type,
		storage.ProjectID,
		storage.Params,
		storage.ReadOnly,
	)

	if err != nil {
		return
	}

	newStorage = storage
	newStorage.ID = insertID

	if err = d.SaveSecretSync(secretSyncFromStorage(newStorage)); err != nil {
		return
	}

	err = d.fillStorageSync(&newStorage)
	return
}

func (d *SqlDb) GetSecretStorage(projectID int, storageID int) (storage db.SecretStorage, err error) {

	err = d.getObject(projectID, db.SecretStorageProps, storageID, &storage)
	if err != nil {
		return
	}

	err = d.fillStorageSync(&storage)
	return
}

func (d *SqlDb) DeleteSecretStorage(projectID int, storageID int) error {
	return d.deleteObject(projectID, db.SecretStorageProps, storageID)
}

func (d *SqlDb) GetSecretStorageRefs(projectID int, storageID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.SecretStorageProps, storageID)
}

func (d *SqlDb) UpdateSecretStorage(storage db.SecretStorage) error {
	_, err := d.exec("update project__secret_storage set "+
		"name=?, "+
		"type=?, "+
		"params=?, "+
		"readonly=? "+
		"where project_id=? and id=?",
		storage.Name,
		storage.Type,
		storage.Params,
		storage.ReadOnly,
		storage.ProjectID,
		storage.ID)

	if err != nil {
		return err
	}

	return d.SaveSecretSync(secretSyncFromStorage(storage))
}

// secretSyncFromStorage projects a SecretStorage's transfer-only sync fields
// onto a SecretSync payload for persistence.
func secretSyncFromStorage(storage db.SecretStorage) db.SecretSync {
	return db.SecretSync{
		ProjectID:        storage.ProjectID,
		StorageID:        storage.ID,
		SyncEnabled:      storage.SyncEnabled,
		SyncInterval:     storage.SyncInterval,
		LastSyncedAt:     storage.LastSyncedAt,
		LastSyncFailedAt: storage.LastSyncFailedAt,
		Paths:            storage.SyncPaths,
	}
}

func (d *SqlDb) fillStorageSync(storage *db.SecretStorage) error {
	sync, err := d.GetStorageSecretSync(storage.ID)
	if err == db.ErrNotFound {
		storage.SyncEnabled = false
		storage.SyncInterval = 0
		storage.LastSyncedAt = nil
		storage.LastSyncFailedAt = nil
		storage.SyncPaths = []db.SecretSyncPath{}
		return nil
	}
	if err != nil {
		return err
	}
	storage.SyncEnabled = sync.SyncEnabled
	storage.SyncInterval = sync.SyncInterval
	storage.LastSyncedAt = sync.LastSyncedAt
	storage.LastSyncFailedAt = sync.LastSyncFailedAt
	storage.SyncPaths = sync.Paths
	if storage.SyncPaths == nil {
		storage.SyncPaths = []db.SecretSyncPath{}
	}
	return nil
}
