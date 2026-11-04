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

// The pair auth method / credential type is checked when the integration is
// stored, so a webhook never fails at delivery time for a key that could not
// have worked.
func TestAddIntegration_AuthSecretMustFitTheAuthMethod(t *testing.T) {
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
			body := `{"name":"hook","project_id":` + strconv.Itoa(f.project.ID) +
				`,"template_id":` + strconv.Itoa(f.template(t).ID) +
				`,"auth_method":"` + tt.method + `","auth_header":"X-Token","auth_secret_id":` + strconv.Itoa(key.ID) + `}`
			r, _ := f.request(http.MethodPost, body, nil, nil)
			w := httptest.NewRecorder()

			integrationController(noFeatures).AddIntegration(w, r)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			integrations, err := f.store.GetIntegrations(f.project.ID, db.RetrieveQueryParams{}, false)
			require.NoError(t, err)
			if tt.wantCode == http.StatusBadRequest {
				assert.Contains(t, w.Body.String(), "Auth secret type does not fit")
				assert.Empty(t, integrations, "nothing is created")
				return
			}
			require.Len(t, integrations, 1)
			require.NotNil(t, integrations[0].AuthSecretID)
			assert.Equal(t, key.ID, *integrations[0].AuthSecretID)
		})
	}
}

func TestAddIntegration_AuthSecretOfAnotherProjectIsRefused(t *testing.T) {
	f := newResourceFixture(t)
	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	key, err := f.store.CreateAccessKey(db.AccessKey{Name: "hook", Type: db.AccessKeyString, ProjectID: &other.ID})
	require.NoError(t, err)
	body := `{"name":"hook","project_id":` + strconv.Itoa(f.project.ID) +
		`,"template_id":` + strconv.Itoa(f.template(t).ID) +
		`,"auth_method":"token","auth_header":"X-Token","auth_secret_id":` + strconv.Itoa(key.ID) + `}`
	r, _ := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()

	integrationController(noFeatures).AddIntegration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "does not exist in this project")
}

// Switching an integration to BasicAuth while it keeps a string key is the
// pair the form drops on its own; the API refuses it when sent directly.
func TestUpdateIntegration_AuthSecretMustFitTheAuthMethod(t *testing.T) {
	f := newResourceFixture(t)
	stringKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "hook", Type: db.AccessKeyString, ProjectID: &f.project.ID})
	require.NoError(t, err)
	loginKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "bob", Type: db.AccessKeyLoginPassword, ProjectID: &f.project.ID})
	require.NoError(t, err)
	template := f.template(t)
	stored, err := f.store.CreateIntegration(db.Integration{
		Name: "hook", ProjectID: f.project.ID, TemplateID: &template.ID,
		AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token", AuthSecretID: &stringKey.ID,
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
			body := `{"id":` + strconv.Itoa(stored.ID) + `,"name":"hook","project_id":` + strconv.Itoa(f.project.ID) +
				`,"template_id":` + strconv.Itoa(template.ID) +
				`,"auth_method":"` + tt.method + `","auth_header":"X-Token","auth_secret_id":` + strconv.Itoa(tt.keyID) + `}`
			r, _ := f.request(http.MethodPut, body, map[string]any{"integration": stored}, nil)
			w := httptest.NewRecorder()

			integrationController(noFeatures).UpdateIntegration(w, r)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
		})
	}
}
