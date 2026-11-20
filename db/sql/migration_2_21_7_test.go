package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_21_7_AddsInventoryGroups(t *testing.T) {
	before := "2.21.6"
	store := InitConfigCreateTestStoreAt(&before)

	target := "2.21.7"
	require.NoError(t, db.Migrate(store, &target))

	for _, query := range []string{
		"select count(*) from project__inventory_group",
		"select count(*) from project__inventory_group_host",
		"select count(*) from project__inventory_group_child",
	} {
		count, err := store.Sql().SelectInt(query)
		require.NoError(t, err, query)
		assert.Equal(t, int64(0), count, query)
	}

	require.NoError(t, db.Rollback(store, "2.21.6"))
}
