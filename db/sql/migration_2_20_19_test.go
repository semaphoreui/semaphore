package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oldIntegration is a project__integration row of the pre-2.20.19 schema,
// which the store of this version cannot write.
type oldIntegration struct {
	name       string
	templateID *int
	workflowID *int
	authMethod string
	secretID   *int
	authHeader string
	searchable bool
}

func seedOldIntegration(t *testing.T, store *SqlDb, projectID int, row oldIntegration) int {
	t.Helper()
	_, err := store.Sql().Exec(store.PrepareQuery(
		"insert into project__integration (project_id, name, template_id, workflow_id, auth_method, auth_secret_id, auth_header, searchable) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?)"),
		projectID, row.name, row.templateID, row.workflowID, row.authMethod, row.secretID, row.authHeader, row.searchable)
	require.NoError(t, err)
	id, err := store.Sql().SelectInt(store.PrepareQuery(
		"select id from project__integration where project_id = ? and name = ?"), projectID, row.name)
	require.NoError(t, err)
	return int(id)
}

func seedOldAlias(t *testing.T, store *SqlDb, projectID int, integrationID *int, alias string) {
	t.Helper()
	_, err := store.Sql().Exec(store.PrepareQuery(
		"insert into project__integration_alias (project_id, integration_id, alias) values (?, ?, ?)"),
		projectID, integrationID, alias)
	require.NoError(t, err)
}

func endpointNamed(t *testing.T, endpoints []db.Endpoint, name string) db.Endpoint {
	t.Helper()
	for _, e := range endpoints {
		if e.Name == name {
			return e
		}
	}
	require.Failf(t, "endpoint not found", "no endpoint named %q", name)
	return db.Endpoint{}
}

func aliasesOf(t *testing.T, store *SqlDb, endpoint db.Endpoint) []string {
	t.Helper()
	aliases, err := store.GetEndpointAliases(endpoint.ProjectID, endpoint.ID)
	require.NoError(t, err)
	res := make([]string, 0, len(aliases))
	for _, a := range aliases {
		res = append(res, a.Alias)
	}
	return res
}

// legacyIntegrations is the pre-redesign fixture: one project exercising every
// rule of the migration, one with a searchable integration but no project
// alias, one with neither.
type legacyIntegrations struct {
	project, searchableOnly, ownOnly db.Project
	key                              db.AccessKey
	template                         db.Template
	searchableToken                  int // searchable, token auth, a matcher, a dead alias of its own
	searchableNone                   int // searchable, auth 'none' (the column default)
	ownAliases                       int // not searchable, hmac auth, two aliases, an extract value
	noAliases                        int // not searchable, no aliases
	workflowHook                     int // not searchable, runs a workflow, one alias
	remoteSearchable                 int // in searchableOnly
	remoteOwn                        int // in ownOnly
	taskID                           int
}

