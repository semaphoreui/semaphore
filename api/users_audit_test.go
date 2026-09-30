package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func userRequest(store db.Store, method string, path string, body string, editor db.User, target db.User) (*http.Request, *audittest.Recorder) {
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "_user", target)
	return withAuditRecorder(asActor(r, editor))
}

func makeAdmin(t *testing.T, store db.Store, user db.User) db.User {
	t.Helper()
	user.Admin = true
	require.NoError(t, store.UpdateUser(db.UserWithPwd{User: user}))
	return user
}

func eventJSON(t *testing.T, recorded audittest.Recorded) string {
	t.Helper()
	raw, err := json.Marshal(recorded.Event)
	require.NoError(t, err)
	return string(raw)
}

func TestAddUser_IsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
	r, rec := userRequest(store, http.MethodPost, "/api/users",
		`{"username":"carol","name":"Carol","email":"carol@example.com","password":"verystrongpassword1"}`, admin, db.User{})
	w := httptest.NewRecorder()

	NewUsersController(nil).AddUser(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	var created db.User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	got := onlyEvent(t, rec, audit.IAMUserCreate)
	assert.Equal(t, audit.UserTarget(created.ID, "carol"), got.Event.Target)
	assert.Equal(t, audit.UserCreateMetadata{}, got.Event.Metadata)
	assert.Equal(t, "admin", got.Actor.Name)
}

func TestUpdateUser_RecordsChangedFields(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
	target := createUserOptionsTestUser(t, store, "bob")
	r, rec := userRequest(store, http.MethodPut, "/api/users/2",
		`{"username":"bob","name":"Robert","email":"bob@example.com","admin":true}`, admin, target)

	NewUsersController(nil).UpdateUser(httptest.NewRecorder(), r)

	assert.Equal(t, audit.UserUpdateMetadata{Fields: []string{"name", "admin"}, Admin: &audit.BoolChange{Old: false, New: true}},
		onlyEvent(t, rec, audit.IAMUserUpdate).Event.Metadata)
}

func TestUpdateUser_PasswordInBodyIsAdminReset(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
	target := createUserOptionsTestUser(t, store, "bob")
	r, rec := userRequest(store, http.MethodPut, "/api/users/2",
		`{"username":"bob","name":"bob","email":"bob@example.com","password":"anotherpassword2"}`, admin, target)

	NewUsersController(nil).UpdateUser(httptest.NewRecorder(), r)

	kinds, err := rec.Kinds()
	require.NoError(t, err)
	assert.Equal(t, []audit.Kind{audit.IAMUserUpdate, audit.IAMUserPasswordAdminReset}, kinds)
	assert.NotContains(t, eventJSON(t, rec.All()[0]), "anotherpassword2")
}

func TestUpdateUserPassword_IsRecorded(t *testing.T) {
	tests := []struct {
		name       string
		self       bool
		body       string
		wantKind   audit.Kind
		wantReason audit.Reason
	}{
		{"self change", true, `{"current_password":"verystrongpassword1","password":"newpassword2"}`, audit.IAMUserPasswordChange, audit.ReasonNone},
		{"wrong current password", true, `{"current_password":"wrong","password":"newpassword2"}`, audit.IAMUserPasswordChange, audit.ReasonInvalidCurrentPassword},
		{"admin reset", false, `{"password":"resetbyanadmin1"}`, audit.IAMUserPasswordAdminReset, audit.ReasonNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := setupSessionTest(t)
			target := createUserOptionsTestUser(t, store, "bob")
			editor := target
			if !tt.self {
				editor = makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
			}
			r, rec := userRequest(store, http.MethodPost, "/api/users/1/password", tt.body, editor, target)

			NewUsersController(nil).UpdateUserPassword(httptest.NewRecorder(), r)

			got := onlyEvent(t, rec, tt.wantKind)
			assert.Equal(t, tt.wantReason, got.Event.Reason)
			assert.Equal(t, audit.UserTarget(target.ID, "bob"), got.Event.Target)
			assert.NotContains(t, eventJSON(t, got), "newpassword2")
			assert.NotContains(t, eventJSON(t, got), "resetbyanadmin1")
		})
	}
}

func TestDeleteUser_IsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
	target := createUserOptionsTestUser(t, store, "bob")
	r, rec := userRequest(store, http.MethodDelete, "/api/users/2", "", admin, target)

	NewUsersController(nil).DeleteUser(httptest.NewRecorder(), r)

	assert.Equal(t, audit.UserTarget(target.ID, "bob"), onlyEvent(t, rec, audit.IAMUserDelete).Event.Target)
}

