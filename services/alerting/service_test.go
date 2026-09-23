package alerting

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Notify_LegacyServerChannels(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Name: "Homelab", Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/chat", Trusted: true}
	mail := newFakeChannel("mail", db.AlertEvents{db.AlertEventError})
	mail.instance = &Destination{Trusted: true}

	svc, _ := newTestService(store, testConfig(), chat, mail)
	tpl := db.Template{ID: 1, ProjectID: 1, Name: "deploy"}

	require.NoError(t, svc.Notify(context.Background(), statusTask(10, 1), tpl, task_logger.TaskSuccessStatus, nil))
	require.NoError(t, svc.Notify(context.Background(), statusTask(11, 1), tpl, task_logger.TaskFailStatus, nil))
	require.NoError(t, svc.Notify(context.Background(), statusTask(12, 1), tpl, task_logger.TaskWaitingConfirmation, nil))
	require.NoError(t, svc.Notify(context.Background(), statusTask(13, 1), tpl, task_logger.TaskRunningStatus, nil))

	assert.Len(t, chat.sends(), 3, "chat channels report success, error and waiting confirmation")
	assert.Len(t, mail.sends(), 1, "e-mail reports failures only")
	assert.Equal(t, `Task "deploy" failed`, mail.sends()[0].Msg.Subject)
	assert.True(t, chat.sends()[0].Dest.Trusted)
}

func TestService_Notify_ProjectWithoutAlertFlagSendsNothingFromConfig(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: false})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/chat", Trusted: true}

	svc, _ := newTestService(store, testConfig(), chat)
	require.NoError(t, svc.Notify(context.Background(), statusTask(1, 1), db.Template{ID: 1, ProjectID: 1}, task_logger.TaskFailStatus, nil))

	assert.Empty(t, chat.sends())
}

func TestService_Notify_TemplateSuppressFlags(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/chat", Trusted: true}

	svc, _ := newTestService(store, testConfig(), chat)
	tpl := db.Template{ID: 1, ProjectID: 1, SuppressSuccessAlerts: true}

	require.NoError(t, svc.Notify(context.Background(), statusTask(1, 1), tpl, task_logger.TaskSuccessStatus, nil))
	require.NoError(t, svc.Notify(context.Background(), statusTask(2, 1), tpl, task_logger.TaskFailStatus, nil))
	require.NoError(t, svc.Notify(context.Background(), statusTask(3, 1), tpl, task_logger.TaskWaitingConfirmation, nil))

	sends := chat.sends()
	require.Len(t, sends, 2)
	assert.Equal(t, task_logger.TaskFailStatus, sends[0].Msg.Payload.Task.Status)
	assert.Equal(t, task_logger.TaskWaitingConfirmation, sends[1].Msg.Payload.Task.Status)
}

func TestService_Notify_ProjectAlerts(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: false})
	chat := newFakeChannel("chat", chatEvents())
	svc, _ := newTestService(store, testConfig(), chat)

	url := "https://hooks.example/a"
	body := `{{ .Project.ID }}:{{ .Task.ID }}`
	defaultAlert := store.addAlert(db.Alert{Name: "default", Type: "chat", Enabled: true, IsDefault: true, URL: &url, Body: &body})
	errorsOnly := store.addAlert(db.Alert{Name: "errors", Type: "chat", Enabled: true, URL: &url, Events: db.AlertEvents{db.AlertEventError}})
	disabled := store.addAlert(db.Alert{Name: "off", Type: "chat", Enabled: false, URL: &url})

	logger := &recordingLogger{}
	tpl := db.Template{ID: 1, ProjectID: 1, AlertMode: db.AlertModeIDs, AlertIDs: []int{defaultAlert.ID, errorsOnly.ID, disabled.ID}}

	task := statusTask(1, 1)
	require.NoError(t, svc.Snapshot(&task, tpl))
	require.NotNil(t, task.AlertSnapshot)
	assert.Equal(t, []int{defaultAlert.ID, errorsOnly.ID, disabled.ID}, task.AlertSnapshot.AlertIDs)
	assert.False(t, task.AlertSnapshot.Instance)
	require.Len(t, task.AlertSnapshot.Deliveries, 2)

	require.NoError(t, svc.Notify(context.Background(), task, tpl, task_logger.TaskSuccessStatus, logger))
	sends := chat.sends()
	require.Len(t, sends, 1, "only the default alert listens to success")
	assert.Equal(t, "1:1", sends[0].Msg.Body, "custom body is rendered")
	assert.False(t, sends[0].Dest.Trusted)

	require.NoError(t, svc.Notify(context.Background(), task, tpl, task_logger.TaskFailStatus, logger))
	assert.Len(t, chat.sends(), 3)
	assert.NotContains(t, logger.lines, "Alert %d skipped: %s")
}

