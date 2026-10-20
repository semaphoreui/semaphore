package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAlertTestProject(t *testing.T, store *SqlDb) (projectID int, repositoryID int) {
	t.Helper()
	return newTemplateTestProject(t, store)
}

func strPtr(s string) *string {
	return &s
}

func TestAlerts_CRUD(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newAlertTestProject(t, store)

	empty, err := store.GetAlerts(projectID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.NotNil(t, empty)
	assert.Empty(t, empty)

	created, err := store.CreateAlert(db.Alert{
		ProjectID: projectID,
		Name:      " Ops chat ",
		Type:      "telegram",
		Enabled:   true,
		IsDefault: true,
		Events:    db.AlertEvents{db.AlertEventError},
		ChatID:    strPtr("123"),
		ThreadID:  strPtr(""),
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Ops chat", created.Name)

	loaded, err := store.GetAlert(projectID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertEvents{db.AlertEventError}, loaded.Events)
	assert.Nil(t, loaded.ThreadID)
	require.NotNil(t, loaded.ChatID)
	assert.Equal(t, "123", *loaded.ChatID)

	loaded.Name = "Ops"
	loaded.Events = nil
	loaded.IsDefault = false
	require.NoError(t, store.UpdateAlert(loaded))

	updated, err := store.GetAlert(projectID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ops", updated.Name)
	assert.Nil(t, updated.Events)
	assert.False(t, updated.IsDefault)

	require.NoError(t, store.DeleteAlert(projectID, created.ID))
	_, err = store.GetAlert(projectID, created.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestAlerts_SetActive(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newAlertTestProject(t, store)

	alert, err := store.CreateAlert(db.Alert{ProjectID: projectID, Name: "Ops", Type: "slack", URL: strPtr("https://hooks.example/a"), Enabled: true})
	require.NoError(t, err)

	require.NoError(t, store.SetAlertActive(projectID, alert.ID, false))
	loaded, err := store.GetAlert(projectID, alert.ID)
	require.NoError(t, err)
	assert.False(t, loaded.Enabled)

	require.NoError(t, store.SetAlertActive(projectID, alert.ID, true))
	loaded, err = store.GetAlert(projectID, alert.ID)
	require.NoError(t, err)
	assert.True(t, loaded.Enabled)

	// Setting the current value again must not be reported as "not found":
	// MySQL counts only changed rows, so this must not rely on RowsAffected.
	require.NoError(t, store.SetAlertActive(projectID, alert.ID, true))

	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	assert.ErrorIs(t, store.SetAlertActive(other.ID, alert.ID, false), db.ErrNotFound)
}

func TestAlerts_NameMustBeUniquePerProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newAlertTestProject(t, store)

	_, err := store.CreateAlert(db.Alert{ProjectID: projectID, Name: "Ops", Type: "slack", URL: strPtr("https://hooks.example/a")})
	require.NoError(t, err)

	_, err = store.CreateAlert(db.Alert{ProjectID: projectID, Name: "Ops", Type: "slack", URL: strPtr("https://hooks.example/b")})
	assert.ErrorContains(t, err, "already exists")
}

func TestAlerts_KeyAndParamsRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newAlertTestProject(t, store)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &projectID, Name: "smtp", Type: db.AccessKeyLoginPassword})
	require.NoError(t, err)

	created, err := store.CreateAlert(db.Alert{
		ProjectID: projectID,
		Name:      "Mail",
		Type:      "email",
		KeyID:     &key.ID,
		Params:    db.MapStringAnyField{"smtp_host": "mail.example", "smtp_port": "587", "smtp_tls": true},
	})
	require.NoError(t, err)

	loaded, err := store.GetAlert(projectID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.KeyID)
	assert.Equal(t, key.ID, *loaded.KeyID)
	assert.Equal(t, "mail.example", loaded.ParamString("smtp_host"))
	assert.Equal(t, "587", loaded.ParamString("smtp_port"))
	assert.True(t, loaded.ParamBool("smtp_tls"))

	// The key is listed as used by the alert, so the key store refuses deletion.
	refs, err := store.GetAccessKeyRefs(projectID, key.ID)
	require.NoError(t, err)
	require.Len(t, refs.Alerts, 1)
	assert.Equal(t, "Mail", refs.Alerts[0].Name)

	// A key of another project is rejected.
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	foreign, err := store.CreateAccessKey(db.AccessKey{ProjectID: &other.ID, Name: "foreign", Type: db.AccessKeyString})
	require.NoError(t, err)
	loaded.KeyID = &foreign.ID
	assert.ErrorContains(t, store.UpdateAlert(loaded), "does not belong")

	loaded.KeyID = nil
	loaded.Params = nil
	require.NoError(t, store.UpdateAlert(loaded))
	loaded, err = store.GetAlert(projectID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.KeyID)
	assert.Nil(t, loaded.Params)
}

func TestAlerts_ScopedToProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectA, _ := newAlertTestProject(t, store)
	projectB, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	alert, err := store.CreateAlert(db.Alert{ProjectID: projectA, Name: "A", Type: "slack", URL: strPtr("https://hooks.example/a")})
	require.NoError(t, err)

	_, err = store.GetAlert(projectB.ID, alert.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	err = store.UpdateTemplateAlerts(projectB.ID, 1, []int{alert.ID})
	assert.ErrorContains(t, err, "does not belong")
}

func TestAlerts_DefaultIDsOnlyEnabledDefaults(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, _ := newAlertTestProject(t, store)

	def, err := store.CreateAlert(db.Alert{ProjectID: projectID, Name: "default", Type: "slack", URL: strPtr("https://hooks.example/a"), Enabled: true, IsDefault: true})
	require.NoError(t, err)
	_, err = store.CreateAlert(db.Alert{ProjectID: projectID, Name: "disabled default", Type: "slack", URL: strPtr("https://hooks.example/b"), Enabled: false, IsDefault: true})
	require.NoError(t, err)
	_, err = store.CreateAlert(db.Alert{ProjectID: projectID, Name: "not default", Type: "slack", URL: strPtr("https://hooks.example/c"), Enabled: true})
	require.NoError(t, err)

	ids, err := store.GetDefaultAlertIDs(projectID)
	require.NoError(t, err)
	assert.Equal(t, []int{def.ID}, ids)
}

func TestAlerts_TemplateBindings(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repoID := newAlertTestProject(t, store)

	alert, err := store.CreateAlert(db.Alert{ProjectID: projectID, Name: "Ops", Type: "slack", URL: strPtr("https://hooks.example/a"), Enabled: true})
	require.NoError(t, err)

	// Default mode: alert_ids are ignored and nothing is linked.
	tpl, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repoID,
		Name:         "deploy",
		Playbook:     "site.yml",
		AlertIDs:     []int{alert.ID},
		AlertMode:    db.AlertModeDefault,
	})
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeDefault, tpl.AlertMode)
	assert.Empty(t, tpl.AlertIDs)

	// Switch to ids mode with an explicit list.
	tpl.AlertMode = db.AlertModeIDs
	tpl.AlertIDs = []int{alert.ID, alert.ID}
	require.NoError(t, store.UpdateTemplate(tpl))

	loaded, err := store.GetTemplate(projectID, tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeIDs, loaded.AlertMode)
	assert.Equal(t, []int{alert.ID}, loaded.AlertIDs)

	refs, err := store.GetAlertRefs(projectID, alert.ID)
	require.NoError(t, err)
	require.Len(t, refs.Templates, 1)
	assert.Equal(t, tpl.ID, refs.Templates[0].ID)
	assert.ErrorContains(t, store.DeleteAlert(projectID, alert.ID), "used by")

	// A legacy client that omits alert_mode and alert_ids keeps the binding.
	loaded.AlertMode = ""
	loaded.AlertIDs = nil
	require.NoError(t, store.UpdateTemplate(loaded))
	loaded, err = store.GetTemplate(projectID, tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeIDs, loaded.AlertMode)
	assert.Equal(t, []int{alert.ID}, loaded.AlertIDs)

	// Listing exposes the same data.
	listed, err := store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, []int{alert.ID}, listed[0].AlertIDs)

	// Back to default mode clears the binding, so the alert can be deleted.
	loaded.AlertMode = db.AlertModeDefault
	require.NoError(t, store.UpdateTemplate(loaded))
	refs, err = store.GetAlertRefs(projectID, alert.ID)
	require.NoError(t, err)
	assert.Empty(t, refs.Templates)
	require.NoError(t, store.DeleteAlert(projectID, alert.ID))
}

