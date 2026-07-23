package sql

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) GetAccessKey(projectID int, accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(projectID, db.AccessKeyProps, accessKeyID, &key)
	return
}

func (d *SqlDb) GetAccessKeyRefs(projectID int, keyID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.AccessKeyProps, keyID)
}

func (d *SqlDb) GetAccessKeys(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) (keys []db.AccessKey, err error) {
	keys = make([]db.AccessKey, 0)

	q, err := d.makeObjectsQuery(projectID, db.AccessKeyProps, params)

	if err != nil {
		return
	}

	if err = options.Validate(); err != nil {
		return
	}

	if !options.IgnoreOwner {
		q = q.Where(squirrel.Eq{"pe.owner": options.Owner})
	}

	for _, f := range []struct {
		column string
		value  *int
	}{
		{"pe.environment_id", options.EnvironmentID},
		{"pe.storage_id", options.StorageID},
		{"pe.task_id", options.TaskID},
		{"pe.source_storage_id", options.SourceStorageID},
	} {
		if f.value != nil {
			q = q.Where(squirrel.Eq{f.column: *f.value})
		}
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&keys, query, args...)

	for i := range keys {
		keys[i].Empty = keys[i].IsEmpty()
	}

	return
}

func (d *SqlDb) UpdateAccessKey(key db.AccessKey) error {
	err := key.Validate(key.OverrideSecret)

	if err != nil {
		return err
	}

	var res sql.Result

	var args []any
	query := "update access_key set name=?"
	args = append(args, key.Name)

	if !key.IgnorePlain {
		query += ", plain=?"
		args = append(args, key.Plain)
	}

	if key.OverrideSecret {

		query += ", type=?, secret=?, source_storage_id=?, source_storage_key=?, source_storage_type=?"
		args = append(args, key.Type)
		args = append(args, key.Secret)
		args = append(args, key.SourceStorageID)
		args = append(args, key.SourceStorageKey)
		args = append(args, key.SourceStorageType)
	}

	query += " where id=?"
	args = append(args, key.ID)

	query += " and project_id=?"
	args = append(args, key.ProjectID)

	res, err = d.exec(query, args...)

	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {

	var insertID int

	if key.IgnorePlain {
		insertID, err = d.insert(
			"id",
			"insert into access_key ("+
				"name, "+
				"type, "+
				"project_id, "+
				"secret, "+
				"environment_id, "+
				"owner, "+
				"storage_id, "+
				"source_storage_id, "+
				"source_storage_key, "+
				"source_storage_type, "+
				"synchronized, "+
				"task_id, "+
				"expire_at) "+
				"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			key.Name,
			key.Type,
			key.ProjectID,
			key.Secret,
			key.EnvironmentID,
			key.Owner,
			key.StorageID,
			key.SourceStorageID,
			key.SourceStorageKey,
			key.SourceStorageType,
			key.Synchronized,
			key.TaskID,
			key.ExpireAt,
		)
	} else {
		insertID, err = d.insert(
			"id",
			"insert into access_key ("+
				"name, "+
				"type, "+
				"project_id, "+
				"secret, "+
				"plain, "+
				"environment_id, "+
				"owner, "+
				"storage_id, "+
				"source_storage_id, "+
				"source_storage_key, "+
				"source_storage_type, "+
				"synchronized, "+
				"task_id, "+
				"expire_at) "+
				"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			key.Name,
			key.Type,
			key.ProjectID,
			key.Secret,
			key.Plain,
			key.EnvironmentID,
			key.Owner,
			key.StorageID,
			key.SourceStorageID,
			key.SourceStorageKey,
			key.SourceStorageType,
			key.Synchronized,
			key.TaskID,
			key.ExpireAt,
		)

	}

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

func (d *SqlDb) GetTaskAccessKey(projectID int, taskID int) (key db.AccessKey, err error) {
	err = d.selectOne(
		&key,
		"select * from access_key where project_id=? and owner=? and task_id=?",
		projectID,
		db.AccessKeyTaskSecret,
		taskID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) DeleteTaskAccessKeys(projectID int, taskID int) error {
	_, err := d.exec(
		"delete from access_key where project_id=? and owner=? and task_id=?",
		projectID,
		db.AccessKeyTaskSecret,
		taskID)
	return err
}

func (d *SqlDb) DeleteExpiredTaskAccessKeys() error {
	// Do not delete keys for tasks that are still queued or running: an active
	// task with an expired key must fail dispatch with ErrAccessKeyExpired, not
	// silently run with empty survey variables after the row disappears.
	_, err := d.exec(
		`delete from access_key ak
		 where ak.owner=?
		   and ak.expire_at is not null
		   and ak.expire_at < ?
		   and (
		     ak.task_id is null
		     or exists (
		       select 1 from task t
		       where t.id = ak.task_id
		         and t.project_id = ak.project_id
		         and t.status in ('stopped', 'success', 'error')
		     )
		   )`,
		db.AccessKeyTaskSecret,
		tz.Now())
	return err
}
