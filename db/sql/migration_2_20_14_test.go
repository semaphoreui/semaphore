package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_20_14 seeds an integration with everything that refers to
// it on the pre-workflow schema and checks the upgrade (a table rebuild on
// SQLite) keeps those rows and lets an integration point at a workflow.
func TestMigration_2_20_14(t *testing.T) {
	target := "2.20.11"
	store := InitConfigCreateTestStoreAt(&target)

	project, template := createTestTemplate(t, store)
	templateID := template.ID

	// The store already writes workflow_id, which the old schema lacks, so the
	// integration is seeded by hand.
	_, err := store.Sql().Exec(store.PrepareQuery(
		"insert into project__integration (project_id, name, template_id, auth_method, auth_header, searchable) values (?, ?, ?, '', '', 0)"),
		project.ID, "old hook", templateID)
	require.NoError(t, err)
	integrationID, err := store.Sql().SelectInt(store.PrepareQuery(
		"select id from project__integration where project_id = ? and name = ?"), project.ID, "old hook")
	require.NoError(t, err)
	integration := db.Integration{ID: int(integrationID), ProjectID: project.ID}

	_, err = store.CreateIntegrationAlias(db.IntegrationAlias{
		Alias: "alias-2-20-12", ProjectID: project.ID, IntegrationID: &integration.ID,
	})
	require.NoError(t, err)
	_, err = store.CreateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		Name: "m", IntegrationID: integration.ID, MatchType: db.IntegrationMatchHeader,
		Method: db.IntegrationMatchMethodEquals, Key: "X-Event", Value: "push",
	})
	require.NoError(t, err)
	_, err = store.CreateIntegrationExtractValue(project.ID, db.IntegrationExtractValue{
		Name: "v", IntegrationID: integration.ID, ValueSource: db.IntegrationExtractHeaderValue, Key: "X-Ref", Variable: "ref",
	})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{
		ProjectID: project.ID, TemplateID: templateID, IntegrationID: &integration.ID, Status: "success",
	}, 0)
	require.NoError(t, err)

	require.NoError(t, db.Migrate(store, nil))

	// Child rows survived the rebuild and still point at the same integration.
	for table, want := range map[string]int64{
		"project__integration_alias":         1,
		"project__integration_matcher":       1,
		"project__integration_extract_value": 1,
	} {
		count, err := store.Sql().SelectInt("select count(*) from "+table+" where integration_id = ?", integration.ID)
		require.NoError(t, err)
		assert.Equal(t, want, count, table)
	}
	migratedTask, err := store.GetTask(project.ID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, migratedTask.IntegrationID)
	assert.Equal(t, integration.ID, *migratedTask.IntegrationID)

	migrated, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	require.NotNil(t, migrated.TemplateID)
	assert.Equal(t, templateID, *migrated.TemplateID)
	assert.Nil(t, migrated.WorkflowID)

	// The foreign key on integration_id still cascades: deleting the
	// integration removes its children, as before the rebuild.
	workflowID := createTestWorkflow(t, store, project.ID, "wf")
	workflowHook, err := store.CreateIntegration(db.Integration{
		Name: "workflow hook", ProjectID: project.ID, WorkflowID: &workflowID,
	})
	require.NoError(t, err)
	assert.Nil(t, workflowHook.TemplateID)

	require.NoError(t, store.DeleteIntegration(project.ID, integration.ID))
	for _, table := range []string{"project__integration_alias", "project__integration_matcher", "project__integration_extract_value"} {
		count, err := store.Sql().SelectInt("select count(*) from "+table+" where integration_id = ?", integration.ID)
		require.NoError(t, err)
		assert.Zero(t, count, table)
	}
	orphanedTask, err := store.GetTask(project.ID, task.ID)
	require.NoError(t, err)
	assert.Nil(t, orphanedTask.IntegrationID, "task.integration_id must be set null, not cascade-deleted")
}

func TestMigration_2_20_14_MappingTables(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	user, err := store.CreateUserWithoutPassword(db.User{Username: "u", Name: "u", Email: "u@example.com"})
	require.NoError(t, err)
	role, err := store.CreateRole(db.Role{Slug: "dev", Name: "Dev", Permissions: db.CanRunProjectTasks, ProjectID: &project.ID})
	require.NoError(t, err)

	_, err = store.exec("insert into group_role_mapping (provider, group_name, project_id, role_id, created) values ('okta', 'devops', ?, ?, ?)",
		project.ID, role.ID, time.Now().UTC())
	require.NoError(t, err)
	mappingID := roleIDOf(t, store, "select id from group_role_mapping")
	_, err = store.exec("insert into project__user_mapped_role (project_id, user_id, mapping_id) values (?, ?, ?)", project.ID, user.ID, mappingID)
	require.NoError(t, err)

	_, err = store.exec("insert into group_role_mapping (provider, group_name, project_id, role_id, created) values ('okta', 'devops', ?, ?, ?)",
		project.ID, role.ID, time.Now().UTC())
	assert.Error(t, err, "a mapping exists once")
	_, err = store.exec("insert into group_role_mapping (provider, group_name, project_id, role_id, created) values ('okta', 'DevOps', ?, ?, ?)",
		project.ID, role.ID, time.Now().UTC())
	assert.NoError(t, err, "groups are case-sensitive")

	assert.Error(t, store.DeleteRole(role.ID), "a mapped role cannot be deleted")

	_, err = store.exec("delete from group_role_mapping where id=?", mappingID)
	require.NoError(t, err)
	left, err := store.Sql().SelectInt("select count(*) from project__user_mapped_role")
	require.NoError(t, err)
	assert.Zero(t, left, "mapped rows go with their mapping")
}

func TestMigration_2_20_14_Rollback(t *testing.T) {
	store := InitConfigCreateTestStore()
	require.NoError(t, db.Rollback(store, "2.20.13"))

	tables, err := store.Sql().SelectInt(`
		select
		  count(*)
		from sqlite_master
		where type = 'table'
		  and name in ('group_role_mapping', 'project__user_mapped_role')
	`)
	require.NoError(t, err)
	assert.Zero(t, tables)
}
