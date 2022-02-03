package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) GetRepository(projectID int, repositoryID int) (repository db.Repository, err error) {
	err = d.getObject(projectID, db.RepositoryProps, intObjectID(repositoryID), &repository)
	if err != nil {
		return
	}
	repository.SSHKey, err = d.GetAccessKey(projectID, repository.SSHKeyID)
	return
}

func (d *BoltDb) GetRepositoryRefs(projectID int, repositoryID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.RepositoryProps, repositoryID)
}

func (d *BoltDb) GetRepositories(projectID int, params db.RetrieveQueryParams) (repositories []db.Repository, err error) {
	err = d.getObjects(projectID, db.RepositoryProps, params, nil, &repositories)
	return
}

func (d *BoltDb) UpdateRepository(repository db.Repository) error {
	err := repository.Validate()
	if err != nil {
		return err
	}
	return d.updateObject(repository.ProjectID, db.RepositoryProps, repository)
}

func (d *BoltDb) CreateRepository(repository db.Repository) (db.Repository, error) {
	err := repository.Validate()
	if err != nil {
		return db.Repository{}, err
	}
	newRepo, err := d.createObject(repository.ProjectID, db.RepositoryProps, repository)
	return newRepo.(db.Repository), err
}

func (d *BoltDb) DeleteRepository(projectID int, repositoryId int) error {
	return d.deleteObject(projectID, db.RepositoryProps, intObjectID(repositoryId), nil)
}
