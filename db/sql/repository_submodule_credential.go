package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetRepositorySubmoduleCredential(projectID int, repositoryID int, credentialID int) (cred db.RepositorySubmoduleCredential, err error) {
	err = d.getObject(projectID, db.RepositorySubmoduleCredentialProps, credentialID, &cred)
	if err != nil {
		return
	}

	if cred.RepositoryID != repositoryID {
		err = db.ErrNotFound
		return
	}

	cred.AccessKey, err = d.GetAccessKey(projectID, cred.AccessKeyID)

	return
}

func (d *SqlDb) GetRepositorySubmoduleCredentials(projectID int, repositoryID int) (creds []db.RepositorySubmoduleCredential, err error) {
	creds = make([]db.RepositorySubmoduleCredential, 0)

	query, args, err := squirrel.Select("*").
		From("`"+db.RepositorySubmoduleCredentialProps.TableName+"`").
		Where("project_id=? and repository_id=?", projectID, repositoryID).
		OrderBy("host").
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&creds, query, args...)
	if err != nil {
		return
	}

	for i := range creds {
		creds[i].AccessKey, err = d.GetAccessKey(projectID, creds[i].AccessKeyID)
		if err != nil {
			return
		}
	}

	return
}

// validateSubmoduleCredentialAccessKey ensures the referenced access key exists
// and belongs to the same project as the repository, preventing an IDOR where a
// submodule credential could be created pointing at another project's key.
func (d *SqlDb) validateSubmoduleCredentialAccessKey(cred db.RepositorySubmoduleCredential) error {
	_, err := d.GetAccessKey(cred.ProjectID, cred.AccessKeyID)
	return err
}

func (d *SqlDb) CreateRepositorySubmoduleCredential(cred db.RepositorySubmoduleCredential) (newCred db.RepositorySubmoduleCredential, err error) {
	if err = cred.Validate(); err != nil {
		return
	}

	if err = d.validateSubmoduleCredentialAccessKey(cred); err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into `"+db.RepositorySubmoduleCredentialProps.TableName+"` (project_id, repository_id, access_key_id, host) values (?, ?, ?, ?)",
		cred.ProjectID,
		cred.RepositoryID,
		cred.AccessKeyID,
		cred.Host)

	if err != nil {
		return
	}

	newCred = cred
	newCred.ID = insertID
	return
}

func (d *SqlDb) UpdateRepositorySubmoduleCredential(cred db.RepositorySubmoduleCredential) error {
	if err := cred.Validate(); err != nil {
		return err
	}

	if err := d.validateSubmoduleCredentialAccessKey(cred); err != nil {
		return err
	}

	res, err := d.exec(
		"update `"+db.RepositorySubmoduleCredentialProps.TableName+"` set access_key_id=?, host=? where id=? and project_id=? and repository_id=?",
		cred.AccessKeyID,
		cred.Host,
		cred.ID,
		cred.ProjectID,
		cred.RepositoryID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) DeleteRepositorySubmoduleCredential(projectID int, repositoryID int, credentialID int) error {
	_, err := d.GetRepositorySubmoduleCredential(projectID, repositoryID, credentialID)
	if err != nil {
		return err
	}

	return d.deleteObject(projectID, db.RepositorySubmoduleCredentialProps, credentialID)
}
