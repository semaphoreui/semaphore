package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetEnvironment(projectID int, environmentID int) (environment db.Environment, err error) {
	err = d.getObject(projectID, db.EnvironmentProps, environmentID, &environment)
	return
}

func (d *SqlDb) GetEnvironmentRefs(projectID int, environmentID int) (refs db.ObjectReferrers, err error) {
	refs, err = d.getObjectRefs(projectID, db.EnvironmentProps, environmentID)
	if err != nil {
		return
	}

	var extra []db.ObjectReferrer
	_, err = d.selectAll(
		&extra,
		"select t.id, t.name from project__template t "+
			"join project__template_environment pte "+
			"on pte.template_id = t.id and pte.project_id = t.project_id "+
			"where t.project_id = ? and pte.environment_id = ?",
		projectID,
		environmentID,
	)

	if err != nil {
		return
	}

	seen := make(map[int]bool)
	for _, r := range refs.Templates {
		seen[r.ID] = true
	}
	for _, r := range extra {
		if seen[r.ID] {
			continue
		}
		refs.Templates = append(refs.Templates, r)
	}

	return
}

func (d *SqlDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) ([]db.Environment, error) {
	var environment []db.Environment
	err := d.getObjects(projectID, db.EnvironmentProps, params, nil, &environment)
	return environment, err
}

func (d *SqlDb) UpdateEnvironment(env db.Environment) error {
	err := env.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__environment set name=?, json=?, env=?, password=? where id=?",
		env.Name,
		env.JSON,
		env.ENV,
		env.Password,
		env.ID)
	return err
}

func (d *SqlDb) CreateEnvironment(env db.Environment) (newEnv db.Environment, err error) {
	err = env.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__environment "+
			"(project_id, name, json, env, password, secret_storage_id, secret_storage_key_prefix) values "+
			"(?, ?, ?, ?, ?, ?, ?)",
		env.ProjectID,
		env.Name,
		env.JSON,
		env.ENV,
		env.Password,
		env.SecretStorageID,
		env.SecretStorageKeyPrefix)

	if err != nil {
		return
	}

	newEnv = env
	newEnv.ID = insertID
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
