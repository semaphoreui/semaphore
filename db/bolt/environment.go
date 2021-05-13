package bolt

import "github.com/ansible-semaphore/semaphore/db"

func (d *BoltDb) GetEnvironment(projectID int, environmentID int) (environment db.Environment, err error) {
	err = d.getObject(projectID, db.EnvironmentProps, intObjectID(environmentID), &environment)
	return
}

func (d *BoltDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) (environment []db.Environment, err error) {
	err = d.getObjects(projectID, db.EnvironmentProps, params, nil, &environment)
	return
}

func (d *BoltDb) UpdateEnvironment(env db.Environment) error {
	return d.updateObject(env.ProjectID, db.EnvironmentProps, &env)
}

func (d *BoltDb) CreateEnvironment(env db.Environment) (db.Environment, error) {
	newEnv, err := d.createObject(env.ProjectID, db.EnvironmentProps, env)
	return newEnv.(db.Environment), err
}

func (d *BoltDb) DeleteEnvironment(projectID int, environmentID int) error {
	return d.deleteObject(projectID, db.EnvironmentProps, intObjectID(environmentID))
}

func (d *BoltDb) DeleteEnvironmentSoft(projectID int, environmentID int) error {
	return d.deleteObjectSoft(projectID, db.EnvironmentProps, intObjectID(environmentID))
}
