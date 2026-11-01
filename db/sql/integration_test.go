package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestTemplate creates a project with one bash template in it.
func createTestTemplate(t *testing.T, store *SqlDb) (db.Project, db.Template) {
	t.Helper()
	project, err := store.CreateProject(db.Project{Name: "test project"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "test key",
		Type:      db.AccessKeyNone,
		ProjectID: &project.ID,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		Name:      "test repo",
		ProjectID: project.ID,
		GitURL:    "git@example.com:test/test.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	template, err := store.CreateTemplate(db.Template{
		Name:         "test template",
		ProjectID:    project.ID,
		RepositoryID: repo.ID,
		Playbook:     "test.yml",
		App:          db.AppBash,
	})
	require.NoError(t, err)

	return project, template
}

// createTestIntegrationGroup creates an endpoint and a group on it, which
// every integration needs to belong to.
func createTestIntegrationGroup(t *testing.T, store *SqlDb, projectID int) (db.Endpoint, db.IntegrationGroup) {
	t.Helper()
	endpoint, err := store.CreateEndpoint(db.Endpoint{Name: "test endpoint", ProjectID: projectID})
	require.NoError(t, err)

	group, err := store.CreateIntegrationGroup(db.IntegrationGroup{
		Name:       "test group",
		ProjectID:  projectID,
		EndpointID: endpoint.ID,
	})
	require.NoError(t, err)

	return endpoint, group
}

func createTestIntegration(t *testing.T, store *SqlDb) (db.Project, db.Integration) {
	t.Helper()
	project, template := createTestTemplate(t, store)
	_, group := createTestIntegrationGroup(t, store, project.ID)

	integration, err := store.CreateIntegration(db.Integration{
		Name:       "test integration",
		ProjectID:  project.ID,
		GroupID:    group.ID,
		TemplateID: &template.ID,
	})
	require.NoError(t, err)

	return project, integration
}

func TestCreateIntegration_GroupMustBelongToProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, existing := createTestIntegration(t, store)

	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)
	_, otherGroup := createTestIntegrationGroup(t, store, other.ID)

	_, err = store.CreateIntegration(db.Integration{
		Name: "foreign group", ProjectID: project.ID, GroupID: otherGroup.ID, TemplateID: existing.TemplateID,
	})
	assertValidationError(t, err, "Integration group does not exist in the project")

	moved := existing
	moved.GroupID = otherGroup.ID
	err = store.UpdateIntegration(moved)
	assertValidationError(t, err, "Integration group does not exist in the project")

	loaded, err := store.GetIntegration(project.ID, existing.ID)
	require.NoError(t, err)
	assert.Equal(t, existing.GroupID, loaded.GroupID, "a refused update must not change the row")
}

func TestGetIntegrationsByGroup(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, first := createTestIntegration(t, store)
	_, otherGroup := createTestIntegrationGroup(t, store, project.ID)

	second, err := store.CreateIntegration(db.Integration{
		Name: "second", ProjectID: project.ID, GroupID: otherGroup.ID, TemplateID: first.TemplateID,
		TaskParams: &db.TaskParams{Message: "from the hook"},
	})
	require.NoError(t, err)

	inFirst, err := store.GetIntegrationsByGroup(project.ID, first.GroupID, false)
	require.NoError(t, err)
	require.Len(t, inFirst, 1)
	assert.Equal(t, first.ID, inFirst[0].ID)

	inSecond, err := store.GetIntegrationsByGroup(project.ID, otherGroup.ID, true)
	require.NoError(t, err)
	require.Len(t, inSecond, 1)
	assert.Equal(t, second.ID, inSecond[0].ID)
	require.NotNil(t, inSecond[0].TaskParams)
	assert.Equal(t, "from the hook", inSecond[0].TaskParams.Message)

	foreign, err := store.GetIntegrationsByGroup(project.ID+1, first.GroupID, false)
	require.NoError(t, err)
	assert.Empty(t, foreign, "a group's integrations are not visible from another project")
}

func TestUpdateIntegrationMatcher(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)

	matcher, err := store.CreateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		Name:          "original",
		IntegrationID: integration.ID,
		MatchType:     db.IntegrationMatchHeader,
		Method:        db.IntegrationMatchMethodEquals,
		Key:           "X-Event",
		Value:         "push",
	})
	require.NoError(t, err)

	matcher.Name = "updated"
	matcher.MatchType = db.IntegrationMatchBody
	matcher.Method = db.IntegrationMatchMethodContains
	matcher.BodyDataType = db.IntegrationBodyDataJSON
	matcher.Key = "action"
	matcher.Value = "opened"

	err = store.UpdateIntegrationMatcher(project.ID, matcher)
	require.NoError(t, err)

	found, err := store.GetIntegrationMatcher(project.ID, matcher.ID, integration.ID)
	require.NoError(t, err)
	assert.Equal(t, matcher, found)
}

func TestUpdateIntegrationMatcher_InvalidMatcher(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)

	err := store.UpdateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		ID:            1,
		Name:          "no match type",
		IntegrationID: integration.ID,
		Key:           "key",
		Value:         "value",
	})
	assert.ErrorContains(t, err, "No Match Type set")
}

