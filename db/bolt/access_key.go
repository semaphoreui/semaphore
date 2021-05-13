package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error) {
	var key db.AccessKey
	err := d.getObject(projectID, db.AccessKeyObject, intObjectID(accessKeyID), &key)
	return key, err
}

func (d *BoltDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyObject, params, nil, &keys)
	return keys, err
}

func (d *BoltDb) UpdateAccessKey(key db.AccessKey) error {
	return d.updateObject(*key.ProjectID, db.AccessKeyObject, key)
}

func (d *BoltDb) CreateAccessKey(key db.AccessKey) (db.AccessKey,  error) {
	newKey, err := d.createObject(*key.ProjectID, db.GlobalAccessKeyObject, key)
	return newKey.(db.AccessKey), err
}

func (d *BoltDb) DeleteAccessKey(projectID int, accessKeyID int) error {
	return d.deleteObject(projectID, db.AccessKeyObject, intObjectID(accessKeyID))
}

func (d *BoltDb) DeleteAccessKeySoft(projectID int, accessKeyID int) error {
	return d.deleteObjectSoft(projectID, db.AccessKeyObject, intObjectID(accessKeyID))
}

func (d *BoltDb) GetGlobalAccessKey(accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(0, db.GlobalAccessKeyObject, intObjectID(accessKeyID), &key)
	return
}

func (d *BoltDb) GetGlobalAccessKeys(params db.RetrieveQueryParams) (keys []db.AccessKey, err error) {
	err = d.getObjects(0, db.GlobalAccessKeyObject, params, nil, &keys)
	return
}

func (d *BoltDb) UpdateGlobalAccessKey(key db.AccessKey) error {
	return d.updateObject(0, db.AccessKeyObject, key)
}

func (d *BoltDb) CreateGlobalAccessKey(key db.AccessKey) (db.AccessKey, error) {
	newKey, err := d.createObject(0, db.GlobalAccessKeyObject, key)
	return newKey.(db.AccessKey), err
}

func (d *BoltDb) DeleteGlobalAccessKey(accessKeyID int) error {
	return d.deleteObject(0, db.GlobalAccessKeyObject, intObjectID(accessKeyID))
}

func (d *BoltDb) DeleteGlobalAccessKeySoft(accessKeyID int) error {
	return d.deleteObjectSoft(0, db.GlobalAccessKeyObject, intObjectID(accessKeyID))
}