func TestService_Notify_DefaultModeMergesServerChannelsAndDefaults(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/instance", Trusted: true}
	svc, _ := newTestService(store, testConfig(), chat)

	url := "https://hooks.example/project"
	store.addAlert(db.Alert{Name: "default", Type: "chat", Enabled: true, IsDefault: true, URL: &url})

	task := statusTask(1, 1)
	tpl := db.Template{ID: 1, ProjectID: 1}
	require.NoError(t, svc.Snapshot(&task, tpl))
	assert.True(t, task.AlertSnapshot.Instance)
	assert.Len(t, task.AlertSnapshot.AlertIDs, 1)

	require.NoError(t, svc.Notify(context.Background(), task, tpl, task_logger.TaskSuccessStatus, nil))
	sends := chat.sends()
	require.Len(t, sends, 2)
	assert.Equal(t, "https://hooks.example/instance", sends[0].Dest.URL)
	assert.Equal(t, "https://hooks.example/project", sends[1].Dest.URL)
}

func TestService_Notify_SnapshotIsFrozen(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/instance", Trusted: true}
	svc, _ := newTestService(store, testConfig(), chat)

	task := statusTask(1, 1)
	task.AlertSnapshot = &db.AlertSnapshot{Instance: false, AlertIDs: []int{}, OnSuccess: true, OnError: true}

	require.NoError(t, svc.Notify(context.Background(), task, db.Template{ID: 1, ProjectID: 1}, task_logger.TaskFailStatus, nil))
	assert.Empty(t, chat.sends(), "a task snapshotted as silent stays silent even though the project allows alerts now")
}

func TestService_Notify_UsesSnapshottedAlertState(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: false})
	chat := newFakeChannel("chat", chatEvents())
	svc, _ := newTestService(store, testConfig(), chat)

	url := "https://hooks.example/original"
	body := `{"text":"{{ .Name }}"}`
	alert := store.addAlert(db.Alert{Name: "Ops", Type: "chat", Enabled: true, URL: &url, Body: &body})

	task := statusTask(1, 1)
	tpl := db.Template{ID: 1, ProjectID: 1, Name: "deploy", AlertMode: db.AlertModeIDs, AlertIDs: []int{alert.ID}}
	require.NoError(t, svc.Snapshot(&task, tpl))

	changedURL := "https://hooks.example/changed"
	changedBody := `{"text":"changed"}`
	alert.URL = &changedURL
	alert.Body = &changedBody
	alert.Enabled = false
	require.NoError(t, store.UpdateAlert(alert))

	require.NoError(t, svc.Notify(context.Background(), task, tpl, task_logger.TaskFailStatus, nil))
	require.Len(t, chat.sends(), 1)
	assert.Equal(t, "https://hooks.example/original", chat.sends()[0].Dest.URL)
	assert.Equal(t, "{\"text\":\"deploy\"}", chat.sends()[0].Msg.Body)
}

func TestService_Notify_ClaimPreventsDoubleSend(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/instance", Trusted: true}
	svc, _ := newTestService(store, testConfig(), chat)

	task := statusTask(7, 1)
	tpl := db.Template{ID: 1, ProjectID: 1}
	require.NoError(t, svc.Notify(context.Background(), task, tpl, task_logger.TaskFailStatus, nil))
	require.NoError(t, svc.Notify(context.Background(), task, tpl, task_logger.TaskFailStatus, nil))

	assert.Len(t, chat.sends(), 1)
}

func TestService_Notify_SendErrorsAreReportedAndLogged(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/instance", Trusted: true}
	chat.sendErr = errors.New("boom")
	svc, _ := newTestService(store, testConfig(), chat)

	logger := &recordingLogger{}
	err := svc.Notify(context.Background(), statusTask(1, 1), db.Template{ID: 1, ProjectID: 1}, task_logger.TaskFailStatus, logger)
	assert.ErrorContains(t, err, "boom")
	assert.Contains(t, logger.lines, "Can't send alert %s: %s")
}

