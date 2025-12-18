package bolt

import (
	"sort"

	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetTemplateAdditionalRepositories(projectID int, templateID int) (repos []db.TemplateAdditionalRepository, err error) {
	err = d.getObjects(projectID, db.TemplateAdditionalRepositoryProps, db.RetrieveQueryParams{
		SortBy: "position",
	}, func(i interface{}) bool {
		repo := i.(db.TemplateAdditionalRepository)
		return repo.TemplateID == templateID
	}, &repos)

	if err != nil {
		return
	}

	// Sort by position
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].Position == repos[j].Position {
			return repos[i].ID < repos[j].ID
		}
		return repos[i].Position < repos[j].Position
	})

	// Expand Repository data
	for i := range repos {
		var repo db.Repository
		repo, err = d.GetRepository(projectID, repos[i].RepositoryID)
		if err != nil {
			return
		}
		repos[i].Repository = &repo
	}

	return
}

func (d *BoltDb) UpdateTemplateAdditionalRepositories(projectID int, templateID int, repos []db.TemplateAdditionalRepository) (err error) {
	var oldRepos []db.TemplateAdditionalRepository

	oldRepos, err = d.GetTemplateAdditionalRepositories(projectID, templateID)
	if err != nil {
		return err
	}

	err = d.db.Update(func(tx *bbolt.Tx) error {
		// Delete all old additional repositories
		for _, oldRepo := range oldRepos {
			err = d.deleteObject(projectID, db.TemplateAdditionalRepositoryProps, intObjectID(oldRepo.ID), tx)
			if err != nil {
				return err
			}
		}

		// Create new additional repositories
		for _, repo := range repos {
			// Load repository data for validation
			var repository db.Repository
			repository, err = d.GetRepository(projectID, repo.RepositoryID)
			if err != nil {
				return err
			}
			repo.Repository = &repository

			err = repo.Validate()
			if err != nil {
				return err
			}

			repo.TemplateID = templateID

			_, err = d.createObjectTx(tx, projectID, db.TemplateAdditionalRepositoryProps, repo)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}
