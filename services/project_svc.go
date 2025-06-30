package services

import (
	"github.com/semaphoreui/semaphore/db"
)

type ProjectService interface {
	UpdateProject(project db.Project) error
}

type ProjectServiceImpl struct {
	store db.Store
}

func NewProjectService(store db.Store) *ProjectServiceImpl {
	return &ProjectServiceImpl{
		store: store,
	}
}

func (s *ProjectServiceImpl) UpdateProject(project db.Project) (err error) {
	err = s.store.UpdateProject(project)
	if err != nil {
		return
	}

	keys, err := s.store.GetAccessKeys(project.ID, db.GetAccessKeyOptions{
		Type: db.AccessKeyVault,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		if project.VaultToken != "" {
			_, err = s.store.CreateAccessKey(db.AccessKey{
				Type:      db.AccessKeyLoginPassword,
				ProjectID: &project.ID,
				Secret:    nil,
				String:    project.VaultToken,
				Owner:     db.AccessKeyVault,
				Plain:     nil,
			})
		}
	} else {
		vault := keys[0]
		if project.VaultToken == "" {
			err = s.store.DeleteAccessKey(project.ID, vault.ID)
		} else {
			vault.OverrideSecret = true
			vault.Secret = &project.VaultToken
			err = s.store.UpdateAccessKey(vault)
		}
	}

	return
}
