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
	repoDir := path.Join(util.Config.TmpPath, "repository_123_55")
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	repo := Repository{ID: 123}
	err = repo.ClearCache()
	require.NoError(t, err)

	_, err = os.Stat(repoDir)
	require.Error(t, err, "repo directory not deleted")
	assert.True(t, os.IsNotExist(err))
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
		gitUrl := v.Repository.GetGitURL()
		assert.Equal(t, v.ExpectedGitUrl, gitUrl, "wrong gitUrl")
	}
}
