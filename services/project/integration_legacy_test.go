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

func TestConvertLegacyIntegrations(t *testing.T) {
	key := "hook secret"
	conv := ConvertLegacyIntegrations([]string{"project-one", "project-two"}, []LegacyIntegration{
		{Name: "searchable token", Searchable: true, AuthMethod: "token", AuthHeader: "X-Token", AuthSecret: &key, Aliases: []string{"dead-alias"}},
		{Name: "searchable none", Searchable: true, AuthMethod: "none"},
		{Name: "own aliases", AuthMethod: "hmac", AuthHeader: "X-Sig", AuthSecret: &key, Aliases: []string{"own-one", "own-two"}},
		{Name: "no aliases", AuthMethod: ""},
		{Name: "Default", AuthMethod: "none", Aliases: []string{"clash"}},
	})

	require.Len(t, conv.Endpoints, 4, "Default + one per non-searchable integration")
	def := conv.Endpoints[0]
	assert.Equal(t, "Default", def.Name)
	assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthNone), def.AuthMethod)
	assert.Equal(t, []string{"project-one", "project-two"}, def.Aliases)

	own := conv.Endpoints[1]
	assert.Equal(t, "own aliases", own.Name)
	assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthHmac), own.AuthMethod)
	assert.Equal(t, "X-Sig", own.AuthHeader)
	assert.Equal(t, &key, own.AuthSecretName)
	assert.Equal(t, []string{"own-one", "own-two"}, own.Aliases)

	assert.Equal(t, "no aliases", conv.Endpoints[2].Name)
	assert.Empty(t, conv.Endpoints[2].Aliases)
	assert.Equal(t, "Default (2)", conv.Endpoints[3].Name, "an integration named Default does not collide with the Default endpoint")
	assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthNone), conv.Endpoints[3].AuthMethod, "'none' becomes no auth")

	require.Len(t, conv.Groups, 5)
	assert.Equal(t, []string{"searchable token", "searchable none", "own aliases", "no aliases", "Default"}, conv.GroupOf)

	token := conv.Groups[0]
	assert.Equal(t, "Default", token.Endpoint)
	assert.True(t, token.AuthOverride)
	assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthToken), token.AuthMethod)
	assert.Equal(t, "X-Token", token.AuthHeader)
	assert.Equal(t, &key, token.AuthSecretName)

	none := conv.Groups[1]
	assert.Equal(t, "Default", none.Endpoint)
	assert.False(t, none.AuthOverride, "'none' auth on a searchable integration inherits")
	assert.Nil(t, none.AuthSecretName)

	for n, name := range []string{"own aliases", "no aliases", "Default (2)"} {
		group := conv.Groups[2+n]
		assert.Equal(t, name, group.Endpoint)
		assert.False(t, group.AuthOverride)
	}

	// Nothing to reach the project through: no Default endpoint.
	only := ConvertLegacyIntegrations(nil, []LegacyIntegration{{Name: "own", Aliases: []string{"a"}}})
	require.Len(t, only.Endpoints, 1)
	assert.Equal(t, "own", only.Endpoints[0].Name)

	// A searchable integration alone still needs a Default endpoint, with no alias.
	searchable := ConvertLegacyIntegrations(nil, []LegacyIntegration{{Name: "s", Searchable: true}})
	require.Len(t, searchable.Endpoints, 1)
	assert.Equal(t, "Default", searchable.Endpoints[0].Name)
	assert.Empty(t, searchable.Endpoints[0].Aliases)
}

