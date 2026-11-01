package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The pair auth method / credential type is checked when the endpoint is
// stored, so a webhook never fails at delivery time for a key that could not
// have worked.
func TestAddEndpoint_AuthSecretMustFitTheAuthMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		keyType  db.AccessKeyType
		owner    db.AccessKeyOwner
		wantCode int
	}{
		{"token with a string key", "token", db.AccessKeyString, db.AccessKeyShared, http.StatusCreated},
		{"hmac with a string key", "hmac", db.AccessKeyString, db.AccessKeyShared, http.StatusCreated},
		{"token with a login/password key", "token", db.AccessKeyLoginPassword, db.AccessKeyShared, http.StatusCreated},
		{"basic with a login/password key", "basic", db.AccessKeyLoginPassword, db.AccessKeyShared, http.StatusCreated},
		{"basic with a string key", "basic", db.AccessKeyString, db.AccessKeyShared, http.StatusBadRequest},
		{"token with an ssh key", "token", db.AccessKeySSH, db.AccessKeyShared, http.StatusBadRequest},
		{"token with a none key", "token", db.AccessKeyNone, db.AccessKeyShared, http.StatusBadRequest},
		{"token with a variable-group string key", "token", db.AccessKeyString, db.AccessKeyEnvironment, http.StatusBadRequest},
		{"no auth with any key", "none", db.AccessKeySSH, db.AccessKeyShared, http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			key, err := f.store.CreateAccessKey(db.AccessKey{Name: "hook", Type: tt.keyType, Owner: tt.owner, ProjectID: &f.project.ID})
			require.NoError(t, err)
			body := `{"name":"hooks","project_id":` + strconv.Itoa(f.project.ID) +
				`,"auth_method":"` + tt.method + `","auth_header":"X-Token","auth_secret_id":` + strconv.Itoa(key.ID) + `}`
			r, _ := f.request(http.MethodPost, body, nil, nil)
			w := httptest.NewRecorder()

			AddEndpoint(w, r)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			endpoints, err := f.store.GetEndpoints(f.project.ID, db.RetrieveQueryParams{})
			require.NoError(t, err)
			if tt.wantCode == http.StatusBadRequest {
				assert.Contains(t, w.Body.String(), "Auth secret type does not fit")
				assert.Empty(t, endpoints, "nothing is created")
				return
			}
			require.Len(t, endpoints, 1)
			require.NotNil(t, endpoints[0].AuthSecretID)
			assert.Equal(t, key.ID, *endpoints[0].AuthSecretID)
		})
	}
}

func TestAddEndpoint_AuthSecretOfAnotherProjectIsRefused(t *testing.T) {
	f := newResourceFixture(t)
	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	key, err := f.store.CreateAccessKey(db.AccessKey{Name: "hook", Type: db.AccessKeyString, ProjectID: &other.ID})
	require.NoError(t, err)
	body := `{"name":"hooks","project_id":` + strconv.Itoa(f.project.ID) +
		`,"auth_method":"token","auth_header":"X-Token","auth_secret_id":` + strconv.Itoa(key.ID) + `}`
	r, _ := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()

	AddEndpoint(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "does not exist in this project")
}

// Switching an endpoint to BasicAuth while it keeps a string key is the pair
// the form drops on its own; the API refuses it when sent directly.
func TestUpdateEndpoint_AuthSecretMustFitTheAuthMethod(t *testing.T) {
	f := newResourceFixture(t)
	stringKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "hook", Type: db.AccessKeyString, ProjectID: &f.project.ID})
	require.NoError(t, err)
	loginKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "bob", Type: db.AccessKeyLoginPassword, ProjectID: &f.project.ID})
	require.NoError(t, err)
	stored, err := f.store.CreateEndpoint(db.Endpoint{
		Name: "hooks", ProjectID: f.project.ID,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token", AuthSecretID: &stringKey.ID},
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		method   string
		keyID    int
		wantCode int
	}{
		{"basic keeps the string key", "basic", stringKey.ID, http.StatusBadRequest},
		{"basic with a login/password key", "basic", loginKey.ID, http.StatusNoContent},
		{"hmac keeps the string key", "hmac", stringKey.ID, http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"id":` + strconv.Itoa(stored.ID) + `,"name":"hooks","project_id":` + strconv.Itoa(f.project.ID) +
				`,"auth_method":"` + tt.method + `","auth_header":"X-Token","auth_secret_id":` + strconv.Itoa(tt.keyID) + `}`
			r, _ := f.request(http.MethodPut, body, map[string]any{"endpoint": stored}, nil)
			w := httptest.NewRecorder()

			UpdateEndpoint(w, r)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
		})
	}
}

// A group's own key is checked only when the group overrides the endpoint's
// auth; without the override the auth fields are dropped, not refused.
func TestAddIntegrationGroup_OwnAuthSecretMustFitWhenOverriding(t *testing.T) {
	f := newResourceFixture(t)
	endpoint := f.endpoint(t)
	stringKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "hook", Type: db.AccessKeyString, ProjectID: &f.project.ID})
	require.NoError(t, err)

	body := func(override bool) string {
		return `{"name":"deploys","project_id":` + strconv.Itoa(f.project.ID) + `,"endpoint_id":` + strconv.Itoa(endpoint.ID) +
			`,"auth_override":` + strconv.FormatBool(override) + `,"auth_method":"basic","auth_secret_id":` + strconv.Itoa(stringKey.ID) + `}`
	}

	r, _ := f.request(http.MethodPost, body(true), nil, nil)
	w := httptest.NewRecorder()
	AddIntegrationGroup(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "Auth secret type does not fit")

	r, _ = f.request(http.MethodPost, body(false), nil, nil)
	w = httptest.NewRecorder()
	AddIntegrationGroup(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	groups, err := f.store.GetIntegrationGroups(f.project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.False(t, groups[0].AuthOverride)
	assert.Equal(t, db.IntegrationAuth{}, groups[0].IntegrationAuth, "an inheriting group keeps no auth of its own")
}
