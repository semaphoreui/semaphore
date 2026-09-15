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
		{
			Repository: Repository{
				GitURL: "HTTPS://github.com/user/project.git",
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
				GitURL: "HTTP://insecure.local/user/project.git",
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
				GitURL: "HTTPS://user:secret@github.com/user/project.git",
			},
			Secure:         true,
			ExpectedGitUrl: "https://github.com/user/project.git",
		},
		{
			Repository: Repository{
				GitURL: "https://user:pass@github.com/user/project.git?access_token=secrettoken",
			},
			Secure:         true,
			ExpectedGitUrl: "https://github.com/user/project.git?access_token=***",
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

func TestRepository_ValidatePlainHTTPPassword(t *testing.T) {
	repo := Repository{
		Name:      "Insecure Repo",
		GitURL:    "http://insecure.local/user/project.git",
		GitBranch: "main",
		SSHKey: AccessKey{
			Type: AccessKeyLoginPassword,
			LoginPassword: LoginPassword{
				Login:    "user",
				Password: "secretpassword",
			},
		},
	}

	err := repo.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Uppercase HTTP scheme with password must also fail validation
	repo.GitURL = "HTTP://insecure.local/user/project.git"
	err = repo.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Mixed-case Http scheme with password must also fail validation
	repo.GitURL = "Http://insecure.local/user/project.git"
	err = repo.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// HTTPS repo with password should succeed validation
	repo.GitURL = "https://secure.local/user/project.git"
	assert.NoError(t, repo.Validate())

	// Uppercase HTTPS repo with password should succeed validation
	repo.GitURL = "HTTPS://secure.local/user/project.git"
	assert.NoError(t, repo.Validate())

	// Plain HTTP URL with embedded credentials must fail even without login_password key type
	embeddedRepo := Repository{
		Name:      "Embedded Credentials Repo",
		GitURL:    "http://user:secret@insecure.local/user/project.git",
		GitBranch: "main",
	}
	err = embeddedRepo.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Uppercase HTTP with embedded credentials must also fail
	embeddedRepo.GitURL = "HTTP://user:secret@insecure.local/user/project.git"
	err = embeddedRepo.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// HTTPS with embedded credentials is valid
	embeddedRepo.GitURL = "https://user:secret@secure.local/user/project.git"
	assert.NoError(t, embeddedRepo.Validate())
}

func TestRepository_GetTypeCaseInsensitive(t *testing.T) {
	assert.Equal(t, RepositoryHTTP, Repository{GitURL: "http://example.com/repo.git"}.GetType())
	assert.Equal(t, RepositoryHTTP, Repository{GitURL: "HTTP://example.com/repo.git"}.GetType())
	assert.Equal(t, RepositoryHTTP, Repository{GitURL: "Http://example.com/repo.git"}.GetType())
	assert.Equal(t, RepositoryHTTP, Repository{GitURL: "https://example.com/repo.git"}.GetType())
	assert.Equal(t, RepositoryHTTP, Repository{GitURL: "HTTPS://example.com/repo.git"}.GetType())
	assert.Equal(t, RepositoryHTTP, Repository{GitURL: "Https://example.com/repo.git"}.GetType())
}