func TestAlerts_ScheduleBindings(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repoID := newAlertTestProject(t, store)

	alert, err := store.CreateAlert(db.Alert{ProjectID: projectID, Name: "Ops", Type: "slack", URL: strPtr("https://hooks.example/a"), Enabled: true})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repoID, Name: "deploy", Playbook: "site.yml"})
	require.NoError(t, err)

	schedule, err := store.CreateSchedule(db.Schedule{
		ProjectID:  projectID,
		TemplateID: tpl.ID,
		Name:       "nightly",
		CronFormat: "0 0 * * *",
		AlertMode:  db.AlertModeIDs,
		AlertIDs:   []int{alert.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, []int{alert.ID}, schedule.AlertIDs)

	loaded, err := store.GetSchedule(projectID, schedule.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeIDs, loaded.AlertMode)
	assert.Equal(t, []int{alert.ID}, loaded.AlertIDs)

	refs, err := store.GetAlertRefs(projectID, alert.ID)
	require.NoError(t, err)
	require.Len(t, refs.Schedules, 1)
	assert.Equal(t, schedule.ID, refs.Schedules[0].ID)

	// inherit mode drops the binding.
	loaded.AlertMode = db.AlertModeInherit
	require.NoError(t, store.UpdateSchedule(loaded))
	loaded, err = store.GetSchedule(projectID, schedule.ID)
	require.NoError(t, err)
	assert.Equal(t, db.AlertModeInherit, loaded.AlertMode)
	assert.Empty(t, loaded.AlertIDs)

	withTpl, err := store.GetProjectSchedules(projectID, false, true)
	require.NoError(t, err)
	require.Len(t, withTpl, 1)
	assert.Equal(t, db.AlertModeInherit, withTpl[0].AlertMode)
	assert.NotNil(t, withTpl[0].AlertIDs)
}

func TestAlerts_ClaimAlertSendIsIdempotent(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repoID := newAlertTestProject(t, store)

	tpl, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repoID, Name: "deploy", Playbook: "site.yml"})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: projectID, TemplateID: tpl.ID}, 0)
	require.NoError(t, err)

	claimed, err := store.ClaimAlertSend(task.ID, "instance:slack", db.AlertEventError)
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = store.ClaimAlertSend(task.ID, "instance:slack", db.AlertEventError)
	require.NoError(t, err)
	assert.False(t, claimed)

	claimed, err = store.ClaimAlertSend(task.ID, "instance:slack", db.AlertEventSuccess)
	require.NoError(t, err)
	assert.True(t, claimed)
}

func TestTasks_AlertSnapshotRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repoID := newAlertTestProject(t, store)

	tpl, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repoID, Name: "deploy", Playbook: "site.yml"})
	require.NoError(t, err)

	snap := db.AlertSnapshot{Instance: true, AlertIDs: []int{7}, OnSuccess: true, OnError: true}
	task, err := store.CreateTask(db.Task{ProjectID: projectID, TemplateID: tpl.ID, AlertSnapshot: &snap}, 0)
	require.NoError(t, err)

	loaded, err := store.GetTask(projectID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.AlertSnapshot)
	assert.Equal(t, snap, *loaded.AlertSnapshot)

	noSnap, err := store.CreateTask(db.Task{ProjectID: projectID, TemplateID: tpl.ID}, 0)
	require.NoError(t, err)
	loaded, err = store.GetTask(projectID, noSnap.ID)
	require.NoError(t, err)
	assert.Nil(t, loaded.AlertSnapshot)
}
