package sql

import "github.com/semaphoreui/semaphore/db"

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
		storages[i].SyncPaths, err = d.GetSecretStorageSyncPaths(storages[i].ID)
		if err != nil {
			return
		}
	}

	return
}

func (d *SqlDb) CreateSecretStorage(storage db.SecretStorage) (newStorage db.SecretStorage, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__secret_storage (name, type, project_id, params, readonly, sync_enabled) values (?, ?, ?, ?, ?, ?)",
		storage.Name,
		storage.Type,
		storage.ProjectID,
		storage.Params,
		storage.ReadOnly,
		storage.SyncEnabled,
	)

	if err != nil {
		return
	}

	newStorage = storage
	newStorage.ID = insertID

	err = d.ReplaceSecretStorageSyncPaths(newStorage.ID, storage.SyncPaths)
	if err != nil {
		return
	}

	newStorage.SyncPaths, err = d.GetSecretStorageSyncPaths(newStorage.ID)
	return
}

func (d *SqlDb) GetSecretStorage(projectID int, storageID int) (key db.SecretStorage, err error) {

	err = d.getObject(projectID, db.SecretStorageProps, storageID, &key)
	if err != nil {
		return
	}

	key.SyncPaths, err = d.GetSecretStorageSyncPaths(key.ID)
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
		"readonly=?, "+
		"sync_enabled=? "+
		"where project_id=? and id=?",
		storage.Name,
		storage.Type,
		storage.Params,
		storage.ReadOnly,
		storage.SyncEnabled,
		storage.ProjectID,
		storage.ID)

	if err != nil {
		return err
	}

	return d.ReplaceSecretStorageSyncPaths(storage.ID, storage.SyncPaths)
}

func (d *SqlDb) GetSecretStorageSyncPaths(storageID int) (paths []db.SecretStorageSyncPath, err error) {
	paths = make([]db.SecretStorageSyncPath, 0)
	_, err = d.selectAll(
		&paths,
		"select id, storage_id, path, prefix, `separator` "+
			"from project__secret_storage__sync_path where storage_id=? order by id",
		storageID,
	)
	return
}

func (d *SqlDb) ReplaceSecretStorageSyncPaths(storageID int, paths []db.SecretStorageSyncPath) error {
	if _, err := d.exec("delete from project__secret_storage__sync_path where storage_id=?", storageID); err != nil {
		return err
	}

	for _, p := range paths {
		if _, err := d.insert(
			"id",
			"insert into project__secret_storage__sync_path (storage_id, path, prefix, `separator`) values (?, ?, ?, ?)",
			storageID,
			p.Path,
			p.Prefix,
			p.Separator,
		); err != nil {
			return err
		}
	}

	return nil
}
