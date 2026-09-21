package tasks

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
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

func TestParseTelegramChat(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		chatID   string
		threadID string
	}{
		{"empty", "", "", ""},
		{"chat only", "-1001234567890", "-1001234567890", ""},
		{"chat and thread", "-1001234567890:42", "-1001234567890", "42"},
		{"whitespace", "  -1001234567890:7  ", "-1001234567890", "7"},
		{"empty thread", "-1001234567890:", "-1001234567890", ""},
		{"non-numeric thread", "-1001234567890:abc", "-1001234567890", ""},
		{"username", "@mychannel", "@mychannel", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatID, threadID := parseTelegramChat(tt.raw)
			assert.Equal(t, tt.chatID, chatID)
			assert.Equal(t, tt.threadID, threadID)
		})
	}
}

func TestTelegramTemplate_MessageThreadID(t *testing.T) {
	tpl, err := template.ParseFS(templates, "templates/telegram.tmpl")
	require.NoError(t, err)

	t.Run("with thread", func(t *testing.T) {
		var body bytes.Buffer
		err := tpl.Execute(&body, Alert{
			Name:   "demo",
			Author: "alice",
			Chat: alertChat{
				ID:              "-1001234567890",
				MessageThreadID: "42",
			},
			Task: alertTask{
				ID:     "1",
				URL:    "https://example.com",
				Result: "success",
			},
		})
		require.NoError(t, err)
		assert.Contains(t, body.String(), `"chat_id": "-1001234567890"`)
		assert.Contains(t, body.String(), `"message_thread_id": 42`)
	})

	t.Run("without thread", func(t *testing.T) {
		var body bytes.Buffer
		err := tpl.Execute(&body, Alert{
			Name:   "demo",
			Author: "alice",
			Chat: alertChat{
				ID: "-1001234567890",
			},
			Task: alertTask{
				ID:     "1",
				URL:    "https://example.com",
				Result: "success",
			},
		})
		require.NoError(t, err)
		assert.Contains(t, body.String(), `"chat_id": "-1001234567890"`)
		assert.NotContains(t, body.String(), "message_thread_id")
	})
}
