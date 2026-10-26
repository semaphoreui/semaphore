package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	target := "2.20.18"
	require.NoError(t, db.Migrate(store, &target))

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
}
