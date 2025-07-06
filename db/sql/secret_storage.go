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
		"insert into project__secret_storage (name, type, project_id) values (?, ?, ?)",
		storage.Name,
		storage.Type,
		storage.ProjectID,
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

func (d *SqlDb) UpdateSecretStorage(storage db.SecretStorage) (db.SecretStorage, error) {
	//TODO implement me
	panic("implement me")
}
