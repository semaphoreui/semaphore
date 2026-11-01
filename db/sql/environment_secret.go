package sql

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// isUniqueViolation recognises the unique-index error of SQLite, MySQL and Postgres.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

func environmentSecretExistsError(secret db.EnvironmentSecret) error {
	return common_errors.NewValidationError(
		"secret " + secret.Name + " (" + string(secret.Type) + ") already exists in this variable group")
}

func (d *SqlDb) GetEnvironmentSecretKeys(projectID int, environmentID int) (res []db.EnvironmentSecretWithKey, err error) {
	res = make([]db.EnvironmentSecretWithKey, 0)

	var secrets []db.EnvironmentSecret
	_, err = d.selectAll(&secrets,
		"select * from project__environment_secret where project_id=? and environment_id=? order by type, name",
		projectID, environmentID)
	if err != nil || len(secrets) == 0 {
		return
	}

	ids := make([]int, 0, len(secrets))
	for _, s := range secrets {
		ids = append(ids, s.KeyID)
	}

	query, args, err := squirrel.Select("*").From("access_key").
		Where("project_id = ?", projectID).
		Where(squirrel.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return
	}

	var keys []db.AccessKey
	if _, err = d.selectAll(&keys, query, args...); err != nil {
		return
	}

	byID := make(map[int]db.AccessKey, len(keys))
	for _, k := range keys {
		byID[k.ID] = k
	}

	for _, s := range secrets {
		key, ok := byID[s.KeyID]
		if !ok {
			// The foreign key keeps this from happening; a missing key must not silently drop the secret.
			err = common_errors.NewValidationError("secret " + s.Name + " points to a key that does not exist")
			return
		}
		s.KeyName = key.Name
		s.Owned = key.EnvironmentID != nil && *key.EnvironmentID == environmentID
		res = append(res, db.EnvironmentSecretWithKey{EnvironmentSecret: s, Key: key})
	}

	return
}

func (d *SqlDb) GetEnvironmentSecret(projectID int, environmentID int, secretID int) (secret db.EnvironmentSecret, err error) {
	err = d.selectOne(&secret,
		"select * from project__environment_secret where project_id=? and environment_id=? and id=?",
		projectID, environmentID, secretID)
	return
}

func (d *SqlDb) CreateEnvironmentSecret(secret db.EnvironmentSecret) (newSecret db.EnvironmentSecret, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__environment_secret (project_id, environment_id, type, name, key_id) values (?, ?, ?, ?, ?)",
		secret.ProjectID,
		secret.EnvironmentID,
		secret.Type,
		secret.Name,
		secret.KeyID)
	if err != nil {
		if isUniqueViolation(err) {
			err = environmentSecretExistsError(secret)
		} else {
			err = validateMutationResult(nil, err)
		}
		return
	}

	newSecret = secret
	newSecret.ID = insertID
	return
}

func (d *SqlDb) UpdateEnvironmentSecret(secret db.EnvironmentSecret) error {
	res, err := d.exec(
		"update project__environment_secret set type=?, name=?, key_id=? where project_id=? and environment_id=? and id=?",
		secret.Type,
		secret.Name,
		secret.KeyID,
		secret.ProjectID,
		secret.EnvironmentID,
		secret.ID)
	if isUniqueViolation(err) {
		return environmentSecretExistsError(secret)
	}
	if err = validateMutationResult(res, err); err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		_, err = d.GetEnvironmentSecret(secret.ProjectID, secret.EnvironmentID, secret.ID)
		return err
	}
	return nil
}

func (d *SqlDb) DeleteEnvironmentSecret(projectID int, environmentID int, secretID int) error {
	return requireDeletedRow(d.exec(
		"delete from project__environment_secret where project_id=? and environment_id=? and id=?",
		projectID, environmentID, secretID))
}

func (d *SqlDb) DeleteEnvironmentSecretsByKey(projectID int, keyID int) error {
	_, err := d.exec("delete from project__environment_secret where project_id=? and key_id=?", projectID, keyID)
	return err
}

func (d *SqlDb) CountEnvironmentSecretsByKey(projectID int, keyID int) (int, error) {
	n, err := d.Sql().SelectInt(
		d.PrepareQuery("select count(*) from project__environment_secret where project_id=? and key_id=?"),
		projectID, keyID)
	return int(n), err
}

func (d *SqlDb) GetAccessKeyEnvironmentRefs(projectID int, keyID int) (refs []db.ObjectReferrer, err error) {
	refs = make([]db.ObjectReferrer, 0)
	_, err = d.selectAll(&refs,
		"select distinct e.id, e.name from project__environment_secret s "+
			"join project__environment e on e.id = s.environment_id "+
			"where s.project_id=? and s.key_id=? order by e.name",
		projectID, keyID)
	return
}
