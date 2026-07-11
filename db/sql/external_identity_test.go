package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalIdentityCRUD(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe",
		Name:     "John Doe",
		Email:    "jdoe@example.com",
		External: true,
	})
	require.NoError(t, err)

	// Not found before creation.
	_, err = store.GetExternalIdentity("ldap", "cn=jdoe,dc=example,dc=org")
	assert.ErrorIs(t, err, db.ErrNotFound)

	created, err := store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "ldap",
		ExternalUID: "cn=jdoe,dc=example,dc=org",
	})
	require.NoError(t, err)
	assert.NotZero(t, created.Created)
	assert.Equal(t, "oidc", created.Type) // defaults to oidc

	found, err := store.GetExternalIdentity("ldap", "cn=jdoe,dc=example,dc=org")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.UserID)

	// Unique (type, provider, external_uid).
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "ldap",
		ExternalUID: "cn=jdoe,dc=example,dc=org",
	})
	assert.Error(t, err)

	// Same user, second provider — allowed.
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "keycloak",
		ExternalUID: "8b53f1e0-0000-0000-0000-000000000000",
	})
	require.NoError(t, err)

	list, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2)

	require.NoError(t, store.DeleteExternalIdentity(user.ID, "keycloak"))
	list, err = store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Cascade on user deletion.
	require.NoError(t, store.DeleteUser(user.ID))
	_, err = store.GetExternalIdentity("ldap", "cn=jdoe,dc=example,dc=org")
	assert.ErrorIs(t, err, db.ErrNotFound)
}
