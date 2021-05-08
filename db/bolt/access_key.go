package bolt

import "github.com/ansible-semaphore/semaphore/db"

func (d *BoltDb) GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error) {
	var key db.AccessKey
	err := d.getObject(projectID, db.AccessKeyObject, accessKeyID, &key)
	return key, err
}

func (d *BoltDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyObject, params, &keys)
	return keys, err
}

func (d *BoltDb) UpdateAccessKey(key db.AccessKey) error {
	return nil
}

func (d *BoltDb) CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	return
}

func (d *BoltDb) DeleteAccessKey(projectID int, accessKeyID int) error {
	return d.deleteObject(projectID, db.AccessKeyObject, accessKeyID)
}

func (d *BoltDb) DeleteAccessKeySoft(projectID int, accessKeyID int) error {
	return d.deleteObjectSoft(projectID, db.AccessKeyObject, accessKeyID)
}


func (d *BoltDb) GetGlobalAccessKey(accessKeyID int) (db.AccessKey, error) {
	var key db.AccessKey
	err := d.getObject(0, db.GlobalAccessKeyObject, accessKeyID, &key)
	return key, err
}

func (d *BoltDb) GetGlobalAccessKeys(params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(0, db.GlobalAccessKeyObject, params, &keys)
	return keys, err
}

func (d *BoltDb) UpdateGlobalAccessKey(key db.AccessKey) error {
	return nil
}

func (d *BoltDb) CreateGlobalAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	return
}

func (d *BoltDb) DeleteGlobalAccessKey(accessKeyID int) error {
	return d.deleteObject(0, db.GlobalAccessKeyObject, accessKeyID)
}

func (d *BoltDb) DeleteGlobalAccessKeySoft(accessKeyID int) error {
	return d.deleteObjectSoft(0, db.GlobalAccessKeyObject, accessKeyID)
}
