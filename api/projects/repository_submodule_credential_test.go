package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	proServer "github.com/semaphoreui/semaphore/pro/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSubmoduleCredentialTest creates a real (in-memory) store with a
// project, an access key, and a repository owned by that key, plus a second
// project with its own access key -- used to prove a submodule credential
// can't be created pointing at a key from a different project (IDOR).
func setupSubmoduleCredentialTest(t *testing.T) (store db.Store, repository db.Repository, foreignAccessKey db.AccessKey) {
	t.Helper()

	s := sql.InitConfigCreateTestStore()

	project, err := s.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)

	key, err := s.CreateAccessKey(db.AccessKey{
		Name:      "main key",
		Type:      db.AccessKeyNone,
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	repository, err = s.CreateRepository(db.Repository{
		Name:      "test repo",
		ProjectID: project.ID,
		GitURL:    "https://gitserver/group/main.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	otherProject, err := s.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	foreignAccessKey, err = s.CreateAccessKey(db.AccessKey{
		Name:      "foreign key",
		Type:      db.AccessKeyNone,
		ProjectID: &otherProject.ID,
	})
	require.NoError(t, err)

	return s, repository, foreignAccessKey
}

func newSubmoduleCredentialRequest(method, url string, store db.Store, repository db.Repository, user *db.User, body string) (*http.Request, *httptest.ResponseRecorder) {
	var reqBody *strings.Reader
	if body == "" {
		reqBody = strings.NewReader("")
	} else {
		reqBody = strings.NewReader(body)
	}

	r := httptest.NewRequest(method, url, reqBody)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "repository", repository)
	r = helpers.SetContextValue(r, "user", user)
	r = helpers.SetContextValue(r, "log_writer", proServer.NewLogWriteService())

	return r, httptest.NewRecorder()
}

func TestAddRepositorySubmoduleCredential_CreatesMapping(t *testing.T) {
	store, repository, _ := setupSubmoduleCredentialTest(t)

	submoduleKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "submodule key",
		Type:      db.AccessKeyLoginPassword,
		ProjectID: &repository.ProjectID,
		LoginPassword: db.LoginPassword{
			Login:    "subuser",
			Password: "subpass",
		},
	})
	require.NoError(t, err)

	user := &db.User{ID: 1}
	body := `{"host":"gitserver","access_key_id":` + strconv.Itoa(submoduleKey.ID) + `}`
	r, w := newSubmoduleCredentialRequest(http.MethodPost,
		"/api/project/1/repositories/1/submodule_credentials", store, repository, user, body)

	AddRepositorySubmoduleCredential(w, r)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	creds, err := store.GetRepositorySubmoduleCredentials(repository.ProjectID, repository.ID)
	require.NoError(t, err)
	require.Len(t, creds, 1)
	assert.Equal(t, "gitserver", creds[0].Host)
	assert.Equal(t, submoduleKey.ID, creds[0].AccessKeyID)
}

func TestAddRepositorySubmoduleCredential_RejectsForeignProjectAccessKey(t *testing.T) {
	store, repository, foreignAccessKey := setupSubmoduleCredentialTest(t)

	user := &db.User{ID: 1}
	body := `{"host":"gitserver","access_key_id":` + strconv.Itoa(foreignAccessKey.ID) + `}`
	r, w := newSubmoduleCredentialRequest(http.MethodPost,
		"/api/project/1/repositories/1/submodule_credentials", store, repository, user, body)

	AddRepositorySubmoduleCredential(w, r)

	assert.NotEqual(t, http.StatusCreated, w.Code, "must not create a mapping to another project's access key")

	creds, err := store.GetRepositorySubmoduleCredentials(repository.ProjectID, repository.ID)
	require.NoError(t, err)
	assert.Empty(t, creds, "no mapping should have been persisted")
}

func TestAddRepositorySubmoduleCredential_RejectsInvalidHost(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{"url with scheme and path", "https://gitserver.example.com/org/repo"},
		{"path only", "gitserver/path"},
		{"leading whitespace", " gitserver"},
		{"trailing whitespace", "gitserver "},
		{"query string", "gitserver?query=1"},
		{"fragment", "gitserver#frag"},
		{"userinfo", "user@gitserver"},
		{"wildcard", "*"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, repository, _ := setupSubmoduleCredentialTest(t)

			submoduleKey, err := store.CreateAccessKey(db.AccessKey{
				Name:      "submodule key",
				Type:      db.AccessKeyNone,
				ProjectID: &repository.ProjectID,
			})
			require.NoError(t, err)

			body := `{"host":` + strconv.Quote(tt.host) + `,"access_key_id":` + strconv.Itoa(submoduleKey.ID) + `}`
			user := &db.User{ID: 1}
			r, w := newSubmoduleCredentialRequest(http.MethodPost,
				"/api/project/1/repositories/1/submodule_credentials", store, repository, user, body)

			AddRepositorySubmoduleCredential(w, r)

			assert.NotEqual(t, http.StatusCreated, w.Code, "must reject a host that can never match a submodule URL: %q", tt.host)

			creds, err := store.GetRepositorySubmoduleCredentials(repository.ProjectID, repository.ID)
			require.NoError(t, err)
			assert.Empty(t, creds, "no mapping should have been persisted")
		})
	}
}

