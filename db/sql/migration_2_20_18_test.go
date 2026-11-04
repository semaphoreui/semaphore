package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

<<<<<<< HEAD
func TestMigration_2_20_18_LinksLegacyGroupSecrets(t *testing.T) {
	before := "2.20.16"
	store := InitConfigCreateTestStoreAt(&before)

	project, err := store.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	envID, err := store.insert("id", "insert into project__environment (project_id, name, json, env) values (?, 'E', '{}', '{}')", project.ID)
	require.NoError(t, err)
	otherEnvID, err := store.insert("id", "insert into project__environment (project_id, name, json, env) values (?, 'F', '{}', '{}')", project.ID)
	require.NoError(t, err)

	insertKey := func(name string, owner db.AccessKeyOwner, env *int) int {
		id, err := store.insert("id",
			"insert into access_key (name, type, project_id, environment_id, owner, secret) values (?, 'string', ?, ?, ?, 'x')",
			name, project.ID, env, owner)
		require.NoError(t, err)
		return id
	}
	plain := insertKey("DB_PASSWORD", db.AccessKeyEnvironment, &envID)
	prefixed := insertKey("var.token", db.AccessKeyVariable, &envID)
	insertKey("token", db.AccessKeyVariable, &envID) // duplicate of the prefixed one, newer, wins
	dupWinner, err := store.Sql().SelectInt(store.PrepareQuery("select max(id) from access_key"))
	require.NoError(t, err)
	otherGroup := insertKey("token", db.AccessKeyVariable, &otherEnvID)
	insertKey("shared", db.AccessKeyShared, nil)
	insertKey("vault-key", db.AccessKeySecretStorage, &envID)
=======
func TestMigration_2_20_18_GivesReadToRolesAndGrants(t *testing.T) {
	before := "2.20.17"
	store := InitConfigCreateTestStoreAt(&before)
	projectID, repositoryID := newTemplateTestProject(t, store)
	template, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "deploy", Playbook: "site.yml"})
	require.NoError(t, err)
	workflowID := insertWorkflow(t, store, projectID, "release")
	ops, err := store.CreateRole(db.Role{Slug: "ops", Name: "Ops", Permissions: 1 | 16 | 64, ProjectID: &projectID})
	require.NoError(t, err)
	nobody, err := store.CreateRole(db.Role{Slug: "nobody", Name: "Nobody", ProjectID: &projectID})
	require.NoError(t, err)
	templateGrant, err := store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: template.ID, RoleID: nobody.ID, Permissions: 1})
	require.NoError(t, err)
	workflowGrant, err := store.CreateWorkflowRole(db.WorkflowRolePerm{ProjectID: projectID, WorkflowID: workflowID, RoleID: nobody.ID, Permissions: 16})
	require.NoError(t, err)
	emptyGrant, err := store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: template.ID, RoleID: ops.ID, Permissions: 0})
	require.NoError(t, err)
>>>>>>> develop

	target := "2.20.18"
	require.NoError(t, db.Migrate(store, &target))

<<<<<<< HEAD
	secrets, err := store.GetEnvironmentSecretKeys(project.ID, envID)
	require.NoError(t, err)
	require.Len(t, secrets, 2, "one env secret and one var secret, the duplicate name collapsed")

	assert.Equal(t, db.EnvironmentSecretEnv, secrets[0].Type)
	assert.Equal(t, "DB_PASSWORD", secrets[0].Name)
	assert.Equal(t, plain, secrets[0].KeyID)
	assert.True(t, secrets[0].Owned)

	assert.Equal(t, db.EnvironmentSecretVar, secrets[1].Type)
	assert.Equal(t, "token", secrets[1].Name, "the legacy var. prefix is not part of the secret name")
	assert.Equal(t, int(dupWinner), secrets[1].KeyID, "the newer key wins a duplicate name")
	assert.NotEqual(t, prefixed, secrets[1].KeyID)

	prefixedKey, err := store.GetAccessKey(project.ID, prefixed)
	require.NoError(t, err)
	assert.Equal(t, "var.token", prefixedKey.Name, "key names are left untouched")

	other, err := store.GetEnvironmentSecretKeys(project.ID, otherEnvID)
	require.NoError(t, err)
	require.Len(t, other, 1)
	assert.Equal(t, otherGroup, other[0].KeyID, "the same name in another group is a separate secret")

	total, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from project__environment_secret"))
	require.NoError(t, err)
	assert.Equal(t, int64(3), total, "shared and storage keys are not group secrets")
=======
	for slug, mask := range map[string]db.ProjectUserPermission{"owner": 255, "guest": 128, "ops": 1 | 16 | 64 | 128, "nobody": 128} {
		role, err := store.GetProjectOrSharedRoleBySlug(projectID, slug)
		require.NoError(t, err)
		assert.Equal(t, mask, role.Permissions, slug)
	}
	storedTemplateGrant, err := store.GetTemplateRole(projectID, template.ID, templateGrant.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ProjectUserPermission(1|128), storedTemplateGrant.Permissions)
	storedWorkflowGrant, err := store.GetWorkflowRole(projectID, workflowID, workflowGrant.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ProjectUserPermission(16|128), storedWorkflowGrant.Permissions)
	storedEmptyGrant, err := store.GetTemplateRole(projectID, template.ID, emptyGrant.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ProjectUserPermission(0), storedEmptyGrant.Permissions, "a grant without a verb stays empty")

	require.NoError(t, db.Rollback(store, "2.20.17"))
	role, err := store.GetProjectOrSharedRoleBySlug(projectID, "ops")
	require.NoError(t, err)
	assert.Equal(t, ops.Permissions, role.Permissions, "rollback keeps the old bits only")
	storedTemplateGrant, err = store.GetTemplateRole(projectID, template.ID, templateGrant.ID)
	require.NoError(t, err)
	assert.Equal(t, db.ProjectUserPermission(1), storedTemplateGrant.Permissions)
>>>>>>> develop
}