func TestService_Notify_FailedSendCanRetry(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/instance", Trusted: true}
	chat.sendErr = errors.New("boom")
	svc, _ := newTestService(store, testConfig(), chat)

	task := statusTask(1, 1)
	err := svc.Notify(context.Background(), task, db.Template{ID: 1, ProjectID: 1}, task_logger.TaskFailStatus, nil)
	require.ErrorContains(t, err, "boom")

	chat.sendErr = nil
	require.NoError(t, svc.Notify(context.Background(), task, db.Template{ID: 1, ProjectID: 1}, task_logger.TaskFailStatus, nil))
	assert.Len(t, chat.sends(), 2)
}

func TestService_EmailFallsBackToOptedInMembers(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	store.members = []db.UserWithProjectRole{
		{User: db.User{ID: 1, Email: "yes@example.com", Alert: true}},
		{User: db.User{ID: 2, Email: "no@example.com", Alert: false}},
	}
	store.admins = []db.User{
		{ID: 1, Email: "yes@example.com", Alert: true},
		{ID: 3, Email: "admin@example.com", Alert: true},
	}

	mail := newFakeChannel("mail", db.AlertEvents{db.AlertEventError})
	mail.fields = []Field{{Name: FieldRecipients}}
	mail.instance = &Destination{Trusted: true}
	svc, _ := newTestService(store, testConfig(), mail)

	require.NoError(t, svc.Notify(context.Background(), statusTask(1, 1), db.Template{ID: 1, ProjectID: 1}, task_logger.TaskFailStatus, nil))
	require.Len(t, mail.sends(), 1)
	assert.Equal(t, []string{"yes@example.com", "admin@example.com"}, mail.sends()[0].Dest.Recipients)

	explicit := "ops@example.com"
	alert := store.addAlert(db.Alert{Name: "ops", Type: "mail", Enabled: true, Recipients: &explicit})
	require.NoError(t, svc.SendTest(context.Background(), store.project, alert))
	assert.Equal(t, []string{"ops@example.com"}, mail.sends()[1].Dest.Recipients)
}

func TestService_ValidateAlert(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1})
	svc, _ := newTestService(store, testConfig(), NewRegistry().List()...)

	url := "https://hooks.slack.com/x"
	chat := "123"
	keyID := 42
	alert := db.Alert{Name: "Ops", Type: "SLACK", URL: &url, ChatID: &chat, KeyID: &keyID, Params: db.MapStringAnyField{"smtp_host": "x"}}
	require.NoError(t, svc.ValidateAlert(&alert))
	assert.Equal(t, TypeSlack, alert.Type)
	assert.Nil(t, alert.ChatID, "fields the channel does not use are dropped")
	assert.Nil(t, alert.KeyID, "channels without a secret drop the key")
	assert.Nil(t, alert.Params, "params the channel does not declare are dropped")

	bad := db.Alert{Name: "Ops", Type: "slack"}
	assert.ErrorContains(t, svc.ValidateAlert(&bad), "URL")

	unknown := db.Alert{Name: "Ops", Type: "pager"}
	assert.ErrorContains(t, svc.ValidateAlert(&unknown), "unknown alert type")

	broken := "{{ .Name "
	withBody := db.Alert{Name: "Ops", Type: "slack", URL: &url, Body: &broken}
	assert.ErrorContains(t, svc.ValidateAlert(&withBody), "invalid message template")

	// Telegram without a server token needs an own key of the right type.
	noToken := db.Alert{ProjectID: 1, Name: "TG", Type: "telegram", ChatID: &chat}
	assert.ErrorContains(t, svc.ValidateAlert(&noToken), "telegram bot token")

	wrongType := store.addKey(db.AccessKeyLoginPassword, "p", "u")
	withWrongKey := db.Alert{ProjectID: 1, Name: "TG", Type: "telegram", ChatID: &chat, KeyID: &wrongType.ID}
	assert.ErrorContains(t, svc.ValidateAlert(&withWrongKey), "must be of type string")

	own := store.addKey(db.AccessKeyString, "bot-token", "")
	withKey := db.Alert{ProjectID: 1, Name: "TG", Type: "telegram", ChatID: &chat, KeyID: &own.ID}
	assert.NoError(t, svc.ValidateAlert(&withKey))

	missing := 999
	withMissingKey := db.Alert{ProjectID: 1, Name: "TG", Type: "telegram", ChatID: &chat, KeyID: &missing}
	assert.ErrorContains(t, svc.ValidateAlert(&withMissingKey), "does not belong")
}