// legacyBackup is a project backup as a version before 2.21.11 wrote it:
// aliases and auth on the integrations, project aliases at the top level.
const legacyBackup = `{
  "meta": {"name": "legacy", "alert": false, "max_parallel_tasks": 0},
  "templates": [{"name": "Deploy", "playbook": "deploy.yml", "repository": "Repo", "app": "bash", "type": ""}],
  "repositories": [{"name": "Repo", "git_url": "git@example.com:test/test", "git_branch": "master", "ssh_key": "none"}],
  "keys": [{"name": "none", "type": "none"}, {"name": "hook secret", "type": "string"}],
  "views": [], "inventories": [], "environments": [], "schedules": [], "secret_storages": [], "roles": [], "runners": [], "workflows": [],
  "integrations": [
    {"name": "searchable token", "template": "Deploy", "auth_method": "token", "auth_header": "X-Token", "auth_secret": "hook secret", "searchable": true, "aliases": ["dead-alias"],
     "matchers": [{"name": "m", "match_type": "header", "method": "equals", "body_data_type": "", "key": "X-Event", "value": "push"}], "extract_values": []},
    {"name": "own aliases", "template": "Deploy", "auth_method": "hmac", "auth_header": "X-Sig", "auth_secret": "hook secret", "searchable": false, "aliases": ["own-one", "own-two"],
     "matchers": [], "extract_values": []}
  ],
  "integration_aliases": ["project-one"]
}`

// A backup written before 2.21.11 restores to the layout the database
// migration produces: a Default endpoint for the project aliases, an endpoint
// per integration that had its own, a group per integration.
func TestRestore_LegacyBackupBecomesEndpointsAndGroups(t *testing.T) {
	util.Config = &util.ConfigType{TmpPath: "/tmp", Apps: map[string]util.App{"bash": {Active: true}}}
	store := sql.InitConfigCreateTestStore()

	backup := &BackupFormat{}
	require.NoError(t, backup.Unmarshal(legacyBackup))
	require.NoError(t, backup.Verify())

	require.Len(t, backup.Endpoints, 2)
	require.Len(t, backup.IntegrationGroups, 2)
	assert.Nil(t, backup.IntegrationAliases, "the project aliases moved to the Default endpoint")

	user, err := store.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "legacy", Name: "legacy", Email: "legacy@example.com", Admin: true}})
	require.NoError(t, err)
	project, err := backup.Restore(user, store, proFactory.NewWorkflowStore(store))
	require.NoError(t, err)

	def, err := store.GetEndpointByAlias("project-one")
	require.NoError(t, err)
	assert.Equal(t, project.ID, def.ProjectID)
	assert.Equal(t, "Default", def.Name)
	assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthNone), def.AuthMethod)

	own, err := store.GetEndpointByAlias("own-two")
	require.NoError(t, err)
	assert.Equal(t, "own aliases", own.Name)
	assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthHmac), own.AuthMethod)
	assert.Equal(t, "X-Sig", own.AuthHeader)
	assert.NotNil(t, own.AuthSecretID)

	_, err = store.GetEndpointByAlias("dead-alias")
	assert.ErrorIs(t, err, db.ErrNotFound, "the alias of a searchable integration matched nothing and is dropped")

	integrations, err := store.GetIntegrations(project.ID, db.RetrieveQueryParams{}, false)
	require.NoError(t, err)
	require.Len(t, integrations, 2)
	for _, integration := range integrations {
		group, err := store.GetIntegrationGroup(project.ID, integration.GroupID)
		require.NoError(t, err)
		assert.Equal(t, integration.Name, group.Name)
		switch integration.Name {
		case "searchable token":
			assert.Equal(t, def.ID, group.EndpointID)
			assert.True(t, group.AuthOverride)
			assert.Equal(t, db.IntegrationAuthMethod(db.IntegrationAuthToken), group.AuthMethod)
			assert.Equal(t, "X-Token", group.AuthHeader)
			assert.NotNil(t, group.AuthSecretID)
			matchers, err := store.GetIntegrationMatchers(project.ID, db.RetrieveQueryParams{}, integration.ID)
			require.NoError(t, err)
			assert.Len(t, matchers, 1)
		case "own aliases":
			assert.Equal(t, own.ID, group.EndpointID)
			assert.False(t, group.AuthOverride)
		}
	}
}
