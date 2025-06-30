package services

import (
	"github.com/semaphoreui/semaphore/db"
)

type ProjectService interface {
	UpdateProject(project db.Project) error
	DeleteProject(projectID int) error
}

func NewProjectService(
	projectRepo db.ProjectStore,
	keyRepo db.AccessKeyManager,
) ProjectService {
	return &ProjectServiceImpl{
		projectRepo: projectRepo,
		keyRepo:     keyRepo,
	}
}

type ProjectServiceImpl struct {
	projectRepo db.ProjectStore
	keyRepo     db.AccessKeyManager
}

func (s *ProjectServiceImpl) DeleteProject(projectID int) error {
	return s.projectRepo.DeleteProject(projectID)
}

func (s *ProjectServiceImpl) UpdateProject(project db.Project) (err error) {
	err = s.projectRepo.UpdateProject(project)
	if err != nil {
		return
	}

	keys, err := s.keyRepo.GetAccessKeys(project.ID, db.GetAccessKeyOptions{
		Owner: db.AccessKeyVault,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		if project.VaultToken != "" {
			_, err = s.keyRepo.CreateAccessKey(db.AccessKey{
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
			err = s.keyRepo.DeleteAccessKey(project.ID, vault.ID)
		} else {
			vault.OverrideSecret = true
			vault.Secret = &project.VaultToken
			err = s.keyRepo.UpdateAccessKey(vault)
		}
	}

	return
}
