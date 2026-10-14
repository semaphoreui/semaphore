package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_20_8 seeds a workflow on the pre-revision schema and checks
// the upgrade gives it revision 1 with the nodes, edges and runs attached.
func TestMigration_2_20_8(t *testing.T) {
	target := "2.20.7"
	store := InitConfigCreateTestStoreAt(&target)

	proj, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	_, err = store.Sql().Exec(
		"insert into project__workflow_template (project_id, name) values (?, ?)", proj.ID, "wf")
	require.NoError(t, err)
	workflowID, err := store.Sql().SelectInt("select id from project__workflow_template where name = 'wf'")
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		_, err = store.Sql().Exec(
			"insert into project__workflow_node (workflow_template_id, template_id) values (?, ?)", workflowID, 0)
		require.NoError(t, err)
	}
	_, err = store.Sql().Exec(
		"insert into project__workflow_edge (workflow_template_id, source_node_id, destination_node_id, `condition`) "+
			"select ?, min(id), max(id), 'on_success' from project__workflow_node where workflow_template_id = ?",
		workflowID, workflowID)
	require.NoError(t, err)
	_, err = store.Sql().Exec(
		"insert into project__workflow_run (project_id, workflow_template_id, status) values (?, ?, 'success')",
		proj.ID, workflowID)
	require.NoError(t, err)

	require.NoError(t, db.Migrate(store, nil))

	var revisions []db.WorkflowRevision
	_, err = store.Sql().Select(&revisions,
		"select * from project__workflow_revision where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	require.Len(t, revisions, 1)
	assert.Equal(t, 1, revisions[0].Number)
	assert.Equal(t, proj.ID, revisions[0].ProjectID)
	assert.False(t, revisions[0].Created.IsZero())

	revisionID := revisions[0].ID
	for _, table := range []string{"project__workflow_node", "project__workflow_edge", "project__workflow_run"} {
		orphans, err := store.Sql().SelectInt(
			"select count(*) from "+table+" where workflow_template_id = ? and (revision_id is null or revision_id <> ?)",
			workflowID, revisionID)
		require.NoError(t, err)
		assert.Zero(t, orphans, table)
	}

	// A revision pinned by a run is protected by the foreign key; once the run
	// is gone the revision can be deleted and takes its nodes and edges along.
	_, err = store.Sql().Exec("delete from project__workflow_revision where id = ?", revisionID)
	assert.Error(t, err, "revision with a run must not be deletable")

	_, err = store.Sql().Exec("delete from project__workflow_run where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	_, err = store.Sql().Exec("delete from project__workflow_revision where id = ?", revisionID)
	require.NoError(t, err)
	nodes, err := store.Sql().SelectInt("select count(*) from project__workflow_node where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	assert.Zero(t, nodes)
	edges, err := store.Sql().SelectInt("select count(*) from project__workflow_edge where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	assert.Zero(t, edges)
}

// TestMigration_2_20_8_Idempotent re-runs PostApply on an upgraded schema and
// checks it neither duplicates revisions nor touches attached rows.
func TestMigration_2_20_8_Idempotent(t *testing.T) {
	store := InitConfigCreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	_, err = store.Sql().Exec(
		"insert into project__workflow_template (project_id, name) values (?, ?)", proj.ID, "wf")
	require.NoError(t, err)
	workflowID, err := store.Sql().SelectInt("select id from project__workflow_template where name = 'wf'")
	require.NoError(t, err)

	// A template created after the upgrade has no revision yet (the store
	// creates one on save); PostApply must give it exactly one, once.
	for i := 0; i < 2; i++ {
		tx, err := store.Sql().Begin()
		require.NoError(t, err)
		require.NoError(t, migration_2_20_8{db: store}.PostApply(tx))
		require.NoError(t, tx.Commit())
	}

	count, err := store.Sql().SelectInt(
		"select count(*) from project__workflow_revision where workflow_template_id = ?", workflowID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
