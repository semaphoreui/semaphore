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
				GitURL: "https://github.com/user/project.git",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "user@corp.com",
						Password: "p#ss%word/with/slashes@123",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "https://user%40corp.com:p%23ss%25word%2Fwith%2Fslashes%40123@github.com/user/project.git",
		},
		{
			Repository: Repository{
				GitURL: "http://insecure.local/user/project.git",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "user",
						Password: "secretpassword",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "http://insecure.local/user/project.git",
		},
		{
			Repository: Repository{
				GitURL: "https://user:secret@github.com/user/project.git",
			},
			Secure:         true,
			ExpectedGitUrl: "https://github.com/user/project.git",
		},
	} {
		gitUrl := v.Repository.GetGitURL(v.Secure)
		assert.Equal(t, v.ExpectedGitUrl, gitUrl, "wrong gitUrl")
	}
}

func TestRepository_GetCheckoutDirName(t *testing.T) {
	repo := Repository{ID: 1, GitURL: "https://example.com/hello/world"}

	repo.GitBranch = "main"
	main := repo.GetCheckoutDirName(2)
	assert.Equal(t, "repository_1_template_2_main_b28b7af69320201d1cf206ebf2837398", main)

	repo.GitBranch = "prod"
	prod := repo.GetCheckoutDirName(2)

	assert.NotEqual(t, main, prod)

	repo.GitBranch = "main"
	assert.Equal(t, main, repo.GetCheckoutDirName(2))

	repo.GitBranch = "feature/login"
	assert.NotContains(t, repo.GetCheckoutDirName(2), "/")

	slashed := repo.GetCheckoutDirName(2)
	repo.GitBranch = "feature-login"
	assert.NotEqual(t, slashed, repo.GetCheckoutDirName(2))
}
