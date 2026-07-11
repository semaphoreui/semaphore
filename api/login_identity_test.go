package api

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentityTest(t *testing.T, emailMatching string) db.Store {
	t.Helper()
	// IMPORTANT: CreateTestStore() overwrites util.Config (db/sql/SqlDb.go:124),
	// so the option must be set AFTER creating the store.
	store := sql.CreateTestStore()
	util.Config.ExternalAuthEmailMatching = emailMatching
	return store
}

func ldapProfile(uid string, email string) externalUserProfile {
	return externalUserProfile{
		Type:            db.IdentityTypeLdap,
		Provider:        "ldap",
		ExternalUID:     uid,
		Username:        "jdoe",
		Name:            "John Doe",
		Email:           email,
		MatchByUsername: true,
	}
}

func TestResolveExternalUser_CreatesUserAndIdentity(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	user, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	require.NoError(t, err)
	assert.True(t, user.External)

	ids, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, "cn=jdoe,dc=x", ids[0].ExternalUID)
}

func TestResolveExternalUser_FindsByIdentityAfterEmailChange(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	first, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	require.NoError(t, err)

	// Email changed at the IdP: same identity, attributes synced, no new user.
	second, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "john@example.com"))
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "john@example.com", second.Email)
}

func TestResolveExternalUser_AutoAdoptsLegacyExternalUser(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	// Pre-2.20 external user: no identity rows.
	legacy, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	user, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	require.NoError(t, err)
	assert.Equal(t, legacy.ID, user.ID)

	ids, err := store.GetUserExternalIdentities(legacy.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestResolveExternalUser_AutoDoesNotAdoptPinnedUser(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	// User already pinned to a different provider identity.
	pinned, err := resolveExternalUser(store, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "sub-1",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	// Same email arrives from another provider: must NOT reuse the account.
	other, err := resolveExternalUser(store, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "okta", ExternalUID: "sub-2",
		Username: "jdoe2", Name: "John Doe", Email: "jdoe@example.com",
	})
	if err == nil {
		assert.NotEqual(t, pinned.ID, other.ID)
	} else {
		// Unique email constraint may forbid the second user - also acceptable,
		// but it must NOT be a silent merge.
		assert.Error(t, err)
	}
}

func TestResolveExternalUser_AlwaysLinksSecondProvider(t *testing.T) {
	store := setupIdentityTest(t, "always")

	first, err := resolveExternalUser(store, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "sub-1",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	second, err := resolveExternalUser(store, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "okta", ExternalUID: "sub-2",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)

	ids, err := store.GetUserExternalIdentities(first.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestResolveExternalUser_NeverSkipsEmailMatching(t *testing.T) {
	store := setupIdentityTest(t, "never")

	_, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// Legacy user is NOT adopted; creation of a duplicate-email user must fail
	// rather than silently merge.
	_, err = resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	assert.Error(t, err)
}

func TestResolveExternalUser_NeverAdoptsLocalUser(t *testing.T) {
	store := setupIdentityTest(t, "always")

	local, err := store.CreateUser(db.UserWithPwd{
		Pwd: "verystrongpassword1",
		User: db.User{
			Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
		},
	})
	require.NoError(t, err)

	res, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	if err == nil {
		assert.NotEqual(t, local.ID, res.ID)
	} else {
		assert.Error(t, err)
	}

	// The local account must remain identity-free.
	ids, err := store.GetUserExternalIdentities(local.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestLinkExternalIdentity_LocalUser(t *testing.T) {
	store := setupIdentityTest(t, "never")

	local, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com"},
	})
	require.NoError(t, err)

	require.NoError(t, linkExternalIdentity(store, local, db.IdentityTypeOidc, "keycloak", "sub-1"))

	// Idempotent for the same user.
	require.NoError(t, linkExternalIdentity(store, local, db.IdentityTypeOidc, "keycloak", "sub-1"))

	ids, err := store.GetUserExternalIdentities(local.ID)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	// User stays local: password login must still work.
	fresh, err := store.GetUser(local.ID)
	require.NoError(t, err)
	assert.False(t, fresh.External)

	// After linking, SSO login resolves to the local user...
	resolved, err := resolveExternalUser(store, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "sub-1",
		Username: "x", Name: "IdP Name", Email: "idp@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, local.ID, resolved.ID)

	// ...and the IdP does NOT overwrite the local profile.
	fresh, err = store.GetUser(local.ID)
	require.NoError(t, err)
	assert.Equal(t, "jdoe@example.com", fresh.Email)
	assert.Equal(t, "John Doe", fresh.Name)
}

func TestLinkExternalIdentity_Conflicts(t *testing.T) {
	store := setupIdentityTest(t, "never")

	alice, err := store.CreateUserWithoutPassword(db.User{
		Username: "alice", Name: "Alice", Email: "alice@example.com", External: true,
	})
	require.NoError(t, err)
	bob, err := store.CreateUserWithoutPassword(db.User{
		Username: "bob", Name: "Bob", Email: "bob@example.com", External: true,
	})
	require.NoError(t, err)

	require.NoError(t, linkExternalIdentity(store, alice, db.IdentityTypeOidc, "keycloak", "sub-a"))

	// UID owned by another user. The sentinel is what oidcRedirect maps
	// to the user-visible "link_conflict" error code.
	assert.ErrorIs(t, linkExternalIdentity(store, bob, db.IdentityTypeOidc, "keycloak", "sub-a"), errIdentityLinkedToAnother)

	// Second identity for the same provider → "link_provider_exists".
	assert.ErrorIs(t, linkExternalIdentity(store, alice, db.IdentityTypeOidc, "keycloak", "sub-a2"), errProviderAlreadyLinked)
}

func TestLinkExternalIdentity_SameProviderNameDifferentType(t *testing.T) {
	store := setupIdentityTest(t, "never")

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "alice", Name: "Alice", Email: "alice@example.com", External: true,
	})
	require.NoError(t, err)

	// "corp" LDAP and "corp" OIDC are different providers.
	require.NoError(t, linkExternalIdentity(store, user, db.IdentityTypeLdap, "corp", "cn=alice,dc=example,dc=org"))
	require.NoError(t, linkExternalIdentity(store, user, db.IdentityTypeOidc, "corp", "sub-alice"))

	ids, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestSyncExternalUserAttrs_SkipsLocalUsers(t *testing.T) {
	store := setupIdentityTest(t, "never")

	local, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com"},
	})
	require.NoError(t, err)

	synced, err := syncExternalUserAttrs(store, local, externalUserProfile{
		Type: db.IdentityTypeOidc, Provider: "keycloak", ExternalUID: "sub-1",
		Email: "idp@example.com", Name: "IdP Name",
	})
	require.NoError(t, err)
	assert.Equal(t, "jdoe@example.com", synced.Email)
}
