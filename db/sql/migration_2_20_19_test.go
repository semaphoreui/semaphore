package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_19_BackfillsRegistry(t *testing.T) {
	before := "2.20.18"
	store := InitConfigCreateTestStoreAt(&before)
	project, template := createTestTemplate(t, store)
	created := time.Now().UTC()
	first, err := store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: template.ID, Status: "success", Created: created.Add(-time.Hour)}, 0)
	require.NoError(t, err)
	second, err := store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: template.ID, Status: "error", Created: created}, 0)
	require.NoError(t, err)
	for _, row := range []db.AnsibleTaskHost{
		{TaskID: first.ID, Host: "Edge-01", Ok: 1},
		{TaskID: first.ID, Host: "edge-02", Failed: 1},
		{TaskID: first.ID, Host: "localhost", Ok: 1},
		{TaskID: second.ID, Host: "edge-01", Unreachable: 1},
		{TaskID: second.ID, Host: "edge-03", Unreachable: 1},
	} {
		_, err = store.exec("insert into task__ansible_host (task_id, project_id, host, failed, ignored, changed, ok, rescued, skipped, unreachable) values (?, ?, ?, ?, 0, 0, ?, 0, 0, ?)",
			row.TaskID, project.ID, row.Host, row.Failed, row.Ok, row.Unreachable)
		require.NoError(t, err)
	}

	target := "2.20.19"
	require.NoError(t, db.Migrate(store, &target))

	hosts, err := store.Sql().SelectInt("select count(*) from host")
	require.NoError(t, err)
	assert.Equal(t, int64(3), hosts)
	var edge struct {
		ID            int        `db:"id"`
		DisplayName   string     `db:"display_name"`
		Runs          int        `db:"runs"`
		LastStatus    string     `db:"last_status"`
		LastTaskID    int        `db:"last_task_id"`
		LastAutomated *time.Time `db:"last_automated"`
	}
	require.NoError(t, store.Sql().SelectOne(&edge, "select id, display_name, runs, last_status, last_task_id, last_automated from host where name = 'edge-01'"))
	assert.Equal(t, "Edge-01", edge.DisplayName)
	assert.Equal(t, 2, edge.Runs)
	assert.Equal(t, "unreachable", edge.LastStatus)
	assert.Equal(t, second.ID, edge.LastTaskID)
	var firstTask struct {
		Created time.Time `db:"created"`
	}
	require.NoError(t, store.Sql().SelectOne(&firstTask, "select created from task where id = ?", first.ID))
	require.NotNil(t, edge.LastAutomated)
	assert.Equal(t, firstTask.Created, *edge.LastAutomated)
	unreachableOnly, err := store.Sql().SelectInt("select count(*) from host where name = 'edge-03' and last_automated is null")
	require.NoError(t, err)
	assert.Equal(t, int64(1), unreachableOnly)
	linked, err := store.Sql().SelectInt("select count(*) from task__ansible_host where host_id is not null and host <> 'localhost'")
	require.NoError(t, err)
	assert.Equal(t, int64(4), linked)
	local, err := store.Sql().SelectInt("select count(*) from task__ansible_host where host_id is null and host = 'localhost'")
	require.NoError(t, err)
	assert.Equal(t, int64(1), local)
	projects, err := store.Sql().SelectInt("select count(*) from host__project")
	require.NoError(t, err)
	assert.Equal(t, int64(3), projects)
	templates, err := store.Sql().SelectInt("select count(*) from host__template")
	require.NoError(t, err)
	assert.Equal(t, int64(3), templates)

	require.NoError(t, db.Rollback(store, "2.20.18"))
}
