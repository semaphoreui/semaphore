package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPSettings_StoredInOptions(t *testing.T) {
	store := InitConfigCreateTestStore()

	settings, err := db.ReadMCPSettings(store)
	require.NoError(t, err)
	assert.Equal(t, db.DefaultMCPSettings(), settings, "no rows give the defaults")

	require.NoError(t, db.WriteMCPSettings(store, db.MCPSettings{AllowRun: false, AllowDestroy: true, AdminAllowWrite: true}))
	options, err := store.GetOptions(db.RetrieveQueryParams{Filter: "mcp"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{db.MCPAllowRunOption: "false", db.MCPAllowDestroyOption: "true",
		db.MCPAdminAllowWriteOption: "true", db.MCPAdminAllowDestroyOption: "false"}, options)

	settings, err = db.ReadMCPSettings(store)
	require.NoError(t, err)
	assert.Equal(t, db.MCPSettings{AllowRun: false, AllowDestroy: true, AdminAllowWrite: true}, settings)

	require.NoError(t, db.WriteMCPSettings(store, db.MCPSettings{AllowRun: true, AdminAllowDestroy: true}))
	settings, err = db.ReadMCPSettings(store)
	require.NoError(t, err)
	assert.Equal(t, db.MCPSettings{AllowRun: true, AdminAllowDestroy: true}, settings, "a second write updates the rows")
}

// The MCP settings live in the option table only: FillConfigFromDB must not lay
// them, or any other mcp.* row, over util.Config — the mcp section is a ceiling
// set by the file or the environment (db:"-").
func TestFillConfigFromDB_IgnoresMCPOptions(t *testing.T) {
	store := InitConfigCreateTestStore()
	util.Config.MCP = &util.MCPConfig{Enabled: false}
	mcpConfig := util.Config.MCP

	require.NoError(t, db.WriteMCPSettings(store, db.MCPSettings{AllowRun: false, AllowDestroy: true}))
	require.NoError(t, store.SetOption("mcp.enabled", "true"))

	require.NoError(t, db.FillConfigFromDB(store))

	assert.Same(t, mcpConfig, util.Config.MCP, "the mcp section is left as the file set it")
	assert.Equal(t, util.MCPConfig{Enabled: false}, *util.Config.MCP)
	assert.False(t, util.MCPEnabled())
}
