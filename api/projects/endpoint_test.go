package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// endpoint stores an endpoint without auth in the fixture's project.
func (f resourceFixture) endpoint(t *testing.T) db.Endpoint {
	t.Helper()
	endpoint, err := f.store.CreateEndpoint(db.Endpoint{Name: "hooks", ProjectID: f.project.ID})
	require.NoError(t, err)
	return endpoint
}

// integrationGroup stores an endpoint and a group on it, which an integration
// needs to belong to.
func (f resourceFixture) integrationGroup(t *testing.T) db.IntegrationGroup {
	t.Helper()
	endpoint := f.endpoint(t)
	group, err := f.store.CreateIntegrationGroup(db.IntegrationGroup{Name: "group", ProjectID: f.project.ID, EndpointID: endpoint.ID})
	require.NoError(t, err)
	return group
}

func TestEndpointEvents(t *testing.T) {
	f := newResourceFixture(t)

	r, rec := f.request(http.MethodPost, `{"name":"hooks","project_id":`+strconv.Itoa(f.project.ID)+`,"auth_method":"none"}`, nil, nil)
	w := httptest.NewRecorder()
	AddEndpoint(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := only(t, rec, audit.ResourceEndpointCreate)
	assert.Equal(t, audit.EndpointMetadata{AuthMethod: "none"}, created.Event.Metadata)
	endpointID := mustAtoi(t, created.Event.Target.ID)
	endpoint, err := f.store.GetEndpoint(f.project.ID, endpointID)
	require.NoError(t, err)

	r, rec = f.request(http.MethodPut, `{"id":`+strconv.Itoa(endpointID)+`,"name":"renamed","project_id":`+strconv.Itoa(f.project.ID)+`,"auth_method":"token","auth_header":"X-Token"}`, map[string]any{"endpoint": endpoint}, nil)
	w = httptest.NewRecorder()
	UpdateEndpoint(w, r)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	updated := only(t, rec, audit.ResourceEndpointUpdate)
	assert.Equal(t, audit.ResourceTarget(audit.TargetEndpoint, endpointID, "renamed"), updated.Event.Target)
	assert.Equal(t, audit.EndpointMetadata{AuthMethod: "token"}, updated.Event.Metadata)

	r, rec = f.request(http.MethodPost, "", map[string]any{"endpoint": endpoint}, nil)
	w = httptest.NewRecorder()
	AddEndpointAlias(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"url"`)
	alias := only(t, rec, audit.ResourceEndpointAliasCreate)
	assert.Empty(t, alias.Event.Target.Name, "the alias value is a bearer secret")
	assert.Equal(t, audit.EndpointPartMetadata{EndpointID: endpointID}, alias.Event.Metadata)

	aliasID := mustAtoi(t, alias.Event.Target.ID)
	r, rec = f.request(http.MethodDelete, "", map[string]any{"endpoint": endpoint}, map[string]string{"alias_id": strconv.Itoa(aliasID)})
	w = httptest.NewRecorder()
	RemoveEndpointAlias(w, r)
	require.Equal(t, http.StatusNoContent, w.Code)
	deletedAlias := only(t, rec, audit.ResourceEndpointAliasDelete)
	assert.Equal(t, audit.EndpointPartMetadata{EndpointID: endpointID}, deletedAlias.Event.Metadata)

	r, rec = f.request(http.MethodDelete, "", map[string]any{"endpoint": endpoint}, nil)
	w = httptest.NewRecorder()
	DeleteEndpoint(w, r)
	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, audit.ResourceTarget(audit.TargetEndpoint, endpointID, "hooks"), only(t, rec, audit.ResourceEndpointDelete).Event.Target)
}

func TestRemoveEndpointAlias_OfAnotherEndpointIsRefused(t *testing.T) {
	f := newResourceFixture(t)
	owner := f.endpoint(t)
	other, err := f.store.CreateEndpoint(db.Endpoint{Name: "other", ProjectID: f.project.ID})
	require.NoError(t, err)
	alias, err := f.store.CreateEndpointAlias(db.EndpointAlias{Alias: "a1b2c3", ProjectID: f.project.ID, EndpointID: owner.ID})
	require.NoError(t, err)

	r, rec := f.request(http.MethodDelete, "", map[string]any{"endpoint": other}, map[string]string{"alias_id": strconv.Itoa(alias.ID)})
	w := httptest.NewRecorder()
	RemoveEndpointAlias(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, rec.All())
	aliases, err := f.store.GetEndpointAliases(f.project.ID, owner.ID)
	require.NoError(t, err)
	assert.Len(t, aliases, 1, "the alias is kept")
}

func TestDeleteEndpoint_RefusedWhileGroupsUseIt(t *testing.T) {
	f := newResourceFixture(t)
	group := f.integrationGroup(t)
	endpoint, err := f.store.GetEndpoint(f.project.ID, group.EndpointID)
	require.NoError(t, err)

	r, rec := f.request(http.MethodDelete, "", map[string]any{"endpoint": endpoint}, nil)
	w := httptest.NewRecorder()
	DeleteEndpoint(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "used by integration groups")
	assert.Empty(t, rec.All(), "a refused delete is not recorded")

	r, _ = f.request(http.MethodGet, "", map[string]any{"endpoint": endpoint}, nil)
	w = httptest.NewRecorder()
	GetEndpointRefs(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"integration_groups"`)
	assert.Contains(t, w.Body.String(), `"name":"group"`)
}

func TestIntegrationGroupEvents(t *testing.T) {
	f := newResourceFixture(t)
	endpoint := f.endpoint(t)
	template := f.template(t)

	r, rec := f.request(http.MethodPost, `{"name":"deploys","project_id":`+strconv.Itoa(f.project.ID)+`,"endpoint_id":`+strconv.Itoa(endpoint.ID)+`}`, nil, nil)
	w := httptest.NewRecorder()
	AddIntegrationGroup(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := only(t, rec, audit.ResourceIntegrationGroupCreate)
	assert.Equal(t, audit.IntegrationGroupMetadata{EndpointID: endpoint.ID}, created.Event.Metadata)
	groupID := mustAtoi(t, created.Event.Target.ID)
	group, err := f.store.GetIntegrationGroup(f.project.ID, groupID)
	require.NoError(t, err)

	r, rec = f.request(http.MethodPut, `{"id":`+strconv.Itoa(groupID)+`,"name":"deploys","project_id":`+strconv.Itoa(f.project.ID)+`,"endpoint_id":`+strconv.Itoa(endpoint.ID)+`,"auth_override":true,"auth_method":"token","auth_header":"X-Token"}`, map[string]any{"integration_group": group}, nil)
	w = httptest.NewRecorder()
	UpdateIntegrationGroup(w, r)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	updated := only(t, rec, audit.ResourceIntegrationGroupUpdate)
	assert.Equal(t, audit.IntegrationGroupMetadata{EndpointID: endpoint.ID, AuthOverride: true, AuthMethod: "token"}, updated.Event.Metadata)

	// An integration of the group goes with it.
	integration, err := f.store.CreateIntegration(db.Integration{Name: "hook", ProjectID: f.project.ID, GroupID: groupID, TemplateID: &template.ID})
	require.NoError(t, err)
	r, _ = f.request(http.MethodGet, "", map[string]any{"integration_group": group}, nil)
	w = httptest.NewRecorder()
	GetIntegrationGroupIntegrations(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"hook"`)

	r, rec = f.request(http.MethodDelete, "", map[string]any{"integration_group": group}, nil)
	w = httptest.NewRecorder()
	DeleteIntegrationGroup(w, r)
	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, audit.ResourceTarget(audit.TargetIntegrationGroup, groupID, "deploys"), only(t, rec, audit.ResourceIntegrationGroupDelete).Event.Target)
	_, err = f.store.GetIntegration(f.project.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestAddIntegrationGroup_EndpointOfAnotherProjectIsRefused(t *testing.T) {
	f := newResourceFixture(t)
	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	foreign, err := f.store.CreateEndpoint(db.Endpoint{Name: "foreign", ProjectID: other.ID})
	require.NoError(t, err)

	r, rec := f.request(http.MethodPost, `{"name":"deploys","project_id":`+strconv.Itoa(f.project.ID)+`,"endpoint_id":`+strconv.Itoa(foreign.ID)+`}`, nil, nil)
	w := httptest.NewRecorder()
	AddIntegrationGroup(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "Endpoint does not exist in the project")
	assert.Empty(t, rec.All())
}

func TestAddIntegration_GroupOfAnotherProjectIsRefused(t *testing.T) {
	f := newResourceFixture(t)
	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	foreignEndpoint, err := f.store.CreateEndpoint(db.Endpoint{Name: "foreign", ProjectID: other.ID})
	require.NoError(t, err)
	foreignGroup, err := f.store.CreateIntegrationGroup(db.IntegrationGroup{Name: "foreign", ProjectID: other.ID, EndpointID: foreignEndpoint.ID})
	require.NoError(t, err)

	r, rec := f.request(http.MethodPost, `{"name":"hook","project_id":`+strconv.Itoa(f.project.ID)+`,"group_id":`+strconv.Itoa(foreignGroup.ID)+`,"template_id":`+strconv.Itoa(f.template(t).ID)+`}`, nil, nil)
	w := httptest.NewRecorder()
	integrationController(noFeatures).AddIntegration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "Integration group does not exist in the project")
	assert.Empty(t, rec.All())
}
