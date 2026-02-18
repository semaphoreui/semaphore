package sql

import "github.com/semaphoreui/semaphore/db"

func (d *SqlDb) GetApps() (apps []db.App, err error) {
	_, err = d.selectAll(&apps, "select * from `app` order by `priority` desc, `id`")
	return
}

func (d *SqlDb) GetApp(appID string) (app db.App, err error) {
	err = d.selectOne(&app, "select * from `app` where `id`=?", appID)
	return
}

func (d *SqlDb) CreateApp(app db.App) (db.App, error) {
	_, err := d.insert(
		"",
		"insert into `app` (`id`, `title`, `icon`, `color`, `dark_color`, `active`, `priority`) values (?, ?, ?, ?, ?, ?, ?)",
		app.ID,
		app.Title,
		app.Icon,
		app.Color,
		app.DarkColor,
		app.Active,
		app.Priority,
	)
	return app, err
}

func (d *SqlDb) UpdateApp(app db.App) error {
	_, err := d.exec(
		"update `app` set `title`=?, `icon`=?, `color`=?, `dark_color`=?, `active`=?, `priority`=? where `id`=?",
		app.Title,
		app.Icon,
		app.Color,
		app.DarkColor,
		app.Active,
		app.Priority,
		app.ID,
	)
	return err
}

func (d *SqlDb) DeleteApp(appID string) error {
	res, err := d.exec("delete from `app` where `id`=?", appID)
	return validateMutationResult(res, err)
}

func (d *SqlDb) GetAppVersions(appID string) (versions []db.AppVersion, err error) {
	_, err = d.selectAll(&versions, "select * from `app__version` where `app_id`=? order by `priority` desc, `name`", appID)
	return
}

func (d *SqlDb) GetAppVersion(appID string, versionID int) (version db.AppVersion, err error) {
	err = d.selectOne(&version, "select * from `app__version` where `app_id`=? and `id`=?", appID, versionID)
	return
}

func (d *SqlDb) CreateAppVersion(version db.AppVersion) (db.AppVersion, error) {
	insertID, err := d.insert(
		"id",
		"insert into `app__version` (`app_id`, `name`, `path`, `args`, `active`, `priority`) values (?, ?, ?, ?, ?, ?)",
		version.AppID,
		version.Name,
		version.Path,
		version.Args,
		version.Active,
		version.Priority,
	)
	if err != nil {
		return version, err
	}
	version.ID = insertID
	return version, nil
}

func (d *SqlDb) UpdateAppVersion(version db.AppVersion) error {
	_, err := d.exec(
		"update `app__version` set `name`=?, `path`=?, `args`=?, `active`=?, `priority`=? where `app_id`=? and `id`=?",
		version.Name,
		version.Path,
		version.Args,
		version.Active,
		version.Priority,
		version.AppID,
		version.ID,
	)
	return err
}

func (d *SqlDb) DeleteAppVersion(appID string, versionID int) error {
	res, err := d.exec("delete from `app__version` where `app_id`=? and `id`=?", appID, versionID)
	return validateMutationResult(res, err)
}
