package db

import (
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_GetSchema(t *testing.T) {
	repo := Repository{GitURL: "https://example.com/hello/world"}
	schema := repo.GetType()
	assert.Equal(t, RepositoryHTTP, schema)
}

func TestRepository_ClearCache(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: path.Join(os.TempDir(), util.RandString(rand.Intn(10-4)+4)),
	}
	repoDir := path.Join(util.Config.TmpPath, "project_0", "template_55")
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	// Create a mock template manager that returns a template using this repository
	mockTemplateManager := &mockTemplateManager{
		templates: []Template{
			{ID: 55, ProjectID: 0, RepositoryID: 123},
		},
	}

	repo := Repository{ID: 123, ProjectID: 0}
	err = repo.ClearCache(mockTemplateManager)
	require.NoError(t, err)

	_, err = os.Stat(repoDir)
	require.Error(t, err, "repo directory not deleted")
	assert.True(t, os.IsNotExist(err))
}

func TestRepository_ClearCache_NoTemplatesUsingRepo(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: path.Join(os.TempDir(), util.RandString(rand.Intn(10-4)+4)),
	}
	repoDir := path.Join(util.Config.TmpPath, "project_0", "template_55")
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	// Create a mock template manager with no templates using this repository
	mockTemplateManager := &mockTemplateManager{
		templates: []Template{
			{ID: 55, ProjectID: 0, RepositoryID: 999}, // Different repository ID
		},
	}

	repo := Repository{ID: 123, ProjectID: 0}
	err = repo.ClearCache(mockTemplateManager)
	require.NoError(t, err)

	// Directory should still exist since no templates use this repository
	_, err = os.Stat(repoDir)
	require.NoError(t, err, "repo directory should not be deleted")
}

// Mock template manager for Repository tests
type mockTemplateManager struct {
	templates []Template
}

func (m *mockTemplateManager) GetTemplates(projectID int, filter TemplateFilter, params RetrieveQueryParams) ([]Template, error) {
	var result []Template
	for _, template := range m.templates {
		if template.ProjectID == projectID {
			result = append(result, template)
		}
	}
	return result, nil
}

func (m *mockTemplateManager) GetTemplateRefs(projectID int, templateID int) (ObjectReferrers, error) {
	return ObjectReferrers{}, nil
}

func (m *mockTemplateManager) CreateTemplate(template Template) (Template, error) {
	return Template{}, nil
}

func (m *mockTemplateManager) UpdateTemplate(template Template) error {
	return nil
}

func (m *mockTemplateManager) GetTemplate(projectID int, templateID int) (Template, error) {
	return Template{}, nil
}

func (m *mockTemplateManager) DeleteTemplate(projectID int, templateID int) error {
	return nil
}

func (m *mockTemplateManager) SetTemplateDescription(projectID int, templateID int, description string) error {
	return nil
}

func (m *mockTemplateManager) GetTemplateVaults(projectID int, templateID int) ([]TemplateVault, error) {
	return nil, nil
}

func (m *mockTemplateManager) CreateTemplateVault(vault TemplateVault) (TemplateVault, error) {
	return TemplateVault{}, nil
}

func (m *mockTemplateManager) UpdateTemplateVaults(projectID int, templateID int, vaults []TemplateVault) error {
	return nil
}

func TestRepository_GetGitURL(t *testing.T) {
	for _, v := range []struct {
		Repository     Repository
		ExpectedGitUrl string
	}{
		{
			Repository: Repository{GitURL: "https://github.com/user/project.git", SSHKey: AccessKey{
				Type: AccessKeyLoginPassword,
				LoginPassword: LoginPassword{
					Login:    "login",
					Password: "password",
				},
			},
			},
			ExpectedGitUrl: "https://login:password@github.com/user/project.git",
		},
		{
			Repository: Repository{GitURL: "https://github.com/user/project.git", SSHKey: AccessKey{
				Type: AccessKeyLoginPassword,
				LoginPassword: LoginPassword{
					Password: "password",
				},
			},
			},
			ExpectedGitUrl: "https://password@github.com/user/project.git",
		},
	} {
		gitUrl := v.Repository.GetGitURL(false)
		assert.Equal(t, v.ExpectedGitUrl, gitUrl, "wrong gitUrl")
	}
}
