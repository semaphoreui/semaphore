package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetAccessKey(projectID int, accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(projectID, db.AccessKeyProps, accessKeyID, &key)

	if err != nil {
		return
	}

	err = key.DeserializeSecret()

	return
}

func (d *SqlDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyProps, params, &keys)
	return keys, err
}

func (d *SqlDb) UpdateAccessKey(key db.AccessKey) error {
	err := key.SerializeSecret()
	if err != nil {
		return err
	}

	res, err := d.exec(
		"update access_key set name=?, type=?, secret=? where project_id=? and id=?",
		key.Name,
		key.Type,
		key.Secret,
		key.ProjectID,
		key.ID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	err = key.SerializeSecret()
	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into access_key (name, type, project_id, secret) values (?, ?, ?, ?)",
		key.Name,
		key.Type,
		key.ProjectID,
		key.Secret)

	if err != nil {
		return
	}

	newKey = key
	newKey.ID = insertID
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
	err := key.SerializeSecret()
	if err != nil {
		return err
	}

	res, err := d.exec(
		"update access_key set name=?, type=?, secret=? where id=?",
		key.Name,
		key.Type,
		key.Secret,
		key.ID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateGlobalAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	err = key.SerializeSecret()
	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into access_key (name, type, secret) values (?, ?, ?)",
		key.Name,
		key.Type,
		key.Secret)

	if err != nil {
		return
	}

	newKey = key
	newKey.ID = insertID
	return
}

func (d *SqlDb) DeleteGlobalAccessKey(accessKeyID int) error {
	return d.deleteObject(0, db.GlobalAccessKeyProps, accessKeyID)
}

func (d *SqlDb) DeleteGlobalAccessKeySoft(accessKeyID int) error {
	return d.deleteObjectSoft(0, db.GlobalAccessKeyProps, accessKeyID)
}