func seedLegacyIntegrations(t *testing.T, store *SqlDb) legacyIntegrations {
	t.Helper()
	f := legacyIntegrations{}
	f.project, f.template = createTestTemplate(t, store)
	var err error
	f.key, err = store.CreateAccessKey(db.AccessKey{Name: "hook secret", Type: db.AccessKeyString, String: "s3cret", ProjectID: &f.project.ID})
	require.NoError(t, err)
	workflowID := createTestWorkflow(t, store, f.project.ID, "wf")

	f.searchableToken = seedOldIntegration(t, store, f.project.ID, oldIntegration{
		name: "searchable token", templateID: &f.template.ID, authMethod: "token", secretID: &f.key.ID, authHeader: "X-Token", searchable: true})
	f.searchableNone = seedOldIntegration(t, store, f.project.ID, oldIntegration{
		name: "searchable none", templateID: &f.template.ID, authMethod: "none", searchable: true})
	f.ownAliases = seedOldIntegration(t, store, f.project.ID, oldIntegration{
		name: "own aliases", templateID: &f.template.ID, authMethod: "hmac", secretID: &f.key.ID, authHeader: "X-Sig"})
	f.noAliases = seedOldIntegration(t, store, f.project.ID, oldIntegration{
		name: "no aliases", templateID: &f.template.ID, authMethod: ""})
	f.workflowHook = seedOldIntegration(t, store, f.project.ID, oldIntegration{
		name: "workflow hook", workflowID: &workflowID, authMethod: ""})

	seedOldAlias(t, store, f.project.ID, nil, "project-one")
	seedOldAlias(t, store, f.project.ID, nil, "project-two")
	seedOldAlias(t, store, f.project.ID, &f.searchableToken, "dead-alias")
	seedOldAlias(t, store, f.project.ID, &f.ownAliases, "own-one")
	seedOldAlias(t, store, f.project.ID, &f.ownAliases, "own-two")
	seedOldAlias(t, store, f.project.ID, &f.workflowHook, "workflow-alias")

	_, err = store.Sql().Exec(store.PrepareQuery(
		"insert into project__integration_matcher (name, integration_id, match_type, `method`, body_data_type, `key`, `value`) values ('m', ?, 'header', 'equals', '', 'X-Event', 'push')"),
		f.searchableToken)
	require.NoError(t, err)
	_, err = store.Sql().Exec(store.PrepareQuery(
		"insert into project__integration_extract_value (name, integration_id, value_source, body_data_type, `key`, `variable`, variable_type) values ('v', ?, 'header', '', 'X-Ref', 'ref', 'environment')"),
		f.ownAliases)
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: f.project.ID, TemplateID: f.template.ID, IntegrationID: &f.ownAliases, Status: "success"}, 0)
	require.NoError(t, err)
	f.taskID = task.ID

	f.searchableOnly, err = store.CreateProject(db.Project{Name: "searchable only"})
	require.NoError(t, err)
	remoteTemplate := createTestTemplateIn(t, store, f.searchableOnly.ID)
	f.remoteSearchable = seedOldIntegration(t, store, f.searchableOnly.ID, oldIntegration{
		name: "remote searchable", templateID: &remoteTemplate, authMethod: "", searchable: true})

	f.ownOnly, err = store.CreateProject(db.Project{Name: "own only"})
	require.NoError(t, err)
	ownTemplate := createTestTemplateIn(t, store, f.ownOnly.ID)
	f.remoteOwn = seedOldIntegration(t, store, f.ownOnly.ID, oldIntegration{
		name: "remote own", templateID: &ownTemplate, authMethod: ""})

	return f
}

// createTestTemplateIn adds a bash template to an existing project.
func createTestTemplateIn(t *testing.T, store *SqlDb, projectID int) int {
	t.Helper()
	key, err := store.CreateAccessKey(db.AccessKey{Name: "k", Type: db.AccessKeyNone, ProjectID: &projectID})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{Name: "r", ProjectID: projectID, GitURL: "git@example.com:t/t.git", GitBranch: "main", SSHKeyID: key.ID})
	require.NoError(t, err)
	template, err := store.CreateTemplate(db.Template{Name: "t", ProjectID: projectID, RepositoryID: repo.ID, Playbook: "t.yml", App: db.AppBash})
	require.NoError(t, err)
	return template.ID
}

