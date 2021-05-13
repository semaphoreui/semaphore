package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) GetRepository(projectID int, repositoryID int) (repository db.Repository, err error) {
	err = d.getObject(projectID, db.RepositoryObject, intObjectID(repositoryID), &repository)
	return
}

func (d *BoltDb) GetRepositories(projectID int, params db.RetrieveQueryParams) (repositories []db.Repository, err error) {
	err = d.getObjects(projectID, db.RepositoryObject, params, nil, &repositories)
	return
}

func (d *BoltDb) UpdateRepository(repository db.Repository) error {
	return d.updateObject(repository.ProjectID, db.RepositoryObject, repository)
}

func (d *BoltDb) CreateRepository(repository db.Repository) (db.Repository, error) {
	newRepo, err := d.createObject(repository.ProjectID, db.RepositoryObject, repository)
	return newRepo.(db.Repository), err
}

func (d *BoltDb) DeleteRepository(projectID int, repositoryId int) error {
	return d.deleteObject(projectID, db.RepositoryObject, intObjectID(repositoryId))
}

func (d *BoltDb) DeleteRepositorySoft(projectID int, repositoryId int) error {
	return d.deleteObjectSoft(projectID, db.RepositoryObject, intObjectID(repositoryId))
}