func TestAddRepositorySubmoduleCredential_AcceptsHostWithPort(t *testing.T) {
	store, repository, _ := setupSubmoduleCredentialTest(t)

	submoduleKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "submodule key",
		Type:      db.AccessKeyNone,
		ProjectID: &repository.ProjectID,
	})
	require.NoError(t, err)

	user := &db.User{ID: 1}
	body := `{"host":"gitserver:8443","access_key_id":` + strconv.Itoa(submoduleKey.ID) + `}`
	r, w := newSubmoduleCredentialRequest(http.MethodPost,
		"/api/project/1/repositories/1/submodule_credentials", store, repository, user, body)

	AddRepositorySubmoduleCredential(w, r)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
}

func TestUpdateRepositorySubmoduleCredential_RejectsForeignProjectAccessKey(t *testing.T) {
	store, repository, foreignAccessKey := setupSubmoduleCredentialTest(t)

	submoduleKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "submodule key",
		Type:      db.AccessKeyNone,
		ProjectID: &repository.ProjectID,
	})
	require.NoError(t, err)

	cred, err := store.CreateRepositorySubmoduleCredential(db.RepositorySubmoduleCredential{
		ProjectID:    repository.ProjectID,
		RepositoryID: repository.ID,
		Host:         "gitserver",
		AccessKeyID:  submoduleKey.ID,
	})
	require.NoError(t, err)

	user := &db.User{ID: 1}
	body := `{"host":"gitserver","access_key_id":` + strconv.Itoa(foreignAccessKey.ID) + `}`
	r, w := newSubmoduleCredentialRequest(http.MethodPut,
		"/api/project/1/repositories/1/submodule_credentials/1", store, repository, user, body)
	r = helpers.SetContextValue(r, "submoduleCredential", cred)

	UpdateRepositorySubmoduleCredential(w, r)

	assert.NotEqual(t, http.StatusNoContent, w.Code, "must not repoint a mapping at another project's access key")

	got, err := store.GetRepositorySubmoduleCredential(repository.ProjectID, repository.ID, cred.ID)
	require.NoError(t, err)
	assert.Equal(t, submoduleKey.ID, got.AccessKeyID, "the original access key must be unchanged")
}

func TestRemoveRepositorySubmoduleCredential_Deletes(t *testing.T) {
	store, repository, _ := setupSubmoduleCredentialTest(t)

	submoduleKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "submodule key",
		Type:      db.AccessKeyNone,
		ProjectID: &repository.ProjectID,
	})
	require.NoError(t, err)

	cred, err := store.CreateRepositorySubmoduleCredential(db.RepositorySubmoduleCredential{
		ProjectID:    repository.ProjectID,
		RepositoryID: repository.ID,
		Host:         "gitserver",
		AccessKeyID:  submoduleKey.ID,
	})
	require.NoError(t, err)

	user := &db.User{ID: 1}
	r, w := newSubmoduleCredentialRequest(http.MethodDelete,
		"/api/project/1/repositories/1/submodule_credentials/1", store, repository, user, "")
	r = helpers.SetContextValue(r, "submoduleCredential", cred)

	RemoveRepositorySubmoduleCredential(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

	creds, err := store.GetRepositorySubmoduleCredentials(repository.ProjectID, repository.ID)
	require.NoError(t, err)
	assert.Empty(t, creds)
}

func TestGetRepositorySubmoduleCredentials_ListsMappings(t *testing.T) {
	store, repository, _ := setupSubmoduleCredentialTest(t)

	submoduleKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "submodule key",
		Type:      db.AccessKeyNone,
		ProjectID: &repository.ProjectID,
	})
	require.NoError(t, err)

	_, err = store.CreateRepositorySubmoduleCredential(db.RepositorySubmoduleCredential{
		ProjectID:    repository.ProjectID,
		RepositoryID: repository.ID,
		Host:         "gitserver",
		AccessKeyID:  submoduleKey.ID,
	})
	require.NoError(t, err)

	user := &db.User{ID: 1}
	r, w := newSubmoduleCredentialRequest(http.MethodGet,
		"/api/project/1/repositories/1/submodule_credentials", store, repository, user, "")

	GetRepositorySubmoduleCredentials(w, r)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "gitserver")
}
