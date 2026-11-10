package tasks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
			expected:              false,
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

// alertScheduleStoreStub serves only the schedule lookup; any other store call
// panics on the nil embedded Store.
type alertScheduleStoreStub struct {
	db.Store
	schedule db.Schedule
}

func (s *alertScheduleStoreStub) GetSchedule(projectID int, scheduleID int) (db.Schedule, error) {
	if s.schedule.ProjectID != projectID || s.schedule.ID != scheduleID {
		return db.Schedule{}, db.ErrNotFound
	}
	return s.schedule, nil
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// captureTelegramAlertText runs sendTelegramAlert with the Telegram API call
// intercepted and returns the "text" field of the message it would send.
func captureTelegramAlertText(t *testing.T, tr *TaskRunner) string {
	t.Helper()

	return captureTelegramAlertTextWithWebHost(t, tr, "http://semaphore.example.com:3001")
}

func captureTelegramAlertTextWithWebHost(t *testing.T, tr *TaskRunner, webHost string) string {
	t.Helper()

	prevConfig := util.Config
	prevTransport := http.DefaultTransport
	defer func() {
		util.Config = prevConfig
		http.DefaultTransport = prevTransport
	}()

	util.Config = &util.ConfigType{
		WebHost:       webHost,
		TelegramAlert: true,
		TelegramChat:  "100",
		TelegramToken: "token",
	}

	var payload []byte
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var err error
		payload, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	tr.sendTelegramAlert()

	require.NotEmpty(t, payload, "telegram alert was not sent")

	var msg struct {
		ChatID    string `json:"chat_id"`
		ParseMode string `json:"parse_mode"`
		Text      string `json:"text"`
	}
	require.NoError(t, json.Unmarshal(payload, &msg))
	require.Equal(t, "HTML", msg.ParseMode)

	return msg.Text
}

// newScheduledAlertTaskRunner builds the task from the issue: started by a
// schedule, so it has no user, no version and no message.
func newScheduledAlertTaskRunner() *TaskRunner {
	scheduleID := 7

	return &TaskRunner{
		Task: db.Task{
			ID:         310,
			ProjectID:  1,
			TemplateID: 21,
			Status:     task_logger.TaskSuccessStatus,
			ScheduleID: &scheduleID,
		},
		Template: db.Template{
			ID:        21,
			ProjectID: 1,
			Name:      "Funding stop",
			Type:      db.TemplateTask,
		},
		alert: true,
		pool: &TaskPool{
			logger: make(chan logRecord, 100),
			store: &alertScheduleStoreStub{
				schedule: db.Schedule{
					ID:         scheduleID,
					ProjectID:  1,
					TemplateID: 21,
					Name:       "Funding stop",
				},
			},
		},
	}
}

// Reproduces https://github.com/semaphoreui/semaphore/issues/4377
func TestSendTelegramAlert_ScheduledTask(t *testing.T) {
	const taskURL = "http://semaphore.example.com:3001/project/1/templates/21?t=310"

	text := captureTelegramAlertText(t, newScheduledAlertTaskRunner())
	lines := strings.Split(text, "\n")

	t.Run("task link is an anchor on the task id", func(t *testing.T) {
		assert.Contains(t, text, `<a href="`+taskURL+`">#310</a>`)
		assert.NotContains(t, lines, taskURL, "bare URL must not take its own line")
	})

	t.Run("author is the schedule instead of a dash", func(t *testing.T) {
		assert.NotContains(t, text, "by —")
		assert.Contains(t, lines, "by schedule Funding stop")
	})

	t.Run("no dangling separator without version and message", func(t *testing.T) {
		assert.NotContains(t, text, "<code></code>")
		for _, line := range lines {
			assert.False(t, strings.HasSuffix(line, " - "), "dangling separator in %q", line)
		}
	})

	t.Run("whole message", func(t *testing.T) {
		assert.Equal(t,
			"<code>Funding stop</code>\n"+
				`<a href="`+taskURL+`">#310</a> <b>✅ SUCCESS</b>`+"\n"+
				"by schedule Funding stop",
			text,
		)
	})
}

func TestSendTelegramAlert_VersionAndMessage(t *testing.T) {
	tr := newScheduledAlertTaskRunner()
	version := "1.2.3"
	tr.Task.Version = &version
	tr.Task.Message = "release"

	text := captureTelegramAlertText(t, tr)

	assert.Contains(t, text, "<b>✅ SUCCESS</b> <code>1.2.3</code> - release\n")
}

func TestAlertInfos_Author(t *testing.T) {
	missingScheduleID := 8

	tests := []struct {
		name       string
		scheduleID *int
		expected   string
	}{
		{"no user and no schedule", nil, "—"},
		{"deleted schedule", &missingScheduleID, "schedule"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := newScheduledAlertTaskRunner()
			tr.Task.ScheduleID = tt.scheduleID

			author, _ := tr.alertInfos()

			assert.Equal(t, tt.expected, author)
		})
	}

	t.Run("schedule without a name", func(t *testing.T) {
		tr := newScheduledAlertTaskRunner()
		tr.pool.store.(*alertScheduleStoreStub).schedule.Name = ""

		author, _ := tr.alertInfos()

		assert.Equal(t, "schedule", author)
	})
}

// Telegram rejects the whole message when href is not an absolute URL.
func TestSendTelegramAlert_NoAbsoluteTaskURL(t *testing.T) {
	tests := []struct {
		name    string
		webHost string
	}{
		{"web host unset", ""},
		{"web host is a path", "/semaphore"},
		{"web host without scheme", "semaphore.example.com:3001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := captureTelegramAlertTextWithWebHost(t, newScheduledAlertTaskRunner(), tt.webHost)

			assert.Equal(t,
				"<code>Funding stop</code>\n"+
					"#310 <b>✅ SUCCESS</b>\n"+
					"by schedule Funding stop",
				text,
			)
		})
	}
}
