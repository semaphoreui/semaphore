package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetApps() (apps []db.App, err error) {
	err = d.getObjects(0, db.AppProps, db.RetrieveQueryParams{}, nil, &apps)
	return
}

func (d *BoltDb) GetApp(appID string) (app db.App, err error) {
	err = d.getObject(0, db.AppProps, strObjectID(appID), &app)
	return
}

func (d *BoltDb) CreateApp(app db.App) (db.App, error) {
	_, err := d.createObject(0, db.AppProps, app)
	return app, err
}

func (d *BoltDb) UpdateApp(app db.App) error {
	return d.updateObject(0, db.AppProps, app)
}

func (d *BoltDb) DeleteApp(appID string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteObject(0, db.AppProps, strObjectID(appID), tx)
	})
}

func (d *BoltDb) GetAppVersions(appID string) (versions []db.AppVersion, err error) {
	err = d.getObjects(0, db.AppVersionProps, db.RetrieveQueryParams{}, func(i any) bool {
		v := i.(db.AppVersion)
		return v.AppID == appID
	}, &versions)
	return
}

func (d *BoltDb) GetAppVersion(appID string, versionID int) (version db.AppVersion, err error) {
	err = d.getObject(0, db.AppVersionProps, intObjectID(versionID), &version)
	if err != nil {
		return
	}
	if version.AppID != appID {
		err = db.ErrNotFound
	}
	return
}

func (d *BoltDb) CreateAppVersion(version db.AppVersion) (db.AppVersion, error) {
	newObj, err := d.createObject(0, db.AppVersionProps, version)
	if err != nil {
		return version, err
	}
	return newObj.(db.AppVersion), nil
}

func (d *BoltDb) UpdateAppVersion(version db.AppVersion) error {
	return d.updateObject(0, db.AppVersionProps, version)
}

func (d *BoltDb) DeleteAppVersion(appID string, versionID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteObject(0, db.AppVersionProps, intObjectID(versionID), tx)
	})
}

func (d *BoltDb) SetAppVersionOrder(appID string, order map[int]int) error {
	for id, priority := range order {
		version, err := d.GetAppVersion(appID, id)
		if err != nil {
			return err
		}
		version.Priority = priority
		if err = d.updateObject(0, db.AppVersionProps, version); err != nil {
			return err
		}
	}
	return nil
}
