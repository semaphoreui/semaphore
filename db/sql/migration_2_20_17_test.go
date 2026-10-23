package sql

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_17_SQL(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			apply := strings.Join(getVersionSQL(dialect, "v2.20.17.sql", false), "\n")
			assert.Contains(t, apply, "not null default 'api'")
			rollback := strings.Join(getVersionSQL(dialect, "v2.20.17.err.sql", false), "\n")
			assert.Contains(t, rollback, "drop")
		})
	}
}

func TestMigration_2_20_17_ExistingTokensBecomeAPITokens(t *testing.T) {
	before := "2.20.16"
	store := InitConfigCreateTestStoreAt(&before)
	user, err := store.CreateUserWithoutPassword(db.User{Username: "a", Name: "A", Email: "a@e.st"})
	require.NoError(t, err)
	_, err = store.exec("insert into user__token (id, created, expired, user_id, name) values (?, ?, false, ?, ?)",
		"existingtoken123", "2026-10-01 00:00:00", user.ID, "old")
	require.NoError(t, err)

	target := "2.20.17"
	require.NoError(t, db.Migrate(store, &target))

	token, err := store.GetAPIToken("existingtoken123", db.APITokenTypeAPI)
	require.NoError(t, err, "a token issued before the migration keeps working on the REST API")
	assert.Equal(t, db.APITokenTypeAPI, token.Type)
	_, err = store.GetAPIToken("existingtoken123", db.APITokenTypeMCP)
	assert.ErrorIs(t, err, db.ErrNotFound)

	require.NoError(t, db.Rollback(store, "2.20.16"))
	columns, err := store.Sql().SelectInt(store.PrepareQuery(
		"select count(*) from pragma_table_info('user__token') where name='type'"))
	require.NoError(t, err)
	assert.Zero(t, columns, "the rollback drops the column")
}
