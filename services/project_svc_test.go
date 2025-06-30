package services

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

type mockProjectStore struct {
	UpdateProjectFn func(project db.Project) error
	DeleteProjectFn func(projectID int) error
}

func (m *mockProjectStore) UpdateProject(project db.Project) error {
	if m.UpdateProjectFn != nil {
		return m.UpdateProjectFn(project)
	}
	return nil
}
func (m *mockProjectStore) DeleteProject(projectID int) error {
	if m.DeleteProjectFn != nil {
		return m.DeleteProjectFn(projectID)
	}
	return nil
}

// Stub methods to satisfy db.ProjectStore
func (m *mockProjectStore) GetProject(projectID int) (db.Project, error) { return db.Project{}, nil }
func (m *mockProjectStore) GetAllProjects() ([]db.Project, error)        { return nil, nil }
func (m *mockProjectStore) GetProjects(userID int) ([]db.Project, error) { return nil, nil }
func (m *mockProjectStore) CreateProject(project db.Project) (db.Project, error) {
	return db.Project{}, nil
}
func (m *mockProjectStore) GetProjectUsers(projectID int, params db.RetrieveQueryParams) ([]db.UserWithProjectRole, error) {
	return nil, nil
}
func (m *mockProjectStore) CreateProjectUser(projectUser db.ProjectUser) (db.ProjectUser, error) {
	return db.ProjectUser{}, nil
}
func (m *mockProjectStore) DeleteProjectUser(projectID int, userID int) error { return nil }
func (m *mockProjectStore) GetProjectUser(projectID int, userID int) (db.ProjectUser, error) {
	return db.ProjectUser{}, nil
}
func (m *mockProjectStore) UpdateProjectUser(projectUser db.ProjectUser) error { return nil }

type mockAccessKeyManager struct {
	GetAccessKeysFn   func(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error)
	CreateAccessKeyFn func(key db.AccessKey) (db.AccessKey, error)
	DeleteAccessKeyFn func(projectID, keyID int) error
	UpdateAccessKeyFn func(key db.AccessKey) error
}

func (m *mockAccessKeyManager) GetAccessKeys(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	if m.GetAccessKeysFn != nil {
		return m.GetAccessKeysFn(projectID, opts, params)
	}
	return nil, nil
}
func (m *mockAccessKeyManager) CreateAccessKey(key db.AccessKey) (db.AccessKey, error) {
	if m.CreateAccessKeyFn != nil {
		return m.CreateAccessKeyFn(key)
	}
	return db.AccessKey{}, nil
}
func (m *mockAccessKeyManager) DeleteAccessKey(projectID, keyID int) error {
	if m.DeleteAccessKeyFn != nil {
		return m.DeleteAccessKeyFn(projectID, keyID)
	}
	return nil
}
func (m *mockAccessKeyManager) UpdateAccessKey(key db.AccessKey) error {
	if m.UpdateAccessKeyFn != nil {
		return m.UpdateAccessKeyFn(key)
	}
	return nil
}

// Stub methods to satisfy db.AccessKeyManager
func (m *mockAccessKeyManager) GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error) {
	return db.AccessKey{}, nil
}
func (m *mockAccessKeyManager) GetAccessKeyRefs(projectID int, accessKeyID int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockAccessKeyManager) RekeyAccessKeys(oldKey string) error { return nil }

