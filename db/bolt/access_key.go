package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) GetAccessKey(projectID int, accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(projectID, db.AccessKeyProps, intObjectID(accessKeyID), &key)
	if err != nil {
		return
	}
	err = key.DeserializeSecret()
	return
}

func (d *BoltDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyProps, params, nil, &keys)
	return keys, err
}

func (d *BoltDb) UpdateAccessKey(key db.AccessKey) error {
	err := key.SerializeSecret()
	if err != nil {
		return err
	}
	return d.updateObject(*key.ProjectID, db.AccessKeyProps, key)
}

func (d *BoltDb) CreateAccessKey(key db.AccessKey) (db.AccessKey,  error) {
	err := key.SerializeSecret()
	if err != nil {
		return db.AccessKey{}, err
	}
	newKey, err := d.createObject(*key.ProjectID, db.AccessKeyProps, key)
	return newKey.(db.AccessKey), err
}

func (d *BoltDb) DeleteAccessKey(projectID int, accessKeyID int) error {
	return d.deleteObject(projectID, db.AccessKeyProps, intObjectID(accessKeyID))
}

func (d *BoltDb) DeleteAccessKeySoft(projectID int, accessKeyID int) error {
	return d.deleteObjectSoft(projectID, db.AccessKeyProps, intObjectID(accessKeyID))
}

func (d *BoltDb) GetGlobalAccessKey(accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(0, db.GlobalAccessKeyProps, intObjectID(accessKeyID), &key)
	if err != nil {
		return
	}
	err = key.DeserializeSecret()
	return
}

func (d *BoltDb) GetGlobalAccessKeys(params db.RetrieveQueryParams) (keys []db.AccessKey, err error) {
	err = d.getObjects(0, db.GlobalAccessKeyProps, params, nil, &keys)
	return
}

func (d *BoltDb) UpdateGlobalAccessKey(key db.AccessKey) error {
	err := key.SerializeSecret()
	if err != nil {
		return err
	}
	return d.updateObject(0, db.GlobalAccessKeyProps, key)
}

func (d *BoltDb) CreateGlobalAccessKey(key db.AccessKey) (db.AccessKey, error) {
	err := key.SerializeSecret()
	if err != nil {
		return db.AccessKey{}, err
	}
	newKey, err := d.createObject(0, db.GlobalAccessKeyProps, key)
	return newKey.(db.AccessKey), err
}

func (d *BoltDb) DeleteGlobalAccessKey(accessKeyID int) error {
	return d.deleteObject(0, db.GlobalAccessKeyProps, intObjectID(accessKeyID))
}

func (d *BoltDb) DeleteGlobalAccessKeySoft(accessKeyID int) error {
	return d.deleteObjectSoft(0, db.GlobalAccessKeyProps, intObjectID(accessKeyID))
}
