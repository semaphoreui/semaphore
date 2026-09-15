package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlDb_CreateRepository_PlainHTTPPasswordRejected(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "repo test project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "password key",
		Type:      db.AccessKeyLoginPassword,
		ProjectID: &project.ID,
		LoginPassword: db.LoginPassword{
			Login:    "user",
			Password: "secretpassword",
		},
	})
	require.NoError(t, err)

	// Direct store creation with lowercase http:// and password key must fail
	_, err = store.CreateRepository(db.Repository{
		Name:      "plain http repo",
		ProjectID: project.ID,
		GitURL:    "http://gitlab.local/user/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Direct store creation with uppercase HTTP:// and password key must also fail
	_, err = store.CreateRepository(db.Repository{
		Name:      "uppercase http repo",
		ProjectID: project.ID,
		GitURL:    "HTTP://gitlab.local/user/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Direct store creation with HTTPS and password key must succeed
	repo, err := store.CreateRepository(db.Repository{
		Name:      "https repo",
		ProjectID: project.ID,
		GitURL:    "https://gitlab.local/user/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)
	assert.NotZero(t, repo.ID)
}

func TestSqlDb_UpdateRepository_PlainHTTPPasswordRejected(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "repo update test project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "password key",
		Type:      db.AccessKeyLoginPassword,
		ProjectID: &project.ID,
		LoginPassword: db.LoginPassword{
			Login:    "user",
			Password: "secretpassword",
		},
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		Name:      "https repo",
		ProjectID: project.ID,
		GitURL:    "https://gitlab.local/user/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	// Updating to plain http:// must fail
	repo.GitURL = "http://gitlab.local/user/repo.git"
	err = store.UpdateRepository(repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Updating to uppercase HTTP:// must also fail
	repo.GitURL = "HTTP://gitlab.local/user/repo.git"
	err = store.UpdateRepository(repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")

	// Updating to HTTPS must succeed
	repo.GitURL = "HTTPS://gitlab.local/user/repo.git"
	require.NoError(t, store.UpdateRepository(repo))
}

func TestSqlDb_CreateRepository_EmbeddedCredentialsPlainHTTPRejected(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "embedded repo test project"})
	require.NoError(t, err)

	noneKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "none key",
		Type:      db.AccessKeyNone,
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	// Direct creation of http:// with embedded userinfo must fail
	_, err = store.CreateRepository(db.Repository{
		Name:      "embedded plain http repo",
		ProjectID: project.ID,
		GitURL:    "http://user:secret@gitlab.local/user/repo.git",
		GitBranch: "main",
		SSHKeyID:  noneKey.ID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication is not supported over plain HTTP")
}

func TestValidateRepository_DoesNotMutateRepo(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "mutation test project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "test key",
		Type:      db.AccessKeyLoginPassword,
		ProjectID: &project.ID,
		LoginPassword: db.LoginPassword{
			Login:    "user",
			Password: "secretpassword",
		},
	})
	require.NoError(t, err)

	repo := db.Repository{
		Name:      "valid repo",
		ProjectID: project.ID,
		GitURL:    "https://gitlab.local/user/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	}

	// Before validation, repo.SSHKey is empty
	assert.Equal(t, db.AccessKey{}, repo.SSHKey)

	err = db.ValidateRepository(store, &repo)
	require.NoError(t, err)

	// After validation, repo.SSHKey must still be empty (no mutation)
	assert.Equal(t, db.AccessKey{}, repo.SSHKey)
}
