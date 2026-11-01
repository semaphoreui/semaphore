package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestProjectWithKey(t *testing.T, store *SqlDb) (db.Project, db.AccessKey) {
	t.Helper()
	project, err := store.CreateProject(db.Project{Name: "endpoint project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "hook secret",
		Type:      db.AccessKeyString,
		String:    "s3cret",
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	return project, key
}

func TestEndpoint_CRUD(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, key := createTestProjectWithKey(t, store)

	created, err := store.CreateEndpoint(db.Endpoint{
		Name:        "GitHub",
		Description: "pushes from GitHub",
		ProjectID:   project.ID,
		IntegrationAuth: db.IntegrationAuth{
			AuthMethod:   db.IntegrationAuthGitHub,
			AuthSecretID: &key.ID,
		},
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	loaded, err := store.GetEndpoint(project.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, loaded)

	loaded.Name = "GitHub Enterprise"
	loaded.AuthMethod = db.IntegrationAuthToken
	loaded.AuthHeader = "X-Token"
	require.NoError(t, store.UpdateEndpoint(loaded))

	updated, err := store.GetEndpoint(project.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, loaded, updated)

	list, err := store.GetEndpoints(project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Equal(t, []db.Endpoint{updated}, list)

	require.NoError(t, store.DeleteEndpoint(project.ID, created.ID))
	_, err = store.GetEndpoint(project.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestEndpoint_Validation(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestProjectWithKey(t, store)

	_, err := store.CreateEndpoint(db.Endpoint{ProjectID: project.ID})
	assertValidationError(t, err, "No Name set for endpoint")

	other, otherKey := createTestProjectWithKey(t, store)
	_, err = store.CreateEndpoint(db.Endpoint{
		Name: "foreign key", ProjectID: project.ID,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &otherKey.ID},
	})
	assertValidationError(t, err, "Auth secret does not exist in the project")

	created, err := store.CreateEndpoint(db.Endpoint{Name: "own", ProjectID: project.ID})
	require.NoError(t, err)
	created.AuthSecretID = &otherKey.ID
	err = store.UpdateEndpoint(created)
	assertValidationError(t, err, "Auth secret does not exist in the project")

	_, err = store.GetEndpoint(other.ID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound, "an endpoint is not visible from another project")

	renamed := created
	renamed.AuthSecretID = nil
	renamed.ProjectID = other.ID
	assert.ErrorIs(t, store.UpdateEndpoint(renamed), db.ErrNotFound, "an endpoint cannot be updated from another project")
	assert.ErrorIs(t, store.DeleteEndpoint(other.ID, created.ID), db.ErrNotFound)
}

func TestEndpoint_DeleteRefusedWhileGroupsUseIt(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestProjectWithKey(t, store)
	endpoint, group := createTestIntegrationGroup(t, store, project.ID)

	refs, err := store.GetEndpointRefs(project.ID, endpoint.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{{ID: group.ID, Name: group.Name}}, refs.IntegrationGroups)

	err = store.DeleteEndpoint(project.ID, endpoint.ID)
	assertValidationError(t, err, "Endpoint is used by integration groups")

	require.NoError(t, store.DeleteIntegrationGroup(project.ID, group.ID))
	require.NoError(t, store.DeleteEndpoint(project.ID, endpoint.ID))
}

func TestEndpointAlias(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestProjectWithKey(t, store)
	endpoint, err := store.CreateEndpoint(db.Endpoint{Name: "hooks", ProjectID: project.ID})
	require.NoError(t, err)

	first, err := store.CreateEndpointAlias(db.EndpointAlias{Alias: "alias-one", ProjectID: project.ID, EndpointID: endpoint.ID})
	require.NoError(t, err)
	second, err := store.CreateEndpointAlias(db.EndpointAlias{Alias: "alias-two", ProjectID: project.ID, EndpointID: endpoint.ID})
	require.NoError(t, err)

	_, err = store.CreateEndpointAlias(db.EndpointAlias{Alias: "alias-one", ProjectID: project.ID, EndpointID: endpoint.ID})
	assert.Error(t, err, "an alias is unique across the instance")

	aliases, err := store.GetEndpointAliases(project.ID, endpoint.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.EndpointAlias{first, second}, aliases)

	resolved, err := store.GetEndpointByAlias("alias-two")
	require.NoError(t, err)
	assert.Equal(t, endpoint, resolved)

	_, err = store.GetEndpointByAlias("alias-none")
	assert.ErrorIs(t, err, db.ErrNotFound)

	require.NoError(t, store.DeleteEndpointAlias(project.ID, endpoint.ID, first.ID))
	_, err = store.GetEndpointByAlias("alias-one")
	assert.ErrorIs(t, err, db.ErrNotFound)

	// The aliases go with their endpoint.
	require.NoError(t, store.DeleteEndpoint(project.ID, endpoint.ID))
	_, err = store.GetEndpointByAlias("alias-two")
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestEndpointAlias_WrongProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestProjectWithKey(t, store)
	endpoint, err := store.CreateEndpoint(db.Endpoint{Name: "hooks", ProjectID: project.ID})
	require.NoError(t, err)
	alias, err := store.CreateEndpointAlias(db.EndpointAlias{Alias: "alias-one", ProjectID: project.ID, EndpointID: endpoint.ID})
	require.NoError(t, err)

	other, _ := createTestProjectWithKey(t, store)

	_, err = store.CreateEndpointAlias(db.EndpointAlias{Alias: "alias-x", ProjectID: other.ID, EndpointID: endpoint.ID})
	assert.ErrorIs(t, err, db.ErrNotFound, "a project cannot add a URL to another project's endpoint")

	foreign, err := store.GetEndpointAliases(other.ID, endpoint.ID)
	require.NoError(t, err)
	assert.Empty(t, foreign)

	assert.ErrorIs(t, store.DeleteEndpointAlias(other.ID, endpoint.ID, alias.ID), db.ErrNotFound)
	otherEndpoint, err := store.CreateEndpoint(db.Endpoint{Name: "other hooks", ProjectID: project.ID})
	require.NoError(t, err)
	assert.ErrorIs(t, store.DeleteEndpointAlias(project.ID, otherEndpoint.ID, alias.ID), db.ErrNotFound,
		"an alias is deleted only through its own endpoint")

	kept, err := store.GetEndpointAliases(project.ID, endpoint.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.EndpointAlias{alias}, kept)
}

func TestAccessKeyRefs_EndpointsAndGroups(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, key := createTestProjectWithKey(t, store)

	endpoint, err := store.CreateEndpoint(db.Endpoint{
		Name: "hooks", ProjectID: project.ID,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthSecretID: &key.ID},
	})
	require.NoError(t, err)
	group, err := store.CreateIntegrationGroup(db.IntegrationGroup{
		Name: "own auth", ProjectID: project.ID, EndpointID: endpoint.ID, AuthOverride: true,
		IntegrationAuth: db.IntegrationAuth{AuthMethod: db.IntegrationAuthHmac, AuthSecretID: &key.ID},
	})
	require.NoError(t, err)
	_, err = store.CreateIntegrationGroup(db.IntegrationGroup{Name: "inherits", ProjectID: project.ID, EndpointID: endpoint.ID})
	require.NoError(t, err)

	refs, err := store.GetAccessKeyRefs(project.ID, key.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{{ID: endpoint.ID, Name: endpoint.Name}}, refs.Endpoints)
	assert.Equal(t, []db.ObjectReferrer{{ID: group.ID, Name: group.Name}}, refs.IntegrationGroups)
}
