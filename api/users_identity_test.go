package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAndDeleteUserIdentities(t *testing.T) {
	util.Config = &util.ConfigType{}
	store := sql.InitConfigCreateTestStore()

	admin, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "root", Name: "Root", Email: "root@example.com", Admin: true},
	})
	require.NoError(t, err)

	target, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: target.ID, Type: db.IdentityTypeLdap, Provider: "ldap", ExternalUID: "cn=jdoe,dc=x",
	})
	require.NoError(t, err)

	usersController := NewUsersController(nil)

	// GET list
	r := httptest.NewRequest(http.MethodGet, "/api/users/2/identities", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target)
	w := httptest.NewRecorder()
	usersController.GetUserIdentities(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cn=jdoe,dc=x")

	// DELETE with an invalid identity type is rejected.
	r = httptest.NewRequest(http.MethodDelete, "/api/users/2/identities/bogus/ldap", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target)
	r = mux.SetURLVars(r, map[string]string{"type": "bogus", "provider": "ldap"})
	w = httptest.NewRecorder()
	usersController.DeleteUserIdentity(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// DELETE unlink of the only identity is blocked for external users.
	r = httptest.NewRequest(http.MethodDelete, "/api/users/2/identities/ldap/ldap", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target)
	r = mux.SetURLVars(r, map[string]string{"type": "ldap", "provider": "ldap"})
	w = httptest.NewRecorder()
	usersController.DeleteUserIdentity(w, r)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), errCannotUnlinkLastIdentity.Error())

	ids, err := store.GetUserExternalIdentities(target.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestDeleteUserIdentity_AllowsUnlinkWhenMultipleIdentities(t *testing.T) {
	util.Config = &util.ConfigType{}
	store := sql.InitConfigCreateTestStore()

	admin, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "root", Name: "Root", Email: "root@example.com", Admin: true},
	})
	require.NoError(t, err)

	target, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: target.ID, Type: db.IdentityTypeLdap, Provider: "ldap", ExternalUID: "cn=jdoe,dc=x",
	})
	require.NoError(t, err)
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: target.ID, Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "sub-1",
	})
	require.NoError(t, err)

	usersController := NewUsersController(nil)

	r := httptest.NewRequest(http.MethodDelete, "/api/users/2/identities/ldap/ldap", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target)
	r = mux.SetURLVars(r, map[string]string{"type": "ldap", "provider": "ldap"})
	w := httptest.NewRecorder()
	usersController.DeleteUserIdentity(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)

	ids, err := store.GetUserExternalIdentities(target.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, db.IdentityTypeOidc, ids[0].Type)
}

func TestDeleteUserIdentity_AllowsLocalUserToUnlinkLastIdentity(t *testing.T) {
	util.Config = &util.ConfigType{}
	store := sql.InitConfigCreateTestStore()

	localUser, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com"},
	})
	require.NoError(t, err)

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: localUser.ID, Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "sub-1",
	})
	require.NoError(t, err)

	usersController := NewUsersController(nil)

	r := httptest.NewRequest(http.MethodDelete, "/api/users/1/identities/oidc/keycloak", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &localUser)
	r = helpers.SetContextValue(r, "_user", localUser)
	r = mux.SetURLVars(r, map[string]string{"type": "oidc", "provider": "keycloak"})
	w := httptest.NewRecorder()
	usersController.DeleteUserIdentity(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)

	ids, err := store.GetUserExternalIdentities(localUser.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestLdapProfileMatchesSemaphoreUser(t *testing.T) {
	semaphoreUser := db.User{Username: "jdoe", Email: "jdoe@example.com"}

	assert.True(t, ldapProfileMatchesSemaphoreUser(
		db.User{Username: "jdoe", Email: "jdoe@example.com"},
		semaphoreUser,
	))
	assert.True(t, ldapProfileMatchesSemaphoreUser(
		db.User{Username: "JDOE", Email: "other@example.com"},
		semaphoreUser,
	))
	assert.True(t, ldapProfileMatchesSemaphoreUser(
		db.User{Username: "other", Email: "JDOE@example.com"},
		semaphoreUser,
	))

	assert.False(t, ldapProfileMatchesSemaphoreUser(
		db.User{Username: "alice", Email: "alice@example.com"},
		semaphoreUser,
	))
}

func TestLinkLdapIdentity_LdapDisabled(t *testing.T) {
	util.Config = &util.ConfigType{}
	store := sql.InitConfigCreateTestStore()

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com"},
	})
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/api/user/identities/ldap",
		bytes.NewBufferString(`{"username":"jdoe","password":"secret"}`))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &user)
	w := httptest.NewRecorder()

	linkLdapIdentity(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
