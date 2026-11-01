package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationGroup_CRUD(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, key := createTestProjectWithKey(t, store)
	endpoint, err := store.CreateEndpoint(db.Endpoint{Name: "hooks", ProjectID: project.ID})
	require.NoError(t, err)

	created, err := store.CreateIntegrationGroup(db.IntegrationGroup{
		Name: "deploys", ProjectID: project.ID, EndpointID: endpoint.ID, AuthOverride: true,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &key.ID, AuthHeader: "X-Token"},
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	loaded, err := store.GetIntegrationGroup(project.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, loaded)

	loaded.Name = "deploys and tests"
	loaded.AuthHeader = "X-Deploy-Token"
	require.NoError(t, store.UpdateIntegrationGroup(loaded))

	updated, err := store.GetIntegrationGroup(project.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, loaded, updated)

	list, err := store.GetIntegrationGroups(project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Equal(t, []db.IntegrationGroup{updated}, list)

	require.NoError(t, store.DeleteIntegrationGroup(project.ID, created.ID))
	_, err = store.GetIntegrationGroup(project.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestIntegrationGroup_InheritingGroupKeepsNoAuth(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, key := createTestProjectWithKey(t, store)
	endpoint, err := store.CreateEndpoint(db.Endpoint{Name: "hooks", ProjectID: project.ID})
	require.NoError(t, err)

	created, err := store.CreateIntegrationGroup(db.IntegrationGroup{
		Name: "inherits", ProjectID: project.ID, EndpointID: endpoint.ID,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &key.ID, AuthHeader: "X-Token"},
	})
	require.NoError(t, err)
	assert.Equal(t, db.IntegrationAuth{}, created.IntegrationAuth, "auth is cleared when the group does not override")

	loaded, err := store.GetIntegrationGroup(project.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, db.IntegrationAuth{}, loaded.IntegrationAuth)

	// Switching the override off on update clears it too.
	loaded.AuthOverride = true
	loaded.IntegrationAuth = db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &key.ID}
	require.NoError(t, store.UpdateIntegrationGroup(loaded))
	loaded.AuthOverride = false
	require.NoError(t, store.UpdateIntegrationGroup(loaded))

	reloaded, err := store.GetIntegrationGroup(project.ID, created.ID)
	require.NoError(t, err)
	assert.False(t, reloaded.AuthOverride)
	assert.Equal(t, db.IntegrationAuth{}, reloaded.IntegrationAuth)
}

func TestIntegrationGroup_Validation(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestProjectWithKey(t, store)
	endpoint, err := store.CreateEndpoint(db.Endpoint{Name: "hooks", ProjectID: project.ID})
	require.NoError(t, err)

	other, otherKey := createTestProjectWithKey(t, store)
	otherEndpoint, err := store.CreateEndpoint(db.Endpoint{Name: "foreign hooks", ProjectID: other.ID})
	require.NoError(t, err)

	_, err = store.CreateIntegrationGroup(db.IntegrationGroup{ProjectID: project.ID, EndpointID: endpoint.ID})
	assertValidationError(t, err, "No Name set for integration group")

	_, err = store.CreateIntegrationGroup(db.IntegrationGroup{Name: "no endpoint", ProjectID: project.ID})
	assertValidationError(t, err, "Integration group must use an endpoint")

	_, err = store.CreateIntegrationGroup(db.IntegrationGroup{Name: "foreign endpoint", ProjectID: project.ID, EndpointID: otherEndpoint.ID})
	assertValidationError(t, err, "Endpoint does not exist in the project")

	_, err = store.CreateIntegrationGroup(db.IntegrationGroup{
		Name: "foreign key", ProjectID: project.ID, EndpointID: endpoint.ID, AuthOverride: true,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &otherKey.ID},
	})
	assertValidationError(t, err, "Auth secret does not exist in the project")

	// Without the override the foreign key is dropped rather than refused: it is never used.
	inheriting, err := store.CreateIntegrationGroup(db.IntegrationGroup{
		Name: "inherits", ProjectID: project.ID, EndpointID: endpoint.ID,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &otherKey.ID},
	})
	require.NoError(t, err)
	assert.Nil(t, inheriting.AuthSecretID)

	moved := inheriting
	moved.EndpointID = otherEndpoint.ID
	err = store.UpdateIntegrationGroup(moved)
	assertValidationError(t, err, "Endpoint does not exist in the project")

	_, err = store.GetIntegrationGroup(other.ID, inheriting.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
	assert.ErrorIs(t, store.DeleteIntegrationGroup(other.ID, inheriting.ID), db.ErrNotFound)
}

func TestIntegrationGroup_ByEndpoint(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestProjectWithKey(t, store)
	first, err := store.CreateEndpoint(db.Endpoint{Name: "first", ProjectID: project.ID})
	require.NoError(t, err)
	second, err := store.CreateEndpoint(db.Endpoint{Name: "second", ProjectID: project.ID})
	require.NoError(t, err)

	a, err := store.CreateIntegrationGroup(db.IntegrationGroup{Name: "a", ProjectID: project.ID, EndpointID: first.ID})
	require.NoError(t, err)
	b, err := store.CreateIntegrationGroup(db.IntegrationGroup{Name: "b", ProjectID: project.ID, EndpointID: first.ID})
	require.NoError(t, err)
	_, err = store.CreateIntegrationGroup(db.IntegrationGroup{Name: "c", ProjectID: project.ID, EndpointID: second.ID})
	require.NoError(t, err)

	onFirst, err := store.GetIntegrationGroupsByEndpoint(project.ID, first.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.IntegrationGroup{a, b}, onFirst)

	foreign, err := store.GetIntegrationGroupsByEndpoint(project.ID+1, first.ID)
	require.NoError(t, err)
	assert.Empty(t, foreign)
}

func TestIntegrationGroup_DeleteTakesIntegrationsAndTaskParams(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)

	integration.TaskParams = &db.TaskParams{Message: "hello"}
	require.NoError(t, store.UpdateIntegration(integration))
	withParams, err := store.GetIntegration(project.ID, integration.ID)
	require.NoError(t, err)
	require.NotNil(t, withParams.TaskParamsID)

	matcher, err := store.CreateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		Name: "m", IntegrationID: integration.ID, MatchType: db.IntegrationMatchHeader,
		Method: db.IntegrationMatchMethodEquals, Key: "X-Event", Value: "push",
	})
	require.NoError(t, err)

	refs, err := store.GetIntegrationGroupRefs(project.ID, integration.GroupID)
	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{{ID: integration.ID, Name: integration.Name}}, refs.Integrations)

	require.NoError(t, store.DeleteIntegrationGroup(project.ID, integration.GroupID))

	_, err = store.GetIntegration(project.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
	_, err = store.GetIntegrationMatcher(project.ID, matcher.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
	var orphan db.TaskParams
	err = store.getObject(project.ID, db.TaskParamsProps, *withParams.TaskParamsID, &orphan)
	assert.ErrorIs(t, err, db.ErrNotFound, "the integration's task params go with it")
}

func TestDeleteProject_WithEndpointsAndGroups(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)
	endpoint, err := store.GetEndpoints(project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, endpoint, 1)
	_, err = store.CreateEndpointAlias(db.EndpointAlias{Alias: "gone-with-project", ProjectID: project.ID, EndpointID: endpoint[0].ID})
	require.NoError(t, err)

	require.NoError(t, store.DeleteProject(project.ID))

	_, err = store.GetEndpointByAlias("gone-with-project")
	assert.ErrorIs(t, err, db.ErrNotFound)
	for _, table := range []string{"project__endpoint", "project__integration_group", "project__integration"} {
		count, err := store.Sql().SelectInt(store.PrepareQuery("select count(*) from "+table+" where project_id = ?"), project.ID)
		require.NoError(t, err)
		assert.Zero(t, count, table)
	}
	_, err = store.GetIntegration(project.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}
