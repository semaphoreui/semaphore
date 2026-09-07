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

func TestRepository_GetGitURL(t *testing.T) {
	for _, v := range []struct {
		Repository     Repository
		Secure         bool
		ExpectedGitUrl string
	}{
		{
			Repository: Repository{
				GitURL: "https://github.com/user/project.git",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "login",
						Password: "password",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "https://login:password@github.com/user/project.git",
		},
		{
			Repository: Repository{
				GitURL: "https://github.com/user/project.git",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Password: "password",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "https://password@github.com/user/project.git",
		},
		{
			Repository: Repository{
				GitURL: "https://devops.domain.com/tfs/project/_git/repo",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "user@domain.com",
						Password: "pass#word@123",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "https://user%40domain.com:pass%23word%40123@devops.domain.com/tfs/project/_git/repo",
		},
		{
			Repository: Repository{
				GitURL: "https://devops.domain.com/tfs/project/_git/repo",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "user",
						Password: "p@ss:w%rd+1&2?3",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "https://user:p%40ss%3Aw%25rd+1&2%3F3@devops.domain.com/tfs/project/_git/repo",
		},
		{
			Repository: Repository{
				GitURL: "https://devops.domain.com/tfs/project/_git/repo",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Password: "token%with#special@chars",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "https://token%25with%23special%40chars@devops.domain.com/tfs/project/_git/repo",
		},
		{
			Repository: Repository{
				GitURL: "https://devops.domain.com/tfs/project/_git/repo",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "user@domain.com",
						Password: "pass#word@123",
					},
				},
			},
			Secure:         true,
			ExpectedGitUrl: "https://devops.domain.com/tfs/project/_git/repo",
		},
		{
			Repository: Repository{
				GitURL: "https://user:secret@devops.domain.com:8443/tfs/project/_git/repo",
			},
			Secure:         true,
			ExpectedGitUrl: "https://devops.domain.com:8443/tfs/project/_git/repo",
		},
		{
			Repository: Repository{
				GitURL: "git@github.com:user/project.git",
			},
			Secure:         true,
			ExpectedGitUrl: "git@github.com:user/project.git",
		},
		{
			Repository: Repository{
				GitURL: "/tmp/local/repo",
			},
			Secure:         false,
			ExpectedGitUrl: "/tmp/local/repo",
		},
	} {
		gitUrl := v.Repository.GetGitURL(v.Secure)
		assert.Equal(t, v.ExpectedGitUrl, gitUrl, "wrong gitUrl")
	}
}
