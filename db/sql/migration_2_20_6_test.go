package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedHasDestination(t *testing.T) {
	assert.False(t, seedHasDestination(seed{enabled: true, typ: db.AlertTypeTelegram}))
	assert.True(t, seedHasDestination(seed{enabled: true, typ: db.AlertTypeTelegram, chat: "1"}))
	assert.False(t, seedHasDestination(seed{enabled: true, typ: db.AlertTypeSlack}))
	assert.False(t, seedHasDestination(seed{enabled: true, typ: db.AlertTypeSlack, url: "javascript:x"}))
	assert.True(t, seedHasDestination(seed{enabled: true, typ: db.AlertTypeSlack, url: "https://hooks.example/x"}))
	assert.True(t, seedHasDestination(seed{enabled: true, typ: db.AlertTypeGotify}))
	assert.False(t, seedHasDestination(seed{enabled: false, typ: db.AlertTypeGotify}))
}

func TestMigration_2_20_6_CopiesInstanceChannelsOnce(t *testing.T) {
	store := InitConfigCreateTestStore()
	util.Config.TelegramAlert = true
	util.Config.TelegramChat = "cfg-chat"
	util.Config.SlackAlert = true
	util.Config.SlackUrl = "https://hooks.example/slack"
	util.Config.GotifyAlert = true
	util.Config.GotifyUrl = "https://gotify.example/push"
	util.Config.GotifyToken = "instance-gotify-token"

	chat := "project-chat"
	alertingProject, err := store.CreateProject(db.Project{
		Name:      "alerting",
		Alert:     true,
		AlertChat: &chat,
	})
	require.NoError(t, err)

	quietProject, err := store.CreateProject(db.Project{Name: "quiet"})
	require.NoError(t, err)

	_, alertingRepo := newTemplateTestProjectAt(t, store, alertingProject.ID)
	_, quietRepo := newTemplateTestProjectAt(t, store, quietProject.ID)

	alertingTpl, err := store.CreateTemplate(db.Template{
		ProjectID:    alertingProject.ID,
		RepositoryID: alertingRepo,
		Name:         "nightly",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	silencedTpl, err := store.CreateTemplate(db.Template{
		ProjectID:             alertingProject.ID,
		RepositoryID:          alertingRepo,
		Name:                  "silent",
		Playbook:              "site.yml",
		SuppressSuccessAlerts: true,
		SuppressErrorAlerts:   true,
	})
	require.NoError(t, err)

	quietTpl, err := store.CreateTemplate(db.Template{
		ProjectID:    quietProject.ID,
		RepositoryID: quietRepo,
		Name:         "manual",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	tx, err := store.Sql().Begin()
	require.NoError(t, err)
	require.NoError(t, migration_2_20_6{db: store}.PostApply(tx))
	require.NoError(t, tx.Commit())

	alerts, err := store.GetAlerts(alertingProject.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, alerts, 3)

	var telegram db.Alert
	var gotify db.Alert
	for _, alert := range alerts {
		if alert.Type == db.AlertTypeTelegram {
			telegram = alert
		}
		if alert.Type == db.AlertTypeGotify {
			gotify = alert
		}
	}
	require.Equal(t, db.AlertTypeTelegram, telegram.Type)
	require.NotNil(t, telegram.ChatID)
	assert.Equal(t, "project-chat", *telegram.ChatID)
	assert.True(t, telegram.IsDefault)
	require.Equal(t, db.AlertTypeGotify, gotify.Type)
	assert.Nil(t, gotify.URL)
	assert.Nil(t, gotify.Token)
	assert.True(t, gotify.IsDefault)

	alertingTpl, err = store.GetTemplate(alertingProject.ID, alertingTpl.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeDefault, alertingTpl.AlertMode)
	assert.Empty(t, alertingTpl.AlertIDs)

	silencedTpl, err = store.GetTemplate(alertingProject.ID, silencedTpl.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeIDs, silencedTpl.AlertMode)
	assert.Empty(t, silencedTpl.AlertIDs)

	defaults, err := store.GetDefaultAlertIDs(alertingProject.ID)
	require.NoError(t, err)
	assert.Len(t, defaults, 3)

	quietAlerts, err := store.GetAlerts(quietProject.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Len(t, quietAlerts, 3)
	assert.False(t, quietAlerts[0].IsDefault)

	quietTpl, err = store.GetTemplate(quietProject.ID, quietTpl.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeIDs, quietTpl.AlertMode)
	assert.Empty(t, quietTpl.AlertIDs)

	quietDefaults, err := store.GetDefaultAlertIDs(quietProject.ID)
	require.NoError(t, err)
	assert.Empty(t, quietDefaults)
}

func TestMigration_2_20_6_SkipsIncompleteDestinations(t *testing.T) {
	store := InitConfigCreateTestStore()
	util.Config.TelegramAlert = true
	util.Config.TelegramChat = ""
	util.Config.SlackAlert = true
	util.Config.SlackUrl = ""
	util.Config.GotifyAlert = true
	util.Config.GotifyUrl = "https://gotify.example/push"
	util.Config.GotifyToken = "instance-gotify-token"

	project, err := store.CreateProject(db.Project{Name: "partial"})
	require.NoError(t, err)

	tx, err := store.Sql().Begin()
	require.NoError(t, err)
	require.NoError(t, migration_2_20_6{db: store}.PostApply(tx))
	require.NoError(t, tx.Commit())

	alerts, err := store.GetAlerts(project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, db.AlertTypeGotify, alerts[0].Type)
	assert.Nil(t, alerts[0].URL)
	assert.Nil(t, alerts[0].Token)
}

func newTemplateTestProjectAt(t *testing.T, store *SqlDb, projectID int) (int, int) {
	t.Helper()

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &projectID,
		Type:      db.AccessKeyNone,
	})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: projectID,
		Name:      "repo",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	return projectID, repo.ID
}
