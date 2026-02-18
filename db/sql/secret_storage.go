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
	return
}

func (d *SqlDb) GetSecretStorage(projectID int, storageID int) (key db.SecretStorage, err error) {

	err = d.getObject(projectID, db.SecretStorageProps, storageID, &key)

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
	return err
}
