package export

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendValues_KeepsScopeAndKeyApart(t *testing.T) {
	var values ValueMap[db.InventoryGroup]
	values.setUniqueKeys(true)

	require.NoError(t, values.appendValues([]db.InventoryGroup{{ID: 23}}, "1"))
	require.NoError(t, values.appendValues([]db.InventoryGroup{{ID: 3}}, "12"))

	assert.ErrorContains(t, values.appendValues([]db.InventoryGroup{{ID: 3}}, "12"), "duplicate key 3")
}
