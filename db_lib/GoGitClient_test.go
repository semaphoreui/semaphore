package db_lib

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go-git builds its auth method from the key type; a key type it does not
// know must fail the operation before any network call, not fall through to an
// anonymous clone.
func TestGoGitClient_GetAuthMethod_RejectsUnknownKeyTypes(t *testing.T) {
	tests := []struct {
		name    string
		keyType db.AccessKeyType
	}{
		{"string key", db.AccessKeyString},
		{"unknown type", db.AccessKeyType("token")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := GitRepository{Repository: db.Repository{
				GitURL: "https://git.example.com/ops.git",
				SSHKey: db.AccessKey{Name: "k", Type: tt.keyType, String: "s3cr3t"},
			}}

			auth, err := GoGitClient{}.getAuthMethod(repo)

			require.Error(t, err)
			assert.Nil(t, auth)
			assert.Contains(t, err.Error(), "unsupported auth method")
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

// Galaxy credentials are offered only for a login/password key over HTTP; a
// string key over the same URL must leave the environment untouched.
func TestGalaxyGitEnv_StringKeyAddsNoCredentials(t *testing.T) {
	repo := db.Repository{
		GitURL: "https://git.example.com/ops.git",
		SSHKey: db.AccessKey{Name: "k", Type: db.AccessKeyString, String: "s3cr3t"},
	}

	env := galaxyGitEnv(repo)

	assert.Equal(t, []string{"GIT_TERMINAL_PROMPT=0"}, env)
}
