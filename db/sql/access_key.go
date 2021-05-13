package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error) {
	var key db.AccessKey
	err := d.getObject(projectID, db.AccessKeyProps, accessKeyID, &key)
	return key, err
}

func (d *SqlDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyProps, params, &keys)
	return keys, err
}

func (d *SqlDb) UpdateAccessKey(key db.AccessKey) error {
	res, err := d.sql.Exec(
		"update access_key set name=?, type=?, `key`=?, secret=? where project_id=? and id=?",
		key.Name,
		key.Type,
		key.Key,
		key.Secret,
		key.ProjectID,
		key.ID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	res, err := d.sql.Exec(
		"insert into access_key (name, type, project_id, `key`, secret) values (?, ?, ?, ?, ?)",
		key.Name,
		key.Type,
		key.ProjectID,
		key.Key,
		key.Secret)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newKey = key
	newKey.ID = int(insertID)
	return
}

func (d *SqlDb) DeleteAccessKey(projectID int, accessKeyID int) error {
	return d.deleteObject(projectID, db.AccessKeyProps, accessKeyID)
}

func (d *SqlDb) DeleteAccessKeySoft(projectID int, accessKeyID int) error {
	return d.deleteObjectSoft(projectID, db.AccessKeyProps, accessKeyID)
}


func (d *SqlDb) GetGlobalAccessKey(accessKeyID int) (db.AccessKey, error) {
	var key db.AccessKey
	err := d.getObject(0, db.GlobalAccessKeyProps, accessKeyID, &key)
	return key, err
}

func (d *SqlDb) GetGlobalAccessKeys(params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(0, db.GlobalAccessKeyProps, params, &keys)
	return keys, err
}

func (d *SqlDb) UpdateGlobalAccessKey(key db.AccessKey) error {
	res, err := d.sql.Exec(
		"update access_key set name=?, type=?, `key`=?, secret=? where id=?",
		key.Name,
		key.Type,
		key.Key,
		key.Secret,
		key.ID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateGlobalAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	res, err := d.sql.Exec(
		"insert into access_key (name, type, `key`, secret) values (?, ?, ?, ?)",
		key.Name,
		key.Type,
		key.Key,
		key.Secret)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newKey = key
	newKey.ID = int(insertID)
	return
}

func (d *SqlDb) DeleteGlobalAccessKey(accessKeyID int) error {
	return d.deleteObject(0, db.GlobalAccessKeyProps, accessKeyID)
}

func (d *SqlDb) DeleteGlobalAccessKeySoft(accessKeyID int) error {
	return d.deleteObjectSoft(0, db.GlobalAccessKeyProps, accessKeyID)
}
