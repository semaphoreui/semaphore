package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newUpdateKeyRequest builds a request with the URL-resolved oldKey already in
// context (as KeyMiddleware would set it) and the given JSON body.
func newUpdateKeyRequest(oldKey db.AccessKey, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPut, "/api/project/1/keys/1", strings.NewReader(body))
	req = helpers.SetContextValue(req, "accessKey", oldKey)
	return req, httptest.NewRecorder()
}

func TestUpdateKey_RejectsBodyIDMismatch(t *testing.T) {

	svc := &mockAccessKeyService{}
	ctrl := NewKeyController(svc)

	oldKey := db.AccessKey{ID: 10, ProjectID: new(1)}
	// Body targets a different key id than the one resolved from the URL.
	body := `{"id":42,"name":"x","type":"none","project_id":1}`
	req, w := newUpdateKeyRequest(oldKey, body)

	ctrl.UpdateKey(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "must be the same")
	assert.Empty(t, svc.updated, "store must not be touched on rejection")
}

func TestUpdateKey_RejectsCrossProjectMove(t *testing.T) {
	svc := &mockAccessKeyService{}
	ctrl := NewKeyController(svc)

	oldKey := db.AccessKey{ID: 10, ProjectID: new(1)}
	// Body keeps the same key id but points project_id at a foreign project.
	body := `{"id":10,"name":"x","type":"none","project_id":999}`
	req, w := newUpdateKeyRequest(oldKey, body)

	ctrl.UpdateKey(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "other project")
	assert.Empty(t, svc.updated, "store must not be touched on rejection")
}

func TestUpdateKey_RejectsNilBodyProjectID(t *testing.T) {
	svc := &mockAccessKeyService{}
	ctrl := NewKeyController(svc)

	oldKey := db.AccessKey{ID: 10, ProjectID: new(1)}
	// Body omits project_id entirely.
	body := `{"id":10,"name":"x","type":"none"}`
	req, w := newUpdateKeyRequest(oldKey, body)

	ctrl.UpdateKey(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "other project")
	assert.Empty(t, svc.updated, "store must not be touched on rejection")
}

// newKeyFixture is the resource fixture with the real key service on top of
// the test store, so a stored secret is really encrypted and really reloaded.
func newKeyFixture(t *testing.T) (resourceFixture, *KeyController) {
	t.Helper()
	f := newResourceFixture(t)
	encryption := server.NewAccessKeyEncryptionService(f.store, f.store, f.store, f.store)
	return f, NewKeyController(server.NewAccessKeyService(f.store, encryption, f.store, f.store))
}

// storedSecret reads the ciphertext column of a key as the database holds it.
func storedSecret(t *testing.T, f resourceFixture, keyID int) string {
	t.Helper()
	key, err := f.store.GetAccessKey(f.project.ID, keyID)
	require.NoError(t, err)
	require.NotNil(t, key.Secret, "secret must be stored")
	return *key.Secret
}

// decryptedString loads the key the way a task would and returns its string value.
func decryptedString(t *testing.T, f resourceFixture, keyID int) string {
	t.Helper()
	key, err := f.store.GetAccessKey(f.project.ID, keyID)
	require.NoError(t, err)
	encryption := server.NewAccessKeyEncryptionService(f.store, f.store, f.store, f.store)
	require.NoError(t, encryption.DeserializeSecret(&key))
	return key.String
}

func TestAddKey_StringIsStoredEncryptedAndNeverReturned(t *testing.T) {
	f, ctrl := newKeyFixture(t)
	body := `{"name":"api-token","type":"string","string":"s3cr3t","project_id":` + strconv.Itoa(f.project.ID) + `}`
	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()

	ctrl.AddKey(w, r)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.AccessKey
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, db.AccessKeyString, created.Type)
	assert.Equal(t, db.AccessKeyShared, created.Owner)
	assert.Empty(t, created.String, "the value must not come back in the response")
	assert.NotContains(t, w.Body.String(), "s3cr3t")

	assert.NotEqual(t, "s3cr3t", storedSecret(t, f, created.ID), "the value must not be stored as plaintext")
	assert.Equal(t, "s3cr3t", decryptedString(t, f, created.ID))
	assert.Equal(t, audit.CredentialMetadata{Type: "string"}, only(t, rec, audit.SecretCredentialCreate).Event.Metadata)
}

func TestGetKeys_StringValueIsNotListed(t *testing.T) {
	f, ctrl := newKeyFixture(t)
	body := `{"name":"api-token","type":"string","string":"s3cr3t","project_id":` + strconv.Itoa(f.project.ID) + `}`
	r, _ := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	ctrl.AddKey(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	r, _ = f.request(http.MethodGet, "", nil, nil)
	w = httptest.NewRecorder()
	GetKeys(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"type":"string"`)
	assert.NotContains(t, w.Body.String(), "s3cr3t")
}

func TestAddKey_RejectsUnknownType(t *testing.T) {
	svc := &mockAccessKeyService{}
	ctrl := NewKeyController(svc)
	f := newResourceFixture(t)
	body := `{"name":"x","type":"password","project_id":` + strconv.Itoa(f.project.ID) + `}`
	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()

	ctrl.AddKey(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid key type")
	assert.Empty(t, rec.All(), "nothing was created, nothing to audit")
}

func TestUpdateKey_StringOverrideReplacesSecret(t *testing.T) {
	f, ctrl := newKeyFixture(t)
	body := `{"name":"api-token","type":"string","string":"old","project_id":` + strconv.Itoa(f.project.ID) + `}`
	r, _ := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	ctrl.AddKey(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created db.AccessKey
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	before := storedSecret(t, f, created.ID)

	tests := []struct {
		name       string
		body       string
		wantValue  string
		wantStored func(t *testing.T, after string)
	}{
		{
			name:      "without override only the name changes",
			body:      `"type":"string","string":"new"`,
			wantValue: "old",
			wantStored: func(t *testing.T, after string) {
				assert.Equal(t, before, after, "ciphertext must be untouched")
			},
		},
		{
			name:      "with override the value is replaced",
			body:      `"type":"string","string":"new","override_secret":true`,
			wantValue: "new",
			wantStored: func(t *testing.T, after string) {
				assert.NotEqual(t, before, after, "ciphertext must change")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored, err := f.store.GetAccessKey(f.project.ID, created.ID)
			require.NoError(t, err)
			body := `{"id":` + strconv.Itoa(created.ID) + `,"name":"renamed","project_id":` + strconv.Itoa(f.project.ID) + `,` + tt.body + `}`
			r, _ := f.request(http.MethodPut, body, map[string]any{"accessKey": stored}, nil)
			w := httptest.NewRecorder()

			ctrl.UpdateKey(w, r)

			require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
			tt.wantStored(t, storedSecret(t, f, created.ID))
			assert.Equal(t, tt.wantValue, decryptedString(t, f, created.ID))
		})
	}
}

func TestUpdateKey_RejectsUnknownType(t *testing.T) {
	svc := &mockAccessKeyService{}
	ctrl := NewKeyController(svc)

	oldKey := db.AccessKey{ID: 10, ProjectID: new(1), Type: db.AccessKeyNone}
	body := `{"id":10,"name":"x","type":"token","project_id":1,"override_secret":true}`
	req, w := newUpdateKeyRequest(oldKey, body)

	ctrl.UpdateKey(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid key type")
	assert.Empty(t, svc.updated, "store must not be touched on rejection")
}
