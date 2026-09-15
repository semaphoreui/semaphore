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
	tests := []struct {
		name           string
		Repository     Repository
		Secure         bool
		ExpectedGitUrl string
	}{
		{
			name: "HTTPS login+password credentials embedded",
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
			name: "HTTPS token-only (no login) embedded as user",
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
			name: "HTTPS special chars in login and password are RFC 3986 encoded",
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
			name: "HTTPS password with multiple special chars encoded correctly",
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
			name: "HTTPS token with percent and hash encoded",
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
			name: "HTTPS secure=true strips embedded credentials",
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
			name: "HTTPS with userinfo in URL secure=true strips userinfo",
			Repository: Repository{
				GitURL: "https://user:secret@devops.domain.com:8443/tfs/project/_git/repo",
			},
			Secure:         true,
			ExpectedGitUrl: "https://devops.domain.com:8443/tfs/project/_git/repo",
		},
		{
			name: "SSH URL is returned as-is",
			Repository: Repository{
				GitURL: "git@github.com:user/project.git",
			},
			Secure:         true,
			ExpectedGitUrl: "git@github.com:user/project.git",
		},
		{
			name: "Local path is returned as-is",
			Repository: Repository{
				GitURL: "/tmp/local/repo",
			},
			Secure:         false,
			ExpectedGitUrl: "/tmp/local/repo",
		},
		{
			name: "Plain HTTP credentials NOT embedded (CWE-319)",
			Repository: Repository{
				GitURL: "http://insecure-git.domain.local/project/_git/repo",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "user@domain.com",
						Password: "secretpassword",
					},
				},
			},
			Secure:         false,
			ExpectedGitUrl: "http://insecure-git.domain.local/project/_git/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitUrl := tt.Repository.GetGitURL(tt.Secure)
			assert.Equal(t, tt.ExpectedGitUrl, gitUrl, "wrong gitUrl for scenario: %s", tt.name)
		})
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