func TestMigration_2_20_19_MovesAliasesAndAuthToEndpointsAndGroups(t *testing.T) {
	before := "2.20.16"
	store := InitConfigCreateTestStoreAt(&before)
	f := seedLegacyIntegrations(t, store)

	require.NoError(t, db.Migrate(store, nil))

	// Project aliases → a Default endpoint without auth, holding exactly them.
	endpoints, err := store.GetEndpoints(f.project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, endpoints, 4, "Default + one per non-searchable integration")
	def := endpointNamed(t, endpoints, "Default")
	assert.Equal(t, db.IntegrationAuth{AuthMethod: db.IntegrationAuthNone, AuthHeader: ""}, def.IntegrationAuth)
	assert.ElementsMatch(t, []string{"project-one", "project-two"}, aliasesOf(t, store, def))

	// A non-searchable integration → its own endpoint with its auth and aliases.
	own := endpointNamed(t, endpoints, "own aliases")
	assert.Equal(t, db.IntegrationAuth{AuthMethod: db.IntegrationAuthHmac, AuthSecretID: &f.key.ID, AuthHeader: "X-Sig"}, own.IntegrationAuth)
	assert.ElementsMatch(t, []string{"own-one", "own-two"}, aliasesOf(t, store, own))
	assert.Empty(t, aliasesOf(t, store, endpointNamed(t, endpoints, "no aliases")))
	assert.Equal(t, []string{"workflow-alias"}, aliasesOf(t, store, endpointNamed(t, endpoints, "workflow hook")))

	// The alias of a searchable integration matched nothing before and is gone.
	_, err = store.GetEndpointByAlias("dead-alias")
	assert.ErrorIs(t, err, db.ErrNotFound)

	// One group per integration, named after it.
	groups, err := store.GetIntegrationGroups(f.project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, groups, 5)
	groupOf := func(integrationID int) db.IntegrationGroup {
		integration, err := store.GetIntegration(f.project.ID, integrationID)
		require.NoError(t, err)
		group, err := store.GetIntegrationGroup(f.project.ID, integration.GroupID)
		require.NoError(t, err)
		assert.Equal(t, integration.Name, group.Name)
		return group
	}

	// Searchable with auth → on Default, overriding with its own auth.
	tokenGroup := groupOf(f.searchableToken)
	assert.Equal(t, def.ID, tokenGroup.EndpointID)
	assert.True(t, tokenGroup.AuthOverride)
	assert.Equal(t, db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &f.key.ID, AuthHeader: "X-Token"}, tokenGroup.IntegrationAuth)

	// Searchable with 'none' (the column default) → on Default, inheriting.
	noneGroup := groupOf(f.searchableNone)
	assert.Equal(t, def.ID, noneGroup.EndpointID)
	assert.False(t, noneGroup.AuthOverride)
	assert.Equal(t, db.IntegrationAuth{}, noneGroup.IntegrationAuth)

	// Non-searchable → on its own endpoint, inheriting.
	for _, id := range []int{f.ownAliases, f.noAliases, f.workflowHook} {
		group := groupOf(id)
		assert.Equal(t, endpointNamed(t, endpoints, group.Name).ID, group.EndpointID, group.Name)
		assert.False(t, group.AuthOverride, group.Name)
		assert.Equal(t, db.IntegrationAuth{}, group.IntegrationAuth, group.Name)
	}

	// Two groups now run from Default.
	onDefault, err := store.GetIntegrationGroupsByEndpoint(f.project.ID, def.ID)
	require.NoError(t, err)
	assert.Len(t, onDefault, 2)

	// Children, the task link and the targets are untouched.
	matchers, err := store.GetIntegrationMatchers(f.project.ID, db.RetrieveQueryParams{}, f.searchableToken)
	require.NoError(t, err)
	assert.Len(t, matchers, 1)
	values, err := store.GetIntegrationExtractValues(f.project.ID, db.RetrieveQueryParams{}, f.ownAliases)
	require.NoError(t, err)
	assert.Len(t, values, 1)
	task, err := store.GetTask(f.project.ID, f.taskID)
	require.NoError(t, err)
	require.NotNil(t, task.IntegrationID)
	assert.Equal(t, f.ownAliases, *task.IntegrationID)
	workflowHook, err := store.GetIntegration(f.project.ID, f.workflowHook)
	require.NoError(t, err)
	assert.NotNil(t, workflowHook.WorkflowID)
	assert.Nil(t, workflowHook.TemplateID)

	// A project with a searchable integration but no project alias still gets
	// a Default endpoint, with no alias, so the integration keeps a home.
	remoteEndpoints, err := store.GetEndpoints(f.searchableOnly.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, remoteEndpoints, 1)
	assert.Equal(t, "Default", remoteEndpoints[0].Name)
	assert.Empty(t, aliasesOf(t, store, remoteEndpoints[0]))

	// A project with only non-searchable integrations gets no Default endpoint.
	ownEndpoints, err := store.GetEndpoints(f.ownOnly.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, ownEndpoints, 1)
	assert.Equal(t, "remote own", ownEndpoints[0].Name)

	// The old table and the working columns are gone.
	_, err = store.Sql().SelectInt("select count(*) from project__integration_alias")
	assert.Error(t, err, "project__integration_alias must be dropped")
	for _, column := range []string{"auth_method", "auth_secret_id", "auth_header", "searchable"} {
		_, err = store.Sql().SelectInt("select count(" + column + ") from project__integration")
		assert.Error(t, err, column+" must be dropped from project__integration")
	}
	for _, table := range []string{"project__endpoint", "project__integration_group"} {
		_, err = store.Sql().SelectInt("select count(legacy_integration_id) from " + table)
		assert.Error(t, err, "legacy_integration_id must be dropped from "+table)
	}

	// The new store works on the migrated data: a group needs its endpoint,
	// and deleting a group takes its integration along.
	err = store.DeleteEndpoint(f.project.ID, def.ID)
	assertValidationError(t, err, "Endpoint is used by integration groups")
	require.NoError(t, store.DeleteIntegrationGroup(f.project.ID, tokenGroup.ID))
	_, err = store.GetIntegration(f.project.ID, f.searchableToken)
	assert.ErrorIs(t, err, db.ErrNotFound)
	matchers, err = store.GetIntegrationMatchers(f.project.ID, db.RetrieveQueryParams{}, f.searchableToken)
	require.NoError(t, err)
	assert.Empty(t, matchers, "matchers cascade with their integration")
}

