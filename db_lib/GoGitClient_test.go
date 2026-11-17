package db_lib

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go-git builds its auth method from the key type; a key it can not use must
// fail the operation before any network call, not fall through to an anonymous
// clone. A string key is a token, which only an HTTP(S) server takes.
func TestGoGitClient_GetAuthMethod_RejectsUnusableKeys(t *testing.T) {
	tests := []struct {
		name    string
		gitURL  string
		keyType db.AccessKeyType
	}{
		{"unknown type", "https://git.example.com/ops.git", db.AccessKeyType("token")},
		{"string key over ssh://", "ssh://git@git.example.com/ops.git", db.AccessKeyString},
		{"string key over scp-style ssh", "git@git.example.com:ops.git", db.AccessKeyString},
		{"string key over git://", "git://git.example.com/ops.git", db.AccessKeyString},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := GitRepository{Repository: db.Repository{
				GitURL: tt.gitURL,
				SSHKey: db.AccessKey{Name: "k", Type: tt.keyType, String: "s3cr3t"},
			}}

			auth, err := GoGitClient{}.getAuthMethod(repo)

			require.Error(t, err)
			assert.Nil(t, auth)
			assert.Contains(t, err.Error(), "unsupported auth method")
		})
	}
}

// go-git sends an empty username as it is, and GitHub refuses a token that
// comes with one, so a string key always travels with the token login.
func TestGoGitClient_GetAuthMethod_StringKeyIsTokenOverHTTP(t *testing.T) {
	for _, gitURL := range []string{
		"https://git.example.com/ops.git",
		"http://git.example.com/ops.git",
		"HTTPS://git.example.com/ops.git",
	} {
		t.Run(gitURL, func(t *testing.T) {
			repo := GitRepository{Repository: db.Repository{
				GitURL: gitURL,
				SSHKey: db.AccessKey{Name: "k", Type: db.AccessKeyString, String: "s3cr3t"},
			}}

			auth, err := GoGitClient{}.getAuthMethod(repo)

			require.NoError(t, err)
			assert.Equal(t, &http.BasicAuth{Username: db.GitTokenLogin, Password: "s3cr3t"}, auth)
		})
	}
}

func TestGoGitClient_GetAuthMethod_NoneKeyIsAnonymous(t *testing.T) {
	repo := GitRepository{Repository: db.Repository{
		GitURL: "https://git.example.com/ops.git",
		SSHKey: db.AccessKey{Name: "k", Type: db.AccessKeyNone},
	}}

	auth, err := GoGitClient{}.getAuthMethod(repo)

	require.NoError(t, err)
	assert.Nil(t, auth)
}

// Galaxy clones roles from the repository's own server with the same token
// the repository itself is cloned with.
func TestGalaxyGitEnv_StringKeyOverHTTP(t *testing.T) {
	repo := db.Repository{
		GitURL: "https://git.example.com/ops.git",
		SSHKey: db.AccessKey{Name: "k", Type: db.AccessKeyString, String: "s3cr3t"},
	}

	env := galaxyGitEnv(repo)

	require.Len(t, env, 2)
	assert.Equal(t,
		`GIT_CONFIG_PARAMETERS='url.https://oauth2:s3cr3t@git.example.com/.insteadOf=https://git.example.com/'`,
		env[1])
}

// A string key means nothing to an SSH server, so the environment stays untouched.
func TestGalaxyGitEnv_StringKeyOverSSHAddsNoCredentials(t *testing.T) {
	repo := db.Repository{
		GitURL: "git@git.example.com:ops.git",
		SSHKey: db.AccessKey{Name: "k", Type: db.AccessKeyString, String: "s3cr3t"},
	}

	env := galaxyGitEnv(repo)

	assert.Equal(t, []string{"GIT_TERMINAL_PROMPT=0"}, env)
}
