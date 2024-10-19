package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetAccessKey(projectID int, accessKeyID int) (key db.AccessKey, err error) {
	err = d.getObject(projectID, db.AccessKeyProps, intObjectID(accessKeyID), &key)
	if err != nil {
		return
	}

	return
}

func (d *BoltDb) GetAccessKeyRefs(projectID int, accessKeyID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.AccessKeyProps, accessKeyID)
}

func (d *BoltDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyProps, params, func(i interface{}) bool {
		k := i.(db.AccessKey)
		return k.EnvironmentID == nil
	}, &keys)
	return keys, err
}

func (d *BoltDb) UpdateAccessKey(key db.AccessKey) error {
	err := key.Validate(key.OverrideSecret)

	if err != nil {
		return err
	}

	if key.OverrideSecret {
		err = key.SerializeSecret()
		if err != nil {
			return err
		}
	} else { // accept only new name, ignore other changes
		oldKey, err2 := d.GetAccessKey(*key.ProjectID, key.ID)
		if err2 != nil {
			return err2
		}
		oldKey.Name = key.Name
		key = oldKey
	}

	return d.updateObject(*key.ProjectID, db.AccessKeyProps, key)
}

func (d *BoltDb) CreateAccessKey(key db.AccessKey) (db.AccessKey, error) {
	err := key.SerializeSecret()
	if err != nil {
		return db.AccessKey{}, err
	}
	newKey, err := d.createObject(*key.ProjectID, db.AccessKeyProps, key)
	return newKey.(db.AccessKey), err
}

func (d *BoltDb) DeleteAccessKey(projectID int, accessKeyID int) error {
	return d.deleteObject(projectID, db.AccessKeyProps, intObjectID(accessKeyID), nil)
}

func (d *BoltDb) RekeyAccessKeys(oldKey string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		var allProjects []db.Project

		err := d.getObjectsTx(tx, 0, db.ProjectProps, db.RetrieveQueryParams{}, nil, &allProjects)

		if err != nil {
			return err
		}

		for _, project := range allProjects {
			var keys []db.AccessKey
			err = d.getObjectsTx(tx, project.ID, db.AccessKeyProps, db.RetrieveQueryParams{}, nil, &keys)
			if err != nil {
				return err
			}

			for _, key := range keys {
				err = key.DeserializeSecret2(oldKey)

				if err != nil {
					return err
				}

				err = key.SerializeSecret()
				if err != nil {
					return err
				}

				err = d.updateObjectTx(tx, *key.ProjectID, db.AccessKeyProps, key)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
