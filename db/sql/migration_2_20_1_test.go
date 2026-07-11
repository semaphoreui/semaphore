package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_1_AddsIdentityTypeColumn(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// The type column exists and accepts both values.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'ldap', ?)",
		user.ID, user.Created)
	require.NoError(t, err)

	// Same (provider, external_uid) under a different type is allowed:
	// an OIDC provider may share a name with an LDAP provider.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'oidc', ?)",
		user.ID, user.Created)
	assert.NoError(t, err)

	// Duplicate (type, provider, external_uid) is rejected.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'ldap', ?)",
		user.ID, user.Created)
	assert.Error(t, err)
}
