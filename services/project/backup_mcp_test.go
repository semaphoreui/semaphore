package project

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Until MCP servers are part of the backup format, their tokens must stay out
// of the project key list: restored as project keys they would show up in the
// Key Store without a server, and the store refuses to create them.
func TestBackup_LeavesOutMcpTokens(t *testing.T) {
	previousConfig := util.Config
	t.Cleanup(func() { util.Config = previousConfig })
	util.Config = &util.ConfigType{TmpPath: "/tmp"}

	store := sql.InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "agents"})
	require.NoError(t, err)
	_, err = store.CreateAccessKey(db.AccessKey{Name: "shared", ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID, Name: "mcp", Type: db.InventoryMcp})
	require.NoError(t, err)
	secret := "encrypted"
	_, err = store.SetMcpServers(proj.ID, inv.ID, []db.McpServerChange{{
		Server: db.McpServer{Name: "github", URL: "https://api.githubcopilot.com/mcp/", Enabled: true},
		Token:  &db.McpTokenChange{Operation: db.McpTokenCreate, Key: db.AccessKey{Secret: &secret}},
	}})
	require.NoError(t, err)

	backup, err := GetBackup(proj.ID, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	require.Len(t, backup.Keys, 1)
	assert.Equal(t, "shared", backup.Keys[0].Name)

	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{
		Username: "test", Name: "Test", Email: "test@example.com", Admin: true,
	}})
	require.NoError(t, err)

	backup.Meta.Name = "agents restored"
	_, err = backup.Restore(user, store, proFactory.NewWorkflowStore(store))
	assert.NoError(t, err)
}
