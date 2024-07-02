package bolt

import "github.com/ansible-semaphore/semaphore/db"

func (d *BoltDb) GetEnvironment(projectID int, environmentID int) (environment db.Environment, err error) {
	err = d.getObject(projectID, db.EnvironmentProps, intObjectID(environmentID), &environment)
	return
}

func (d *BoltDb) GetEnvironmentRefs(projectID int, environmentID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.EnvironmentProps, environmentID)
}

func (d *BoltDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) (environment []db.Environment, err error) {
	err = d.getObjects(projectID, db.EnvironmentProps, params, nil, &environment)
	return
}

func (d *BoltDb) UpdateEnvironment(env db.Environment) error {
	err := env.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(env.ProjectID, db.EnvironmentProps, env)
}

func (d *BoltDb) CreateEnvironment(env db.Environment) (db.Environment, error) {
	err := env.Validate()

	if err != nil {
		return db.Environment{}, err
	}

	newEnv, err := d.createObject(env.ProjectID, db.EnvironmentProps, env)
	return newEnv.(db.Environment), err
}

func (d *BoltDb) DeleteEnvironment(projectID int, environmentID int) error {
	return d.deleteObject(projectID, db.EnvironmentProps, intObjectID(environmentID), nil)
}

func (d *BoltDb) GetEnvironmentSecrets(projectID int, environmentID int) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, db.AccessKeyProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		k := i.(db.AccessKey)
		return k.EnvironmentID != nil && *k.EnvironmentID == environmentID
	}, &keys)
	return keys, err
}
