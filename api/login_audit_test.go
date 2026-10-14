package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loginRequest(store db.Store, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	return helpers.SetContextValue(r, "store", store)
}

func TestLogin_PasswordSuccessIsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	r, rec := withAuditRecorder(loginRequest(store, `{"auth":"alice","password":"verystrongpassword1","method":"password"}`))
	w := httptest.NewRecorder()

	login(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	got := onlyEvent(t, rec, audit.AuthLogin)
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthSession, ""), got.Actor)
	assert.Equal(t, audit.UserTarget(user.ID, "alice"), got.Event.Target)
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodPassword}, got.Event.Metadata)
	raw, err := json.Marshal(got.Event)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "verystrongpassword1")
}

func TestLogin_FailuresAreRecorded(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		disablePassword bool
		wantReason      audit.Reason
		wantName        string
	}{
		{"wrong password", `{"auth":"alice","password":"wrong","method":"password"}`, false, audit.ReasonInvalidCredentials, "alice"},
		{"unknown user", `{"auth":"nobody","password":"x","method":"password"}`, false, audit.ReasonUserNotFound, "nobody"},
		{"long login is bounded", `{"auth":"` + strings.Repeat("z", 100) + `","password":"x","method":"password"}`, false, audit.ReasonUserNotFound, strings.Repeat("z", audit.MaxLoginNameBytes)},
		{"password login disabled", `{"auth":"alice","password":"verystrongpassword1","method":"password"}`, true, audit.ReasonMethodDisabled, "alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := setupSessionTest(t)
			createUserOptionsTestUser(t, store, "alice")
			util.Config.PasswordLoginDisable = tt.disablePassword
			r, rec := withAuditRecorder(loginRequest(store, tt.body))
			w := httptest.NewRecorder()

			login(w, r)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			got := onlyEvent(t, rec, audit.AuthLogin)
			assert.Equal(t, audit.OutcomeFailure, got.Event.Outcome)
			assert.Equal(t, tt.wantReason, got.Event.Reason)
			assert.Equal(t, &audit.Target{Type: audit.TargetUser, Name: tt.wantName}, got.Event.Target)
			assert.Equal(t, audit.AnonymousActor(), got.Actor)
		})
	}
}

// The provider name in the request is attacker input, so it is not recorded.
func TestLogin_UnknownLDAPProviderIsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	r, rec := withAuditRecorder(loginRequest(store, `{"auth":"alice","password":"x","method":"ldap","provider":"`+strings.Repeat("p", 5000)+`"}`))
	w := httptest.NewRecorder()

	login(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, "the response does not change")
	got := onlyEvent(t, rec, audit.AuthLogin)
	assert.Equal(t, audit.ReasonMethodDisabled, got.Event.Reason)
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodLDAP}, got.Event.Metadata)
	assert.Equal(t, &audit.Target{Type: audit.TargetUser, Name: "alice"}, got.Event.Target)
}

func TestLDAPFailureReason(t *testing.T) {
	assert.Equal(t, audit.ReasonUserNotFound, ldapFailureReason(nil))
	assert.Equal(t, audit.ReasonInvalidCredentials, ldapFailureReason(loginError{reason: audit.ReasonInvalidCredentials}))
	assert.Equal(t, audit.ReasonProviderError, ldapFailureReason(errors.New("connection refused")))
}

func TestLDAPUserBindError(t *testing.T) {
	err := ldapUserBindError(ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("bad")))
	assert.Equal(t, loginError{reason: audit.ReasonInvalidCredentials}, err)
	other := errors.New("server down")
	assert.Equal(t, other, ldapUserBindError(other))
}

func TestLoginFailureReason(t *testing.T) {
	assert.Equal(t, audit.ReasonUserNotFound, loginFailureReason(loginError{reason: audit.ReasonUserNotFound}))
	assert.Equal(t, audit.ReasonInvalidCredentials, loginFailureReason(db.ErrNotFound))
	assert.Equal(t, audit.ReasonInternalError, loginFailureReason(errors.New("db down")))
	assert.ErrorIs(t, loginError{reason: audit.ReasonUserNotFound}, db.ErrNotFound)
}

func TestResolveExternalUser_Resolution(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	_, resolution, err := resolveExternalUser(store, ldapProfile("cn=new,dc=x", "new@example.com"))
	require.NoError(t, err)
	assert.Equal(t, resolvedProvisioned, resolution)

	_, resolution, err = resolveExternalUser(store, ldapProfile("cn=new,dc=x", "new@example.com"))
	require.NoError(t, err)
	assert.Equal(t, resolvedExisting, resolution)
}

