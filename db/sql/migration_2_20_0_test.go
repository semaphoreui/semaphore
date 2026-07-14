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
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// Table exists and accepts a row.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, created) values (?, 'ldap', 'cn=admin,dc=example,dc=org', ?)",
		user.ID, user.Created)
	assert.NoError(t, err)

	// Unique (provider, external_uid) index rejects a duplicate.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, created) values (?, 'ldap', 'cn=admin,dc=example,dc=org', ?)",
		user.ID, user.Created)
	assert.Error(t, err)
}
