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
	_, err = store.GetExternalIdentity(db.IdentityTypeLdap, "ldap", "cn=jdoe,dc=example,dc=org")
	assert.ErrorIs(t, err, db.ErrNotFound)

	created, err := store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Type:        db.IdentityTypeLdap,
		Provider:    "ldap",
		ExternalUID: "cn=jdoe,dc=example,dc=org",
	})
	require.NoError(t, err)
	assert.NotZero(t, created.Created)
	assert.Equal(t, db.IdentityTypeLdap, created.Type)

	found, err := store.GetExternalIdentity(db.IdentityTypeLdap, "ldap", "cn=jdoe,dc=example,dc=org")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.UserID)

	// Unique (type, provider, external_uid).
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Type:        db.IdentityTypeLdap,
		Provider:    "ldap",
		ExternalUID: "cn=jdoe,dc=example,dc=org",
	})
	assert.Error(t, err)

	// Same user, second provider — allowed; empty Type defaults to oidc.
	kc, err := store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "keycloak",
		ExternalUID: "8b53f1e0-0000-0000-0000-000000000000",
	})
	require.NoError(t, err)
	assert.Equal(t, db.IdentityTypeOidc, kc.Type)

	list, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2)

	require.NoError(t, store.DeleteExternalIdentity(user.ID, db.IdentityTypeOidc, "keycloak"))
	list, err = store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Cascade on user deletion.
	require.NoError(t, store.DeleteUser(user.ID))
	_, err = store.GetExternalIdentity(db.IdentityTypeLdap, "ldap", "cn=jdoe,dc=example,dc=org")
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestExternalIdentity_TypeSeparatesNamespaces(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: user.ID, Type: db.IdentityTypeLdap, Provider: "corp", ExternalUID: "uid-1",
	})
	require.NoError(t, err)
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: user.ID, Type: db.IdentityTypeOidc, Provider: "corp", ExternalUID: "uid-1",
	})
	require.NoError(t, err)

	found, err := store.GetExternalIdentity(db.IdentityTypeLdap, "corp", "uid-1")
	require.NoError(t, err)
	assert.Equal(t, db.IdentityTypeLdap, found.Type)
}