func TestRecordExternalResolution(t *testing.T) {
	user := db.User{ID: 5, Username: "jdoe"}
	meta := audit.AuthMethodMetadata{Method: audit.LoginMethodLDAP, Provider: "corp"}
	tests := []struct {
		resolution externalResolution
		want       []audit.Kind
	}{
		{resolvedExisting, []audit.Kind{}},
		{resolvedLinked, []audit.Kind{audit.IAMExternalIdentityLink}},
		{resolvedProvisioned, []audit.Kind{audit.IAMUserAutoProvision}},
	}
	for _, tt := range tests {
		r, rec := withAuditRecorder(httptest.NewRequest(http.MethodPost, "/api/auth/login", nil))
		recordExternalResolution(r, user, tt.resolution, meta)
		kinds, err := rec.Kinds()
		require.NoError(t, err)
		assert.Equal(t, tt.want, kinds)
	}
}

func TestLinkExternalIdentity_ReportsNewLink(t *testing.T) {
	store := setupIdentityTest(t, "auto")
	user := createUserOptionsTestUser(t, store, "alice")

	linked, err := linkExternalIdentity(store, user, db.IdentityTypeOidc, "keycloak", "sub-1")
	require.NoError(t, err)
	assert.True(t, linked)

	linked, err = linkExternalIdentity(store, user, db.IdentityTypeOidc, "keycloak", "sub-1")
	require.NoError(t, err)
	assert.False(t, linked)
}

func TestLogout_IsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	user := createUserOptionsTestUser(t, store, "alice")
	r := helpers.SetContextValue(httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil), "store", store)
	r = addSessionCookie(t, store, r, user, db.SessionVerificationNone)
	r, rec := withAuditRecorder(r)

	logout(httptest.NewRecorder(), r)

	got := onlyEvent(t, rec, audit.AuthLogout)
	assert.Equal(t, audit.UserActor(user.ID, "alice", audit.AuthSession, ""), got.Actor)
	assert.Equal(t, audit.UserTarget(user.ID, "alice"), got.Event.Target)
}

func TestLogout_WithoutSessionIsNotRecorded(t *testing.T) {
	store := setupSessionTest(t)
	r, rec := withAuditRecorder(helpers.SetContextValue(httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil), "store", store))
	logout(httptest.NewRecorder(), r)
	assert.Empty(t, rec.All())
}

func TestOidcRedirect_MissingStateIsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	util.Config.OidcProviders = map[string]util.OidcProvider{"corp": {}}
	r := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/corp/redirect", nil)
	r = mux.SetURLVars(helpers.SetContextValue(r, "store", store), map[string]string{"provider": "corp"})
	r, rec := withAuditRecorder(r)

	oidcRedirect(httptest.NewRecorder(), r)

	got := onlyEvent(t, rec, audit.AuthLogin)
	assert.Equal(t, audit.ReasonInvalidState, got.Event.Reason)
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodOIDC, Provider: "corp"}, got.Event.Metadata)
	assert.Nil(t, got.Event.Target)
}

func TestOidcRedirect_UnknownProviderIsRecorded(t *testing.T) {
	store := setupSessionTest(t)
	util.Config.OidcProviders = map[string]util.OidcProvider{}
	state := base64.URLEncoding.EncodeToString([]byte(`{"csrf":"abc"}`))
	r := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/nope/redirect?state="+state, nil)
	r.AddCookie(&http.Cookie{Name: "oauthstate", Value: "abc"})
	r = mux.SetURLVars(helpers.SetContextValue(r, "store", store), map[string]string{"provider": "nope"})
	r, rec := withAuditRecorder(r)

	oidcRedirect(httptest.NewRecorder(), r)

	got := onlyEvent(t, rec, audit.AuthLogin)
	assert.Equal(t, audit.ReasonMethodDisabled, got.Event.Reason)
	assert.Equal(t, audit.AuthMethodMetadata{Method: audit.LoginMethodOIDC}, got.Event.Metadata)
}

type failingUserUpdateStore struct {
	db.Store
}

func (failingUserUpdateStore) UpdateUser(db.UserWithPwd) error {
	return errors.New("update failed")
}

// The handler records the link from these results even when the sync fails.
func TestResolveExternalUser_LinkSurvivesFailedSync(t *testing.T) {
	store := setupIdentityTest(t, "auto")
	legacy, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "Old Name", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	user, resolution, err := resolveExternalUser(failingUserUpdateStore{Store: store}, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))

	require.Error(t, err)
	assert.Equal(t, resolvedLinked, resolution)
	assert.Equal(t, legacy.ID, user.ID)
	ids, err := store.GetUserExternalIdentities(legacy.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 1, "the link is stored")
}
