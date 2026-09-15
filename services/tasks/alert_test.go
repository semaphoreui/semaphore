package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSkipStatusAlert(t *testing.T) {
	tests := []struct {
		name                  string
		status                task_logger.TaskStatus
		suppressSuccessAlerts bool
		suppressErrorAlerts   bool
		expected              bool
	}{
		{
			name:                  "success with suppress success",
			status:                task_logger.TaskSuccessStatus,
			suppressSuccessAlerts: true,
			expected:              true,
		},
		{
			name:                "fail with suppress error",
			status:              task_logger.TaskFailStatus,
			suppressErrorAlerts: true,
			expected:            true,
		},
		{
			name:     "fail without flags",
			status:   task_logger.TaskFailStatus,
			expected: false,
		},
		{
			name:     "success without flags",
			status:   task_logger.TaskSuccessStatus,
			expected: false,
		},
		{
			name:                  "fail with only success flag",
			status:                task_logger.TaskFailStatus,
			suppressSuccessAlerts: true,
			expected:              false,
		},
		{
			name:                "success with only error flag",
			status:              task_logger.TaskSuccessStatus,
			suppressErrorAlerts: true,
			expected:            false,
		},
		{
			name:                  "fail with both flags",
			status:                task_logger.TaskFailStatus,
			suppressSuccessAlerts: true,
			suppressErrorAlerts:   true,
			expected:              true,
		},
		{
			name:                  "success with both flags",
			status:                task_logger.TaskSuccessStatus,
			suppressSuccessAlerts: true,
			suppressErrorAlerts:   true,
			expected:              true,
		},
		{
			name:                  "waiting confirmation with both flags",
			status:                task_logger.TaskWaitingConfirmation,
			suppressSuccessAlerts: true,
			suppressErrorAlerts:   true,
			expected:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &TaskRunner{
				Task: db.Task{
					Status: tt.status,
				},
				Template: db.Template{
					SuppressSuccessAlerts: tt.suppressSuccessAlerts,
					SuppressErrorAlerts:   tt.suppressErrorAlerts,
				},
			}

			assert.Equal(t, tt.expected, runner.shouldSkipStatusAlert())
		})
	}
}

func TestProjectAlertReady(t *testing.T) {
	orig := util.Config
	t.Cleanup(func() { util.Config = orig })
	util.Config = &util.ConfigType{}

	hook := "https://hooks.example.com/x"
	token := "gotify-token"

	assert.NoError(t, projectAlertReady(db.Alert{
		Name: "slack",
		Type: db.AlertTypeSlack,
		URL:  &hook,
	}))

	assert.Error(t, projectAlertReady(db.Alert{Name: "slack", Type: db.AlertTypeSlack}))

	util.Config.SlackAlert = true
	util.Config.SlackUrl = hook
	assert.Error(t, projectAlertReady(db.Alert{Name: "slack", Type: db.AlertTypeSlack}))
	assert.NoError(t, projectAlertReady(db.Alert{Name: "slack", Type: db.AlertTypeSlack, URL: &hook}))

	chatID := "12345"
	assert.Error(t, projectAlertReady(db.Alert{Name: "tg", Type: db.AlertTypeTelegram}))
	util.Config.TelegramAlert = true
	util.Config.TelegramToken = "bot"
	assert.Error(t, projectAlertReady(db.Alert{Name: "tg", Type: db.AlertTypeTelegram}))
	assert.NoError(t, projectAlertReady(db.Alert{Name: "tg", Type: db.AlertTypeTelegram, ChatID: &chatID}))

	assert.Error(t, projectAlertReady(db.Alert{Name: "mail", Type: db.AlertTypeEmail}))
	util.Config.EmailAlert = true
	util.Config.EmailHost = "smtp.example.com"
	assert.NoError(t, projectAlertReady(db.Alert{Name: "mail", Type: db.AlertTypeEmail}))

	assert.NoError(t, projectAlertReady(db.Alert{
		Name:  "gotify",
		Type:  db.AlertTypeGotify,
		URL:   &hook,
		Token: &token,
	}))
	assert.Error(t, projectAlertReady(db.Alert{Name: "gotify", Type: db.AlertTypeGotify}))
	assert.Error(t, projectAlertReady(db.Alert{
		Name: "gotify-url-only",
		Type: db.AlertTypeGotify,
		URL:  &hook,
	}))

	httpHook := "http://gotify.example.com"
	assert.NoError(t, projectAlertReady(db.Alert{
		Name:  "gotify-http",
		Type:  db.AlertTypeGotify,
		URL:   &httpHook,
		Token: &token,
	}))

	util.Config.GotifyAlert = true
	util.Config.GotifyUrl = hook
	util.Config.GotifyToken = token
	assert.NoError(t, projectAlertReady(db.Alert{Name: "gotify-instance", Type: db.AlertTypeGotify}))
	assert.NoError(t, projectAlertReady(db.Alert{
		Name:  "gotify-custom-token",
		Type:  db.AlertTypeGotify,
		Token: &token,
	}))

	otherToken := "other-token"
	assert.Error(t, projectAlertReady(db.Alert{
		Name:  "gotify-custom-url-instance-token",
		Type:  db.AlertTypeGotify,
		URL:   &hook,
		Token: nil,
	}))
	assert.NoError(t, projectAlertReady(db.Alert{
		Name:  "gotify-custom-pair",
		Type:  db.AlertTypeGotify,
		URL:   &hook,
		Token: &otherToken,
	}))
}

func TestSendStatusAlerts_NilSnapshotDoesNothing(t *testing.T) {
	runner := &TaskRunner{
		Task: db.Task{Status: task_logger.TaskSuccessStatus},
	}
	assert.NotPanics(t, func() {
		runner.sendStatusAlerts()
	})
}

func TestNewAlertPayload_ScheduleName(t *testing.T) {
	runner := &TaskRunner{
		Task: db.Task{
			Status:  task_logger.TaskSuccessStatus,
			Message: "done",
		},
		Template: db.Template{Name: "Nightly", Playbook: "site.yml"},
	}
	payload := runner.newAlertPayload("telegram", "1", "")
	assert.Empty(t, payload.ScheduleName)

	id := 9
	runner.Task.ScheduleID = &id
	payload = runner.newAlertPayload("telegram", "1", "")
	assert.Empty(t, payload.ScheduleName)
}

func TestDefaultAlertBodies(t *testing.T) {
	bodies, err := DefaultAlertBodies()
	require.NoError(t, err)
	assert.Contains(t, bodies[db.AlertTypeTelegram], "{{ .Task.ID }}")
	assert.NotContains(t, bodies[db.AlertTypeTelegram], "chat_id")
	assert.NotContains(t, bodies[db.AlertTypeTelegram], "message_thread_id")
	assert.Contains(t, bodies[db.AlertTypeGotify], `"title"`)
	assert.Contains(t, bodies[db.AlertTypeEmail], "Task Log")
	assert.NotEqual(t, bodies[db.AlertTypeTelegram], bodies[db.AlertTypeGotify])
}

func TestWrapTelegramMessage(t *testing.T) {
	body, err := wrapTelegramMessage("123", "9", "hello <b>world</b>")
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"chat_id": "123",
		"parse_mode": "HTML",
		"text": "hello <b>world</b>",
		"message_thread_id": 9
	}`, body)

	noThread, err := wrapTelegramMessage("-1001", "", "ping")
	require.NoError(t, err)
	assert.NotContains(t, noThread, "message_thread_id")
	assert.Contains(t, noThread, `"chat_id":"-1001"`)

	_, err = wrapTelegramMessage("123", "not-a-number", "ping")
	require.Error(t, err)
}
