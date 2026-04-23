package sql

import (
	"time"

	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetEnvironment(projectID int, environmentID int) (environment db.Environment, err error) {
	err = d.getObject(projectID, db.EnvironmentProps, environmentID, &environment)
	if err != nil {
		return
	}

	environment.SyncPaths, err = d.GetEnvironmentSyncPaths(environment.ID)
	return
}

func (d *SqlDb) GetEnvironmentRefs(projectID int, environmentID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.EnvironmentProps, environmentID)
}

func (d *SqlDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) ([]db.Environment, error) {
	var environments []db.Environment
	err := d.getObjects(projectID, db.EnvironmentProps, params, nil, &environments)
	if err != nil {
		return environments, err
	}

	for i := range environments {
		environments[i].SyncPaths, err = d.GetEnvironmentSyncPaths(environments[i].ID)
		if err != nil {
			return environments, err
		}
	}

	return environments, nil
}

func (d *SqlDb) UpdateEnvironment(env db.Environment) error {
	err := env.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__environment set "+
			"name=?, json=?, env=?, password=?, "+
			"sync_enabled=?, sync_interval=? "+
			"where id=?",
		env.Name,
		env.JSON,
		env.ENV,
		env.Password,
		env.SyncEnabled,
		env.SyncInterval,
		env.ID)

	if err != nil {
		return err
	}

	return d.ReplaceEnvironmentSyncPaths(env.ID, env.SyncPaths)
}

func (d *SqlDb) CreateEnvironment(env db.Environment) (newEnv db.Environment, err error) {
	err = env.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__environment "+
			"(project_id, name, json, env, password, secret_storage_id, secret_storage_key_prefix, "+
			"sync_enabled, sync_interval) values "+
			"(?, ?, ?, ?, ?, ?, ?, ?, ?)",
		env.ProjectID,
		env.Name,
		env.JSON,
		env.ENV,
		env.Password,
		env.SecretStorageID,
		env.SecretStorageKeyPrefix,
		env.SyncEnabled,
		env.SyncInterval)

	if err != nil {
		return
	}

	newEnv = env
	newEnv.ID = insertID

	err = d.ReplaceEnvironmentSyncPaths(newEnv.ID, env.SyncPaths)
	if err != nil {
		return
	}

	newEnv.SyncPaths, err = d.GetEnvironmentSyncPaths(newEnv.ID)
	return
}

func (d *SqlDb) DeleteEnvironment(projectID int, environmentID int) error {
	return d.deleteObject(projectID, db.EnvironmentProps, environmentID)
}

func (d *SqlDb) GetEnvironmentSecrets(projectID int, environmentID int) (keys []db.AccessKey, err error) {
	keys = make([]db.AccessKey, 0)

	q, err := d.makeObjectsQuery(projectID, db.AccessKeyProps, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	q = q.Where("pe.environment_id = ?", environmentID)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&keys, query, args...)

	return
}

func (d *SqlDb) GetEnvironmentSyncPaths(environmentID int) (paths []db.EnvironmentSyncPath, err error) {
	paths = make([]db.EnvironmentSyncPath, 0)
	_, err = d.selectAll(
		&paths,
		"select id, environment_id, path, prefix, `separator` "+
			"from project__environment__sync_path where environment_id=? order by id",
		environmentID,
	)
	return
}

func (d *SqlDb) ReplaceEnvironmentSyncPaths(environmentID int, paths []db.EnvironmentSyncPath) error {
	if _, err := d.exec("delete from project__environment__sync_path where environment_id=?", environmentID); err != nil {
		return err
	}

	for _, p := range paths {
		if _, err := d.insert(
			"id",
			"insert into project__environment__sync_path (environment_id, path, prefix, `separator`) values (?, ?, ?, ?)",
			environmentID,
			p.Path,
			p.Prefix,
			p.Separator,
		); err != nil {
			return err
		}
	}

	return nil
}

func (d *SqlDb) GetSyncEnabledEnvironments() (environments []db.Environment, err error) {
	environments = make([]db.Environment, 0)
	_, err = d.selectAll(
		&environments,
		"select * from project__environment where sync_enabled=? and sync_interval>0",
		true,
	)
	if err != nil {
		return
	}
	for i := range environments {
		environments[i].SyncPaths, err = d.GetEnvironmentSyncPaths(environments[i].ID)
		if err != nil {
			return
		}
	}
	return
}

func (d *SqlDb) MarkEnvironmentSynced(environmentID int, success bool, at time.Time) error {
	var query string
	if success {
		query = "update project__environment set last_synced_at=? where id=?"
	} else {
		query = "update project__environment set last_sync_failed_at=? where id=?"
	}
	_, err := d.exec(query, at, environmentID)
	return err
}
