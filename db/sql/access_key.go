package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) GetAccessKey(projectID int, accessKeyID int, deserializeSecret bool) (key db.AccessKey, err error) {
	err = d.getObject(projectID, db.AccessKeyProps, accessKeyID, &key)

	if err != nil {
		return
	}

	if deserializeSecret {
		err = key.DeserializeSecret()
	}

	return
}

func (d *SqlDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyProps, params, &keys)
	return keys, err
}

func (d *SqlDb) UpdateAccessKey(key db.AccessKey) error {
	err := key.Validate(key.OverrideSecret)

	if err != nil {
		return err
	}

	err = key.SerializeSecret()

	if err != nil {
		return err
	}

	var res sql.Result

	var args []interface{}
	query := "update access_key set name=?"
	args = append(args, key.Name)

	if key.OverrideSecret {
		query += ", type=?, secret=?"
		args = append(args, key.Type)
		args = append(args, key.Secret)
	}

	query += " where id=?"
	args = append(args, key.ID)

	query += " and project_id=?"
	args = append(args, key.ProjectID)

	res, err = d.exec(query, args...)

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