func TestProjectServiceImpl_DeleteProject(t *testing.T) {
	mockRepo := &mockProjectStore{
		DeleteProjectFn: func(projectID int) error {
			if projectID == 42 {
				return nil
			}
			return errors.New("not found")
		},
	}
	service := &ProjectServiceImpl{projectRepo: mockRepo}

	err := service.DeleteProject(42)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	err = service.DeleteProject(1)
	if err == nil || err.Error() != "not found" {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestProjectServiceImpl_UpdateProject(t *testing.T) {
	project := db.Project{ID: 1, VaultToken: "token"}

	t.Run("no keys, VaultToken set", func(t *testing.T) {
		updated := false
		created := false
		mockRepo := &mockProjectStore{
			UpdateProjectFn: func(p db.Project) error {
				updated = true
				return nil
			},
		}
		mockKey := &mockAccessKeyManager{
			GetAccessKeysFn: func(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
				return nil, nil
			},
			CreateAccessKeyFn: func(key db.AccessKey) (db.AccessKey, error) {
				created = true
				return key, nil
			},
		}
		service := &ProjectServiceImpl{projectRepo: mockRepo, keyRepo: mockKey}
		err := service.UpdateProject(project)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !updated || !created {
			t.Errorf("expected update and create to be called")
		}
	})

	t.Run("no keys, VaultToken empty", func(t *testing.T) {
		mockRepo := &mockProjectStore{
			UpdateProjectFn: func(p db.Project) error { return nil },
		}
		mockKey := &mockAccessKeyManager{
			GetAccessKeysFn: func(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
				return nil, nil
			},
			CreateAccessKeyFn: func(key db.AccessKey) (db.AccessKey, error) {
				t.Errorf("should not create access key")
				return key, nil
			},
		}
		service := &ProjectServiceImpl{projectRepo: mockRepo, keyRepo: mockKey}
		p := db.Project{ID: 2, VaultToken: ""}
		err := service.UpdateProject(p)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("keys exist, VaultToken empty", func(t *testing.T) {
		deleted := false
		mockRepo := &mockProjectStore{
			UpdateProjectFn: func(p db.Project) error { return nil },
		}
		mockKey := &mockAccessKeyManager{
			GetAccessKeysFn: func(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
				return []db.AccessKey{{ID: 5}}, nil
			},
			DeleteAccessKeyFn: func(projectID, keyID int) error {
				if projectID == 3 && keyID == 5 {
					deleted = true
					return nil
				}
				return errors.New("wrong id")
			},
			UpdateAccessKeyFn: func(key db.AccessKey) error {
				t.Errorf("should not update access key")
				return nil
			},
		}
		service := &ProjectServiceImpl{projectRepo: mockRepo, keyRepo: mockKey}
		p := db.Project{ID: 3, VaultToken: ""}
		err := service.UpdateProject(p)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !deleted {
			t.Errorf("expected delete to be called")
		}
	})

	t.Run("keys exist, VaultToken set", func(t *testing.T) {
		updated := false
		mockRepo := &mockProjectStore{
			UpdateProjectFn: func(p db.Project) error { return nil },
		}
		mockKey := &mockAccessKeyManager{
			GetAccessKeysFn: func(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
				return []db.AccessKey{{ID: 6}}, nil
			},
			DeleteAccessKeyFn: func(projectID, keyID int) error {
				return nil
			},
			UpdateAccessKeyFn: func(key db.AccessKey) error {
				updated = true
				if !key.OverrideSecret || key.Secret == nil || *key.Secret != "token2" {
					t.Errorf("unexpected key update: %+v", key)
				}
				return nil
			},
		}
		service := &ProjectServiceImpl{projectRepo: mockRepo, keyRepo: mockKey}
		p := db.Project{ID: 4, VaultToken: "token2"}
		err := service.UpdateProject(p)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !updated {
			t.Errorf("expected update to be called")
		}
	})

	t.Run("UpdateProject returns error", func(t *testing.T) {
		mockRepo := &mockProjectStore{
			UpdateProjectFn: func(p db.Project) error { return errors.New("fail") },
		}
		mockKey := &mockAccessKeyManager{}
		service := &ProjectServiceImpl{projectRepo: mockRepo, keyRepo: mockKey}
		err := service.UpdateProject(project)
		if err == nil || err.Error() != "fail" {
			t.Errorf("expected fail error, got %v", err)
		}
	})

	t.Run("GetAccessKeys returns error", func(t *testing.T) {
		mockRepo := &mockProjectStore{
			UpdateProjectFn: func(p db.Project) error { return nil },
		}
		mockKey := &mockAccessKeyManager{
			GetAccessKeysFn: func(projectID int, opts db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
				return nil, errors.New("failkeys")
			},
		}
		service := &ProjectServiceImpl{projectRepo: mockRepo, keyRepo: mockKey}
		err := service.UpdateProject(project)
		if err == nil || err.Error() != "failkeys" {
			t.Errorf("expected failkeys error, got %v", err)
		}
	})
}
