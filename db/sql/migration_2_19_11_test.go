package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_19_11 verifies that workflow node `limit` values are moved
// into project__task_params rows referenced by the new task_params_id column.
func TestMigration_2_19_11(t *testing.T) {
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

	// Migrate to the schema right before 2.19.11 and seed a node with a limit.
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