func TestMigration_2_20_19_Rollback(t *testing.T) {
	before := "2.20.16"
	store := InitConfigCreateTestStoreAt(&before)
	f := seedLegacyIntegrations(t, store)
	require.NoError(t, db.Migrate(store, nil))

	require.NoError(t, db.Rollback(store, "2.20.16"))

	type oldRow struct {
		ID           int    `db:"id"`
		AuthMethod   string `db:"auth_method"`
		AuthSecretID *int   `db:"auth_secret_id"`
		AuthHeader   string `db:"auth_header"`
		Searchable   bool   `db:"searchable"`
	}
	var rows []oldRow
	_, err := store.Sql().Select(&rows, store.PrepareQuery(
		"select id, auth_method, auth_secret_id, auth_header, searchable from project__integration where project_id = ? order by id"), f.project.ID)
	require.NoError(t, err)
	byID := map[int]oldRow{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	require.Len(t, byID, 5)

	// Two integrations shared the Default endpoint: both are searchable again,
	// each with the auth its group checked.
	assert.Equal(t, oldRow{ID: f.searchableToken, AuthMethod: "token", AuthSecretID: &f.key.ID, AuthHeader: "X-Token", Searchable: true}, byID[f.searchableToken])
	assert.Equal(t, oldRow{ID: f.searchableNone, AuthMethod: "", AuthHeader: "", Searchable: true}, byID[f.searchableNone])
	// An endpoint with one integration gives it back its own auth and aliases.
	assert.Equal(t, oldRow{ID: f.ownAliases, AuthMethod: "hmac", AuthSecretID: &f.key.ID, AuthHeader: "X-Sig", Searchable: false}, byID[f.ownAliases])
	assert.Equal(t, oldRow{ID: f.noAliases, AuthMethod: "", Searchable: false}, byID[f.noAliases])
	assert.Equal(t, oldRow{ID: f.workflowHook, AuthMethod: "", Searchable: false}, byID[f.workflowHook])

	type oldAlias struct {
		Alias         string `db:"alias"`
		IntegrationID *int   `db:"integration_id"`
	}
	var aliases []oldAlias
	_, err = store.Sql().Select(&aliases, store.PrepareQuery(
		"select alias, integration_id from project__integration_alias where project_id = ? order by alias"), f.project.ID)
	require.NoError(t, err)
	assert.Equal(t, []oldAlias{
		{"own-one", &f.ownAliases},
		{"own-two", &f.ownAliases},
		{"project-one", nil},
		{"project-two", nil},
		{"workflow-alias", &f.workflowHook},
	}, aliases, "the dead alias of the searchable integration is not recoverable")

	for _, table := range []string{"project__endpoint", "project__endpoint_alias", "project__integration_group"} {
		_, err = store.Sql().SelectInt("select count(*) from " + table)
		assert.Error(t, err, table+" must be dropped")
	}
	_, err = store.Sql().SelectInt("select count(group_id) from project__integration")
	assert.Error(t, err, "group_id must be dropped from project__integration")

	// The rolled-back schema accepts the migration again.
	require.NoError(t, db.Migrate(store, nil))
	def, err := store.GetEndpointByAlias("project-one")
	require.NoError(t, err)
	assert.Equal(t, "Default", def.Name)
}
