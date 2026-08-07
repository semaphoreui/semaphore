package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestIntegration(t *testing.T, store *SqlDb) (db.Project, db.Integration) {
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

	integration, err := store.CreateIntegration(db.Integration{
		Name:       "test integration",
		ProjectID:  project.ID,
		TemplateID: template.ID,
	})
	require.NoError(t, err)

	return project, integration
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