func TestUpdateIntegrationMatcher_WrongProject(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)

	matcher, err := store.CreateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		Name:          "original",
		IntegrationID: integration.ID,
		MatchType:     db.IntegrationMatchHeader,
		Method:        db.IntegrationMatchMethodEquals,
		Key:           "X-Event",
		Value:         "push",
	})
	require.NoError(t, err)

	updated := matcher
	updated.Name = "updated"

	err = store.UpdateIntegrationMatcher(project.ID+1, updated)
	assert.ErrorIs(t, err, db.ErrNotFound)

	found, err := store.GetIntegrationMatcher(project.ID, matcher.ID, integration.ID)
	require.NoError(t, err)
	assert.Equal(t, matcher, found, "matcher must not be updated for another project")
}

func TestGetIntegrationRefs(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)

	matcher, err := store.CreateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		Name:          "matcher",
		IntegrationID: integration.ID,
		MatchType:     db.IntegrationMatchHeader,
		Method:        db.IntegrationMatchMethodEquals,
		Key:           "X-Event",
		Value:         "push",
	})
	require.NoError(t, err)

	value, err := store.CreateIntegrationExtractValue(project.ID, db.IntegrationExtractValue{
		Name:          "value",
		IntegrationID: integration.ID,
		ValueSource:   db.IntegrationExtractBodyValue,
		BodyDataType:  db.IntegrationBodyDataJSON,
		Key:           "commit",
		Variable:      "commit_hash",
	})
	require.NoError(t, err)

	refs, err := store.GetIntegrationRefs(project.ID, integration.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{{ID: matcher.ID, Name: matcher.Name}}, refs.IntegrationMatchers)
	assert.Equal(t, []db.ObjectReferrer{{ID: value.ID, Name: value.Name}}, refs.IntegrationExtractValues)

	refs, err = store.GetIntegrationRefs(project.ID+1, integration.ID)
	require.NoError(t, err)
	assert.Empty(t, refs.IntegrationMatchers, "refs must not be visible for another project")
	assert.Empty(t, refs.IntegrationExtractValues, "refs must not be visible for another project")
}

func TestIntegrationMatcher_WrongProject(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)

	matcher, err := store.CreateIntegrationMatcher(project.ID, db.IntegrationMatcher{
		Name:          "matcher",
		IntegrationID: integration.ID,
		MatchType:     db.IntegrationMatchHeader,
		Method:        db.IntegrationMatchMethodEquals,
		Key:           "X-Event",
		Value:         "push",
	})
	require.NoError(t, err)

	wrongProjectID := project.ID + 1

	_, err = store.CreateIntegrationMatcher(wrongProjectID, db.IntegrationMatcher{
		Name:          "matcher2",
		IntegrationID: integration.ID,
		MatchType:     db.IntegrationMatchHeader,
		Method:        db.IntegrationMatchMethodEquals,
		Key:           "X-Event",
		Value:         "push",
	})
	assert.ErrorIs(t, err, db.ErrNotFound)

	_, err = store.GetIntegrationMatcher(wrongProjectID, matcher.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	matchers, err := store.GetIntegrationMatchers(wrongProjectID, db.RetrieveQueryParams{}, integration.ID)
	require.NoError(t, err)
	assert.Empty(t, matchers)

	err = store.DeleteIntegrationMatcher(wrongProjectID, matcher.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	found, err := store.GetIntegrationMatcher(project.ID, matcher.ID, integration.ID)
	require.NoError(t, err)
	assert.Equal(t, matcher, found, "matcher must not be deleted for another project")
}

func TestIntegrationExtractValue_WrongProject(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)

	value, err := store.CreateIntegrationExtractValue(project.ID, db.IntegrationExtractValue{
		Name:          "value",
		IntegrationID: integration.ID,
		ValueSource:   db.IntegrationExtractBodyValue,
		BodyDataType:  db.IntegrationBodyDataJSON,
		Key:           "commit",
		Variable:      "commit_hash",
	})
	require.NoError(t, err)

	wrongProjectID := project.ID + 1

	_, err = store.CreateIntegrationExtractValue(wrongProjectID, db.IntegrationExtractValue{
		Name:          "value2",
		IntegrationID: integration.ID,
		ValueSource:   db.IntegrationExtractBodyValue,
		BodyDataType:  db.IntegrationBodyDataJSON,
		Key:           "branch",
		Variable:      "branch_name",
	})
	assert.ErrorIs(t, err, db.ErrNotFound)

	_, err = store.GetIntegrationExtractValue(wrongProjectID, value.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	values, err := store.GetIntegrationExtractValues(wrongProjectID, db.RetrieveQueryParams{}, integration.ID)
	require.NoError(t, err)
	assert.Empty(t, values)

	updated := value
	updated.Name = "updated"
	err = store.UpdateIntegrationExtractValue(wrongProjectID, updated)
	assert.ErrorIs(t, err, db.ErrNotFound)

	err = store.DeleteIntegrationExtractValue(wrongProjectID, value.ID, integration.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	found, err := store.GetIntegrationExtractValue(project.ID, value.ID, integration.ID)
	require.NoError(t, err)
	assert.Equal(t, value, found, "extract value must not be modified for another project")
}