func TestService_TelegramUsesOwnBotToken(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := newFakeStore(db.Project{ID: 1})
	own := store.addKey(db.AccessKeyString, "own-bot", "")
	chat := "5"
	alert := store.addAlert(db.Alert{Name: "TG", Type: "telegram", Enabled: true, ChatID: &chat, KeyID: &own.ID})

	registry := NewRegistry()
	registry.Register(&telegramChannel{channelBase: registry.byType[TypeTelegram].(*telegramChannel).channelBase, api: server.URL + "/bot"})
	svc := NewServiceWithConfig(store, registry, nil, func() *util.ConfigType { return &util.ConfigType{TelegramToken: "server-bot"} })

	require.NoError(t, svc.SendTest(context.Background(), store.project, alert))
	assert.Equal(t, "/botown-bot/sendMessage", path)
}

func TestService_SendProjectTest(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Alert: true})
	chat := newFakeChannel("chat", chatEvents())
	chat.instance = &Destination{URL: "https://hooks.example/instance", Trusted: true}
	svc, _ := newTestService(store, testConfig(), chat)

	url := "https://hooks.example/project"
	store.addAlert(db.Alert{Name: "on", Type: "chat", Enabled: true, URL: &url, Events: db.AlertEvents{db.AlertEventError}})
	store.addAlert(db.Alert{Name: "off", Type: "chat", Enabled: false, URL: &url})

	sent, err := svc.SendProjectTest(context.Background(), store.project)
	require.NoError(t, err)
	assert.Equal(t, 2, sent, "server channel plus the enabled alert, regardless of its events")
	assert.Len(t, chat.sends(), 2)
	assert.Contains(t, chat.sends()[0].Msg.Body, "Test Notification")

	quiet := newFakeStore(db.Project{ID: 1, Alert: false})
	svc, _ = newTestService(quiet, testConfig(), chat)
	sent, err = svc.SendProjectTest(context.Background(), quiet.project)
	require.NoError(t, err)
	assert.Equal(t, 0, sent)
}

func TestService_RealSlackChannelEndToEnd(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		received = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &util.ConfigType{SlackAlert: true, SlackUrl: server.URL, WebHost: "https://semaphore.example"}
	store := newFakeStore(db.Project{ID: 1, Name: "Homelab", Alert: true})
	store.users[5] = db.User{ID: 5, Name: "alice"}
	svc, _ := newTestService(store, cfg, NewRegistry().List()...)

	userID := 5
	task := db.Task{ID: 3, ProjectID: 1, TemplateID: 2, UserID: &userID}
	require.NoError(t, svc.Notify(context.Background(), task, db.Template{ID: 2, ProjectID: 1, Name: "deploy"}, task_logger.TaskFailStatus, nil))

	assert.Contains(t, received, `"title": "Task: deploy"`)
	assert.Contains(t, received, `"color": "danger"`)
	assert.Contains(t, received, `"value": "alice"`)
	assert.Contains(t, received, "https://semaphore.example/project/1/templates/2?t=3")
}

func TestBuildPayload(t *testing.T) {
	store := newFakeStore(db.Project{ID: 1, Name: "Homelab"})
	store.users[5] = db.User{ID: 5, Name: "alice"}
	store.schedules[9] = db.Schedule{ID: 9, Name: "nightly"}
	buildVersion := "1.0.7"
	store.tasks[20] = db.Task{ID: 20, ProjectID: 1, TemplateID: 30, Version: &buildVersion}
	store.templates[30] = db.Template{ID: 30, Type: db.TemplateBuild}

	userID, scheduleID, buildTaskID := 5, 9, 20
	task := db.Task{ID: 3, ProjectID: 1, TemplateID: 2, UserID: &userID, ScheduleID: &scheduleID, BuildTaskID: &buildTaskID, Message: "go"}
	tpl := db.Template{ID: 2, ProjectID: 1, Name: "deploy", Type: db.TemplateDeploy, Playbook: "site.yml"}

	payload := BuildPayload(testConfig(), store, store.project, tpl, task, task_logger.TaskSuccessStatus)

	assert.Equal(t, "alice", payload.Author)
	assert.Equal(t, "build 1.0.7", payload.Task.Version)
	assert.Equal(t, "nightly", payload.ScheduleName)
	assert.Equal(t, "schedule", payload.Task.Trigger)
	assert.Equal(t, "Homelab", payload.Project.Name)
	assert.Equal(t, "site.yml", payload.Playbook)
	assert.Equal(t, "go", payload.Task.Desc)
	assert.Equal(t, "https://semaphore.example/project/1/templates/2?t=3", payload.Task.URL)

	anonymous := BuildPayload(nil, nil, db.Project{}, db.Template{}, db.Task{}, task_logger.TaskFailStatus)
	assert.Equal(t, "—", anonymous.Author)
	assert.Equal(t, "api", anonymous.Task.Trigger)
}
