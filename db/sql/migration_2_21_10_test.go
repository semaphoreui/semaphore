package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_21_10_AddsProxies(t *testing.T) {
	before := "2.21.8"
	store := InitConfigCreateTestStoreAt(&before)

	target := "2.21.10"
	require.NoError(t, db.Migrate(store, &target))

	for _, query := range []string{
		"select count(requires_proxy_id) from project__proxy",
		"select count(proxy_id) from project__inventory",
		"select count(proxy_id) from project__repository",
	} {
		_, err := store.Sql().SelectInt(query)
		require.NoError(t, err, query)
	}

	require.NoError(t, db.Rollback(store, before))

	_, err := store.Sql().SelectInt("select count(proxy_id) from project__inventory")
	assert.Error(t, err)
	_, err = store.Sql().SelectInt("select count(*) from project__proxy")
	assert.Error(t, err)
}
