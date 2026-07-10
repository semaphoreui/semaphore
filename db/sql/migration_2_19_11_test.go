package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMigrationTestDB(t *testing.T) *SqlDb {
	t.Helper()
	util.Config = &util.ConfigType{
		SQLite: &util.DbConfig{
			Hostname: ":memory:",
		},
		Dialect: "sqlite",
		Log: &util.ConfigLog{
			Events: &util.EventLogType{},
			Tasks:  &util.TaskLogType{},
		},
		Process: &util.ConfigProcess{},
	}
	store := CreateDb(util.DbDriverSQLite)
	store.Connect()
	return store
}

// TestMigration_2_19_11 verifies that workflow node `limit` values are moved
// into project__task_params rows referenced by the new task_params_id column.
func TestMigration_2_19_11(t *testing.T) {
	store := setupMigrationTestDB(t)

	target := "2.19.2"
	require.NoError(t, db.Migrate(store, &target))

	proj, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_template (project_id, name) values (?, ?)", proj.ID, "wf")
	require.NoError(t, err)
	workflowID, err := store.Sql().SelectInt("select id from project__workflow_template where name = 'wf'")
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_node (workflow_template_id, template_id, `limit`) values (?, ?, ?)",
		workflowID, 0, `["web*","db"]`)
	require.NoError(t, err)

	require.NoError(t, db.Migrate(store, nil))

	paramsID, err := store.Sql().SelectNullInt(
		"select task_params_id from project__workflow_node where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	require.True(t, paramsID.Valid, "expected node to reference a task params row")

	var taskParams db.TaskParams
	err = store.Sql().SelectOne(&taskParams,
		"select * from project__task_params where id = ?", paramsID.Int64)
	require.NoError(t, err)
	assert.Equal(t, proj.ID, taskParams.ProjectID)
	assert.Equal(t, []any{"web*", "db"}, taskParams.Params["limit"])
}

// TestMigration_2_19_11_InventoryEnvironment verifies per-node inventory and
// environment overrides are migrated into project__task_params.
func TestMigration_2_19_11_InventoryEnvironment(t *testing.T) {
	store := setupMigrationTestDB(t)

	target := "2.19.2"
	require.NoError(t, db.Migrate(store, &target))

	proj, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	env, err := store.CreateEnvironment(db.Environment{
		ProjectID: proj.ID,
		Name:      "env",
		JSON:      `{"FOO":"bar"}`,
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		Name:      "inv",
		Type:      db.InventoryStatic,
		Inventory: "localhost",
	})
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_template (project_id, name) values (?, ?)", proj.ID, "wf")
	require.NoError(t, err)
	workflowID, err := store.Sql().SelectInt("select id from project__workflow_template where name = 'wf'")
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_node (workflow_template_id, template_id, inventory_id, environment_id) values (?, ?, ?, ?)",
		workflowID, 0, inv.ID, env.ID)
	require.NoError(t, err)

	require.NoError(t, db.Migrate(store, nil))

	paramsID, err := store.Sql().SelectNullInt(
		"select task_params_id from project__workflow_node where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	require.True(t, paramsID.Valid)

	var taskParams db.TaskParams
	err = store.Sql().SelectOne(&taskParams,
		"select * from project__task_params where id = ?", paramsID.Int64)
	require.NoError(t, err)
	require.NotNil(t, taskParams.InventoryID)
	assert.Equal(t, inv.ID, *taskParams.InventoryID)
	assert.Equal(t, env.JSON, taskParams.Environment)
}

// TestMigration_2_19_12_RepairsPartialMigration simulates the buggy 2.19.11
// migration that only copied limit, then verifies 2.19.12 backfills inventory.
func TestMigration_2_19_12_RepairsPartialMigration(t *testing.T) {
	store := setupMigrationTestDB(t)

	target := "2.19.2"
	require.NoError(t, db.Migrate(store, &target))

	proj, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{
		ProjectID: proj.ID,
		Name:      "inv",
		Type:      db.InventoryStatic,
		Inventory: "localhost",
	})
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_template (project_id, name) values (?, ?)", proj.ID, "wf")
	require.NoError(t, err)
	workflowID, err := store.Sql().SelectInt("select id from project__workflow_template where name = 'wf'")
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_node (workflow_template_id, template_id, `limit`, inventory_id) values (?, ?, ?, ?)",
		workflowID, 0, `["web"]`, inv.ID)
	require.NoError(t, err)

	// Apply schema change from 2.19.11 without its PostApply logic.
	target = "2.19.11"
	require.NoError(t, db.Migrate(store, &target))

	// Simulate buggy migration: only limit was copied.
	res, err := store.Sql().Exec(
		"insert into project__task_params (project_id, params, environment, message) values (?, ?, '', '')",
		proj.ID, `{"limit":["web"]}`)
	require.NoError(t, err)
	paramsID, err := res.LastInsertId()
	require.NoError(t, err)
	_, err = store.Sql().Exec(
		"update project__workflow_node set task_params_id = ? where workflow_template_id = ?",
		paramsID, workflowID)
	require.NoError(t, err)

	require.NoError(t, db.Migrate(store, nil))

	var taskParams db.TaskParams
	err = store.Sql().SelectOne(&taskParams,
		"select * from project__task_params where id = ?", paramsID)
	require.NoError(t, err)
	assert.Equal(t, []any{"web"}, taskParams.Params["limit"])
	require.NotNil(t, taskParams.InventoryID)
	assert.Equal(t, inv.ID, *taskParams.InventoryID)
}