func TestTotpLifecycle_IsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, rec := userRequest(store, http.MethodPost, "/api/users/1/2fas/totp", "", user, user)
	NewUsersController(nil).EnableTotp(httptest.NewRecorder(), r)
	onlyEvent(t, rec, audit.IAMMFAEnable)

	user, err := store.GetUser(user.ID)
	require.NoError(t, err)
	require.NotNil(t, user.Totp)

	r, rec = userRequest(store, http.MethodGet, "/api/users/1/2fas/totp/1/qr", "", user, user)
	NewUsersController(nil).TotpQr(httptest.NewRecorder(), r)
	onlyEvent(t, rec, audit.IAMMFAViewQR)

	r, rec = userRequest(store, http.MethodDelete, "/api/users/1/2fas/totp/1", "", user, user)
	NewUsersController(nil).DisableTotp(httptest.NewRecorder(), mux.SetURLVars(r, map[string]string{"totp_id": "1"}))
	onlyEvent(t, rec, audit.IAMMFADisable)
}

func TestDeleteUserIdentity_IsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	_, err := store.CreateExternalIdentity(db.UserExternalIdentity{UserID: user.ID, Type: db.IdentityTypeOidc, Provider: "corp", ExternalUID: "sub-1"})
	require.NoError(t, err)

	r, rec := userRequest(store, http.MethodDelete, "/api/users/1/identities/oidc/corp", "", user, user)
	NewUsersController(nil).DeleteUserIdentity(httptest.NewRecorder(), mux.SetURLVars(r, map[string]string{"type": "oidc", "provider": "corp"}))
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodOIDC, Provider: "corp"}, onlyEvent(t, rec, audit.IAMExternalIdentityUnlink).Event.Metadata)

	r, rec = userRequest(store, http.MethodDelete, "/api/users/1/identities/oidc/none", "", user, user)
	NewUsersController(nil).DeleteUserIdentity(httptest.NewRecorder(), mux.SetURLVars(r, map[string]string{"type": "oidc", "provider": "none"}))
	assert.Empty(t, rec.All())
}

func TestAPITokens_AreRecordedByFingerprint(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, rec := userRequest(store, http.MethodPost, "/api/user/tokens", `{"name":"ci"}`, user, user)
	w := httptest.NewRecorder()
	createAPIToken(w, r)

	var token db.APIToken
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &token))
	got := onlyEvent(t, rec, audit.IAMAPITokenCreate)
	assert.Equal(t, &audit.Target{Type: audit.TargetAPIToken, ID: audit.TokenFingerprint(token.ID), Name: "ci"}, got.Event.Target)
	assert.NotContains(t, eventJSON(t, got), token.ID)

	r, rec = userRequest(store, http.MethodDelete, "/api/user/tokens/x", "", user, user)
	deleteAPIToken(httptest.NewRecorder(), mux.SetURLVars(r, map[string]string{"token_id": token.ID[:8]}))
	assert.Equal(t, audit.TokenFingerprint(token.ID), onlyEvent(t, rec, audit.IAMAPITokenDelete).Event.Target.ID)
}

func TestDeleteAPIToken_RecordsEveryTokenTheStoreDeleted(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		deleted []string
	}{
		// SQLite and MySQL compare LIKE without case.
		{name: "upper case prefix", prefix: "ABCDEFGH", deleted: []string{"abcdefgh1111"}},
		// "_" matches any character in LIKE.
		{name: "underscore prefix", prefix: "abcdefg_", deleted: []string{"abcdefgh1111", "abcdefgx2222"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := setupSessionTest(t)
			user := createUserOptionsTestUser(t, store, "alice")
			for _, id := range []string{"abcdefgh1111", "abcdefgx2222", "zzzzzzzz3333"} {
				_, err := store.CreateAPIToken(db.APIToken{ID: id, UserID: user.ID, Name: id})
				require.NoError(t, err)
			}

			r, rec := userRequest(store, http.MethodDelete, "/api/user/tokens/x", "", user, user)
			deleteAPIToken(httptest.NewRecorder(), mux.SetURLVars(r, map[string]string{"token_id": tt.prefix}))

			var recorded []string
			for _, got := range rec.All() {
				assert.Equal(t, audit.IAMAPITokenDelete, got.Event.Kind)
				recorded = append(recorded, got.Event.Target.Name)
			}
			assert.ElementsMatch(t, tt.deleted, recorded)
		})
	}
}

