package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_0_CreatesExternalIdentityTable(t *testing.T) {
	// CreateTestStore (db/sql/SqlDb.go:124) sets util.Config, connects an
	// in-memory sqlite (PRAGMA foreign_keys=ON) and runs db.Migrate(store, nil).
	store := InitConfigCreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// Table exists and accepts a row; type is required.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', ?)",
		user.ID, user.Created)
	require.Error(t, err)

	// Same (provider, external_uid) under a different type is allowed:
	// an OIDC provider may share a name with an LDAP provider.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'ldap', ?)",
		user.ID, user.Created)
	assert.NoError(t, err)

	// Duplicate (type, provider, external_uid) is rejected.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'ldap', ?)",
		user.ID, user.Created)
	assert.Error(t, err)
}
