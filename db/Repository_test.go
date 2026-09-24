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

func TestRepository_GetType_CaseInsensitiveScheme(t *testing.T) {
	tests := []struct {
		name     string
		gitURL   string
		expected RepositoryType
	}{
		{"lowercase https", "https://example.com/hello/world", RepositoryHTTP},
		{"uppercase HTTPS", "HTTPS://example.com/hello/world", RepositoryHTTP},
		{"mixed case Https", "Https://example.com/hello/world", RepositoryHTTP},
		{"uppercase HTTP", "HTTP://example.com/hello/world", RepositoryHTTP},
		{"uppercase SSH", "SSH://git@example.com/hello/world", RepositoryType("ssh")},
		{"scp-style ssh", "git@example.com:hello/world.git", RepositorySSH},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Repository{GitURL: tt.gitURL}.GetType())
		})
	}
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
		name             string
		Repository       Repository
		EmbedCredentials bool
		ExpectedGitUrl   string
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
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://login:password@github.com/user/project.git",
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
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://password@github.com/user/project.git",
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
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://user%40domain.com:pass%23word%40123@devops.domain.com/tfs/project/_git/repo",
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
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://user:p%40ss%3Aw%25rd+1&2%3F3@devops.domain.com/tfs/project/_git/repo",
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
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://token%25with%23special%40chars@devops.domain.com/tfs/project/_git/repo",
		},
		{
			name: "HTTPS embedCredentials=false does not embed the access key",
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
			EmbedCredentials: false,
			ExpectedGitUrl:   "https://devops.domain.com/tfs/project/_git/repo",
		},
		{
			// go_git clones with GetGitURL(false) and takes basic auth from the
			// URL's userinfo when no access key is set, so it must be kept.
			name: "HTTPS embedCredentials=false keeps userinfo typed into the URL",
			Repository: Repository{
				GitURL: "https://TOKEN@github.com/user/project.git",
			},
			EmbedCredentials: false,
			ExpectedGitUrl:   "https://TOKEN@github.com/user/project.git",
		},
		{
			name: "HTTPS without access key keeps userinfo typed into the URL",
			Repository: Repository{
				GitURL: "https://user:secret@devops.domain.com:8443/tfs/project/_git/repo",
			},
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://user:secret@devops.domain.com:8443/tfs/project/_git/repo",
		},
		{
			name: "SSH URL is returned as-is",
			Repository: Repository{
				GitURL: "git@github.com:user/project.git",
			},
			EmbedCredentials: false,
			ExpectedGitUrl:   "git@github.com:user/project.git",
		},
		{
			name: "Local path is returned as-is",
			Repository: Repository{
				GitURL: "/tmp/local/repo",
			},
			EmbedCredentials: true,
			ExpectedGitUrl:   "/tmp/local/repo",
		},
		{
			// Plain http keeps working as it always has. It logs a warning about
			// the cleartext transport, but silently dropping the credentials would
			// break existing installations with an opaque authentication error.
			name: "Plain HTTP still embeds credentials, encoded",
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
			EmbedCredentials: true,
			ExpectedGitUrl:   "http://user%40domain.com:secretpassword@insecure-git.domain.local/project/_git/repo",
		},
		{
			// URL schemes are case-insensitive and ValidateGitURL accepts any
			// spelling, so an uppercase scheme must still receive credentials.
			name: "Uppercase HTTPS scheme still embeds credentials",
			Repository: Repository{
				GitURL: "HTTPS://devops.domain.com/tfs/project/_git/repo",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "login",
						Password: "password",
					},
				},
			},
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://login:password@devops.domain.com/tfs/project/_git/repo",
		},
		{
			name: "No access key leaves the URL untouched",
			Repository: Repository{
				GitURL: "https://github.com/user/project.git",
				SSHKey: AccessKey{Type: AccessKeySSH},
			},
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://github.com/user/project.git",
		},
		{
			// go-git returns the net/url parse error verbatim, and it quotes the
			// whole URL, so a malformed URL must never reach it with credentials.
			name: "Malformed HTTPS embedCredentials=false drops the userinfo typed into the URL",
			Repository: Repository{
				GitURL: "https://user:secret%zz@example.com/repo.git",
			},
			EmbedCredentials: false,
			ExpectedGitUrl:   "https://example.com/repo.git",
		},
		{
			name: "Malformed HTTPS embedCredentials=true drops the userinfo and does not embed the access key",
			Repository: Repository{
				GitURL: "https://user:secret%zz@example.com/repo.git",
				SSHKey: AccessKey{
					Type: AccessKeyLoginPassword,
					LoginPassword: LoginPassword{
						Login:    "login",
						Password: "password",
					},
				},
			},
			EmbedCredentials: true,
			ExpectedGitUrl:   "https://example.com/repo.git",
		},
		{
			name: "Malformed ssh:// URL drops the userinfo",
			Repository: Repository{
				GitURL: "ssh://git:secret%zz@example.com/repo.git",
			},
			EmbedCredentials: false,
			ExpectedGitUrl:   "ssh://example.com/repo.git",
		},
		{
			name: "Malformed URL without userinfo is returned unchanged",
			Repository: Repository{
				GitURL: "https://example.com/re%zzpo.git",
			},
			EmbedCredentials: false,
			ExpectedGitUrl:   "https://example.com/re%zzpo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitUrl := tt.Repository.GetGitURL(tt.EmbedCredentials)
			assert.Equal(t, tt.ExpectedGitUrl, gitUrl, "wrong gitUrl for scenario: %s", tt.name)
			assert.NotContains(t, gitUrl, "secret%zz")
		})
	}
}

func TestRepository_GetRedactedGitURL(t *testing.T) {
	tests := []struct {
		name     string
		gitURL   string
		expected string
	}{
		{"login and password", "https://user:secret@devops.domain.com:8443/tfs/_git/repo", "https://devops.domain.com:8443/tfs/_git/repo"},
		{"token only", "https://ghp_TOKEN@github.com/user/project.git", "https://github.com/user/project.git"},
		{"no userinfo", "https://github.com/user/project.git", "https://github.com/user/project.git"},
		{"plain http", "http://user:secret@git.local/repo.git", "http://git.local/repo.git"},
		{"ssh scheme", "ssh://git@github.com/user/project.git", "ssh://github.com/user/project.git"},
		{"scp-style ssh is unchanged", "git@github.com:user/project.git", "git@github.com:user/project.git"},
		{"local path is unchanged", "/tmp/local/repo", "/tmp/local/repo"},

		// Malformed credentials that net/url either rejects or misreads as the
		// host, path or fragment. None of them may reach the log.
		{"invalid percent escape", "https://user:secret%zz@example.com/repo.git", "https://example.com/repo.git"},
		{"slash in token", "https://tok/en@github.com/user/project.git", "https://github.com/user/project.git"},
		{"slash in password", "https://user:p/ss@github.com/user/project.git", "https://github.com/user/project.git"},
		{"hash in token", "https://tok#en@github.com/user/project.git", "https://github.com/user/project.git"},
		{"question mark in password", "https://user:p?ss@github.com/user/project.git", "https://github.com/user/project.git"},
		{"at sign in password", "https://user:p@ss@github.com/user/project.git", "https://github.com/user/project.git"},
		{"space in password", "https://user:pa ss@github.com/user/project.git", "https://github.com/user/project.git"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := Repository{GitURL: tt.gitURL}.GetRedactedGitURL()

			assert.Equal(t, tt.expected, redacted)
			assert.NotContains(t, redacted, "secret")
			assert.NotContains(t, redacted, "TOKEN")
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