// concurrentTokenDelete reports one deleted token while the rest of the prefix was removed by another request.
type concurrentTokenDelete struct {
	db.Store
}

func (concurrentTokenDelete) DeleteAPIToken(int, string) ([]db.APIToken, error) {
	return []db.APIToken{{ID: "abcdefgh1111", Name: "mine"}}, nil
}

func TestDeleteAPIToken_RecordsOnlyWhatThisRequestDeleted(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")

	r, rec := userRequest(concurrentTokenDelete{Store: store}, http.MethodDelete, "/api/user/tokens/x", "", user, user)
	w := httptest.NewRecorder()
	deleteAPIToken(w, mux.SetURLVars(r, map[string]string{"token_id": "abcdefgh"}))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "mine", onlyEvent(t, rec, audit.IAMAPITokenDelete).Event.Target.Name)
}

func TestSetOption_RecordsKeyOnly(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
	r, rec := userRequest(store, http.MethodPost, "/api/options", `{"key":"some_key","value":"secret-value"}`, admin, db.User{})

	setOption(httptest.NewRecorder(), r)

	got := onlyEvent(t, rec, audit.SystemSettingsUpdate)
	assert.Equal(t, audit.SettingsMetadata{Keys: []string{"some_key"}}, got.Event.Metadata)
	assert.NotContains(t, eventJSON(t, got), "secret-value")
}

func TestUserUpdateMetadata(t *testing.T) {
	before := db.User{Username: "a", Name: "A", Email: "a@x", Pro: true}
	after := db.UserWithPwd{User: db.User{Username: "b", Name: "A", Email: "b@x", Alert: true, Admin: true}, Pwd: "p"}
	assert.Equal(t, audit.UserUpdateMetadata{
		Fields: []string{"username", "email", "alert", "admin", "pro", "password"},
		Admin:  &audit.BoolChange{Old: false, New: true},
		Pro:    &audit.BoolChange{Old: true, New: false},
	}, userUpdateMetadata(before, after))
	assert.Equal(t, audit.UserUpdateMetadata{Fields: []string{}}, userUpdateMetadata(before, db.UserWithPwd{User: before}))
}

func TestUsers_SelfEscalationIsDenied(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		body    string
		self    bool
		handler func(*UsersController, http.ResponseWriter, *http.Request)
	}{
		{"create a user", http.MethodPost, `{"username":"carol","name":"Carol","email":"carol@example.com","password":"verystrongpassword1","admin":true}`, false, (*UsersController).AddUser},
		{"make self admin", http.MethodPut, `{"username":"bob","name":"bob","email":"bob@example.com","admin":true}`, true, (*UsersController).UpdateUser},
		{"make self pro", http.MethodPut, `{"username":"bob","name":"bob","email":"bob@example.com","pro":true}`, true, (*UsersController).UpdateUser},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := setupSessionTest(t)
			bob := createUserOptionsTestUser(t, store, "bob")
			target := db.User{}
			if tt.self {
				target = bob
			}
			r, rec := userRequest(store, tt.method, "/api/users", tt.body, bob, target)
			w := httptest.NewRecorder()

			tt.handler(NewUsersController(nil), w, r)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "the response does not change")
			got := onlyEvent(t, rec, audit.AuthAuthorizationDeny)
			assert.Equal(t, audit.DenyMetadata{Method: tt.method, Permission: "admin"}, got.Event.Metadata)
			assert.Equal(t, "bob", got.Actor.Name)
		})
	}
}

func TestUpdateUser_EventsNameTheUserAlike(t *testing.T) {
	store := setupSessionTest(t)
	admin := makeAdmin(t, store, createUserOptionsTestUser(t, store, "admin"))
	target := createUserOptionsTestUser(t, store, "bob")
	r, rec := userRequest(store, http.MethodPut, "/api/users/2",
		`{"username":"robert","name":"bob","email":"bob@example.com","password":"anotherpassword2"}`, admin, target)

	NewUsersController(nil).UpdateUser(httptest.NewRecorder(), r)

	require.Len(t, rec.All(), 2)
	for _, recorded := range rec.All() {
		assert.Equal(t, audit.UserTarget(target.ID, "robert"), recorded.Event.Target)
	}
}
