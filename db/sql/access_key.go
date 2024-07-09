package sql

import (
	"database/sql"
	"errors"
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) GetAccessKey(projectID int, accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(projectID, db.AccessKeyProps, accessKeyID, &key)
	return
}

func (d *SqlDb) GetAccessKeyRefs(projectID int, keyID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.AccessKeyProps, keyID)
}

func (d *SqlDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) (keys []db.AccessKey, err error) {
	keys = make([]db.AccessKey, 0)

	q := d.makeObjectsQuery(projectID, db.AccessKeyProps, params).Where("pe.environment_id IS NULL")

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&keys, query, args...)

	return
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
		"insert into access_key (name, type, project_id, secret, environment_id) values (?, ?, ?, ?, ?)",
		key.Name,
		key.Type,
		key.ProjectID,
		key.Secret,
		key.EnvironmentID)

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

const RekeyBatchSize = 100

func (d *SqlDb) RekeyAccessKeys(oldKey string) (err error) {

	var globalProps = db.AccessKeyProps
	globalProps.IsGlobal = true

	for i := 0; ; i++ {

		var keys []db.AccessKey
		err = d.getObjects(-1, globalProps, db.RetrieveQueryParams{Count: RekeyBatchSize, Offset: i * RekeyBatchSize}, nil, &keys)

		if err != nil {
			return
		}

		if len(keys) == 0 {
			break
		}

		for _, key := range keys {

			err = key.DeserializeSecret2(oldKey)

			if err != nil {
				return err
			}

			key.OverrideSecret = true
			err = d.UpdateAccessKey(key)

			if err != nil && !errors.Is(err, db.ErrNotFound) {
				return err
			}
		}
	}

	return
}
