package server

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

	return
}
