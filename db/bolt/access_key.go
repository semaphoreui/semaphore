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

func (d *BoltDb) updateAccessKey(key db.AccessKey, isGlobal bool) error {
	err := key.Validate(key.OverrideSecret)

	if err != nil {
		return err
	}

	var projectId int
	if isGlobal {
		projectId = 0
	} else {
		projectId = *key.ProjectID
	}

	if key.OverrideSecret {
		err = key.SerializeSecret()
		if err != nil {
			return err
		}
	} else { // accept only new name, ignore other changes
		oldKey, err2 := d.GetAccessKey(projectId, key.ID)
		if err2 != nil {
			return err2
		}
		oldKey.Name = key.Name
		key = oldKey
	}

	return d.updateObject(projectId, db.AccessKeyProps, key)
}

func (d *BoltDb) UpdateAccessKey(key db.AccessKey) error {
	return d.updateAccessKey(key, false)
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
	return d.updateAccessKey(key, true)
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
