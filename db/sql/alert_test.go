package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAlerts_EmptySlice(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "empty-alerts"})
	require.NoError(t, err)

	alerts, err := store.GetAlerts(project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.NotNil(t, alerts)
	assert.Empty(t, alerts)
}

func TestUpdateAlert_PreservesGotifyToken(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "tok"})
	require.NoError(t, err)

	token := "secret-token"
	created, err := store.CreateAlert(db.Alert{
		ProjectID: project.ID,
		Name:      "Gotify",
		Type:      db.AlertTypeGotify,
		Enabled:   true,
		Token:     &token,
	})
	require.NoError(t, err)

	created.Token = nil
	created.Name = "Gotify prod"
	require.NoError(t, store.UpdateAlert(created))

	loaded, err := store.GetAlert(project.ID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.Token)
	assert.Equal(t, "secret-token", *loaded.Token)
	assert.Equal(t, "Gotify prod", loaded.Name)
}

func TestCreateAlert_GotifyURLRequiresToken(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "gotify-url"})
	require.NoError(t, err)

	url := "https://gotify.example.com"
	_, err = store.CreateAlert(db.Alert{
		ProjectID: project.ID,
		Name:      "Gotify",
		Type:      db.AlertTypeGotify,
		Enabled:   true,
		URL:       &url,
	})
	assert.Error(t, err)
}

func TestAlertCRUDAndRefs(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	chatID := "12345"
	created, err := store.CreateAlert(db.Alert{
		ProjectID: projectID,
		Name:      "Ops Telegram",
		Type:      db.AlertTypeTelegram,
		Enabled:   true,
		ChatID:    &chatID,
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	_, err = store.CreateAlert(db.Alert{
		ProjectID: projectID,
		Name:      "Ops Telegram",
		Type:      db.AlertTypeSlack,
		Enabled:   true,
	})
	assert.Error(t, err)

	alerts, err := store.GetAlerts(projectID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, "Ops Telegram", alerts[0].Name)

	tpl, err := store.CreateTemplate(db.Template{
		ProjectID:      projectID,
		RepositoryID:   repositoryID,
		Name:           "deploy",
		Playbook:       "site.yml",
		AlertIDs:       []int{created.ID},
		AlertOnSuccess: db.BoolPtr(true),
		AlertOnError:   db.BoolPtr(false),
	})
	require.NoError(t, err)
	assert.Equal(t, []int{created.ID}, tpl.AlertIDs)
	require.NotNil(t, tpl.AlertOnSuccess)
	assert.True(t, *tpl.AlertOnSuccess)
	require.NotNil(t, tpl.AlertOnError)
	assert.False(t, *tpl.AlertOnError)

	refs, err := store.GetAlertRefs(projectID, created.ID)
	require.NoError(t, err)
	require.Len(t, refs.Templates, 1)
	assert.Equal(t, tpl.ID, refs.Templates[0].ID)
	assert.Empty(t, refs.Schedules)

	err = store.DeleteAlert(projectID, created.ID)
	assert.Error(t, err)

	require.NoError(t, store.UpdateTemplateAlerts(projectID, tpl.ID, nil))
	require.NoError(t, store.DeleteAlert(projectID, created.ID))

	_, err = store.GetAlert(projectID, created.ID)
	assert.Error(t, err)
}

func TestTaskAlertSnapshotRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	tpl, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "job",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	snap := db.AlertSnapshot{AlertIDs: []int{4, 9}, OnSuccess: false, OnError: true}
	created, err := store.CreateTask(db.Task{
		ProjectID:     projectID,
		TemplateID:    tpl.ID,
		Status:        task_logger.TaskWaitingStatus,
		AlertSnapshot: &snap,
	}, 0)
	require.NoError(t, err)

	loaded, err := store.GetTask(projectID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.AlertSnapshot)
	assert.Equal(t, []int{4, 9}, loaded.AlertSnapshot.AlertIDs)
	assert.False(t, loaded.AlertSnapshot.OnSuccess)
	assert.True(t, loaded.AlertSnapshot.OnError)
}

func TestUpdateTemplate_OmitsAlertIDsLeavesBindings(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	hook := "https://hooks.example.com/x"
	alert, err := store.CreateAlert(db.Alert{
		ProjectID: projectID,
		Name:      "Ops",
		Type:      db.AlertTypeSlack,
		Enabled:   true,
		URL:       &hook,
	})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "bound",
		Playbook:     "site.yml",
		AlertIDs:     []int{alert.ID},
	})
	require.NoError(t, err)

	tpl.AlertIDs = nil
	tpl.Name = "bound-renamed"
	require.NoError(t, store.UpdateTemplate(tpl))

	loaded, err := store.GetTemplate(projectID, tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, "bound-renamed", loaded.Name)
	assert.Equal(t, []int{alert.ID}, loaded.AlertIDs)
}

func TestClaimAlertSend_ExactlyOnce(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	tpl, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "job",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		ProjectID:  projectID,
		TemplateID: tpl.ID,
		Status:     task_logger.TaskWaitingStatus,
	}, 0)
	require.NoError(t, err)

	first, err := store.ClaimAlertSend(task.ID, 7, "error")
	require.NoError(t, err)
	assert.True(t, first)

	second, err := store.ClaimAlertSend(task.ID, 7, "error")
	require.NoError(t, err)
	assert.False(t, second)

	otherEvent, err := store.ClaimAlertSend(task.ID, 7, "success")
	require.NoError(t, err)
	assert.True(t, otherEvent)
}
