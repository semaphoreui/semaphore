package db

import (
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/semaphoreui/semaphore/util"
)

func TestRepository_GetSchema(t *testing.T) {
	repo := Repository{GitURL: "https://example.com/hello/world"}
	schema := repo.GetType()
	assert.Equal(t, RepositoryHTTP, schema)
}

func TestRepository_GetType_WindowsLocalPath(t *testing.T) {
	assert.Equal(t, RepositoryLocal, Repository{GitURL: `D:\repo`}.GetType())
	assert.Equal(t, RepositoryLocal, Repository{GitURL: `D:/repo`}.GetType())
	assert.Equal(t, RepositoryLocal, Repository{GitURL: `D:`}.GetType())
	assert.Equal(t, RepositoryLocal, Repository{GitURL: `\\server\share`}.GetType())
}

func TestRepository_ClearCache(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: path.Join(os.TempDir(), util.RandString(rand.Intn(10-4)+4)),
	}
	repoDir := path.Join(util.Config.TmpPath, "project_0", "repository_123_55")
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	repo := Repository{ID: 123}
	err = repo.ClearCache()
	require.NoError(t, err)

	_, err = os.Stat(repoDir)
	require.Error(t, err, "repo directory not deleted")
	assert.True(t, os.IsNotExist(err))
}

func TestRepository_GetFullPath_IgnoresWorkingCopyPath(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}

	repo := Repository{ID: 1, GitURL: "https://example.com/x.git", WorkingCopyPath: "/task/private/copy"}
	assert.Equal(t, "/tmp/project_0/repository_1_template_5", repo.GetFullPath(5))
}

func TestRepository_GetWorkingCopyPath(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp"}

	tests := []struct {
		name         string
		repository   Repository
		expectedPath string
	}{{
		name:         "WorkingCopyPath wins over the default shared path",
		repository:   Repository{ID: 1, GitURL: "https://example.com/x.git", WorkingCopyPath: "/task/private/copy"},
		expectedPath: "/task/private/copy",
	}, {
		name:         "default shared path when WorkingCopyPath is unset",
		repository:   Repository{ID: 1, GitURL: "https://example.com/x.git"},
		expectedPath: "/tmp/project_0/repository_1_template_5",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedPath, tt.repository.GetWorkingCopyPath(5))
		})
	}
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
