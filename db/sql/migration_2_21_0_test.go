package sql

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_21_0_McpServersApplyAndRollback(t *testing.T) {
	before := "2.20.18"
	store := InitConfigCreateTestStoreAt(&before)

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	shared, err := store.CreateAccessKey(db.AccessKey{Name: "shared", ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	target := "2.21.0"
	require.NoError(t, db.Migrate(store, &target))

	inv, err := store.CreateInventory(db.Inventory{ProjectID: project.ID, Name: "mcp", Type: db.InventoryMcp})
	require.NoError(t, err)
	secret := "enc"
	_, err = store.SetMcpServers(project.ID, inv.ID, []db.McpServerChange{{
		Server: db.McpServer{Name: "a", URL: "https://a", Enabled: true},
		Token:  &db.McpTokenChange{Operation: db.McpTokenCreate, Key: db.AccessKey{Secret: &secret}},
	}})
	require.NoError(t, err)

	// enabled defaults to true for rows written without it.
	_, err = store.Sql().Exec(store.PrepareQuery("insert into project__inventory_mcp (inventory_id, name, url) values (?, ?, ?)"),
		inv.ID, "b", "https://b")
	require.NoError(t, err)
	servers, err := store.GetMcpServers(project.ID, inv.ID)
	require.NoError(t, err)
	require.Len(t, servers, 2)
	assert.True(t, servers[1].Enabled)

	// unique (inventory_id, name) is enforced by the database too.
	_, err = store.Sql().Exec(store.PrepareQuery("insert into project__inventory_mcp (inventory_id, name, url) values (?, ?, ?)"),
		inv.ID, "a", "https://c")
	assert.Error(t, err)

	// One token per server is enforced by the database too.
	_, err = store.Sql().Exec(store.PrepareQuery("insert into access_key (name, type, project_id, owner, mcp_id) values (?, ?, ?, ?, ?)"),
		"dup", db.AccessKeyString, project.ID, db.AccessKeyMcp, servers[0].ID)
	assert.Error(t, err)

	require.NoError(t, db.Rollback(store, before))

	_, err = store.Sql().SelectInt("select count(*) from project__inventory_mcp")
	assert.Error(t, err, "the table is dropped")
	_, err = store.Sql().SelectInt("select count(mcp_id) from access_key")
	assert.Error(t, err, "the column is dropped")
	_, err = store.GetAccessKey(project.ID, shared.ID)
	assert.NoError(t, err, "other keys survive the rollback")
	n, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from access_key where owner = ?"), db.AccessKeyMcp)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n, "tokens are removed with their servers")

	// The migration applies again after a rollback.
	require.NoError(t, db.Migrate(store, &target))
}

func TestMigration_2_21_0_Dialects(t *testing.T) {
	for _, dialect := range []string{"mysql", util.DbDriverSQLite, "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			up := strings.ToLower(strings.Join(getVersionSQL(dialect, "v2.21.0.sql", false), "\n"))
			assert.Contains(t, up, "create table `project__inventory_mcp`")
			assert.Contains(t, up, "unique (`inventory_id`, `name`)")
			assert.Contains(t, up, "create unique index `access_key__mcp_id`")
			assert.Equal(t, 2, strings.Count(up, "on delete cascade"), "servers go with the inventory, tokens with the server")

			down := strings.ToLower(strings.Join(getVersionSQL(dialect, "v2.21.0.err.sql", false), "\n"))
			assert.Contains(t, down, "drop column `mcp_id`")
			assert.Contains(t, down, "drop table `project__inventory_mcp`")

			if dialect == util.DbDriverSQLite {
				assert.NotContains(t, down, "drop foreign key")
			} else {
				assert.Contains(t, down, "drop foreign key `access_key__mcp_fk`")
			}

			if dialect == "mysql" {
				assert.Contains(t, up, "utf8mb4_bin", "names are case-sensitive like on the other dialects")
			}
		})
	}
}
