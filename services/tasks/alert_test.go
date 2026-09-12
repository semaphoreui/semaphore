package tasks

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

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

func TestResolveTelegramDestination(t *testing.T) {
	tests := []struct {
		name          string
		globalChat    string
		globalThread  string
		projectChat   *string
		projectThread *string
		wantChat      string
		wantThread    int64
		wantInvalid   bool
	}{
		{
			name:       "only global chat",
			globalChat: "-1001",
			wantChat:   "-1001",
		},
		{
			name:          "thread without chat still resolves",
			projectThread: strPtr("7"),
			wantThread:    7,
		},
		{
			name:         "global chat and global thread",
			globalChat:   "-1001",
			globalThread: "42",
			wantChat:     "-1001",
			wantThread:   42,
		},
		{
			name:         "project chat without thread does not use global thread",
			globalChat:   "-1001",
			globalThread: "42",
			projectChat:  strPtr("-2002"),
			wantChat:     "-2002",
		},
		{
			name:          "empty project chat with project thread uses global chat",
			globalChat:    "-1001",
			globalThread:  "42",
			projectChat:   strPtr(""),
			projectThread: strPtr("99"),
			wantChat:      "-1001",
			wantThread:    99,
		},
		{
			name:          "nil project chat with project thread uses global chat",
			globalChat:    "-1001",
			globalThread:  "42",
			projectThread: strPtr("99"),
			wantChat:      "-1001",
			wantThread:    99,
		},
		{
			name:          "project chat and project thread",
			globalChat:    "-1001",
			globalThread:  "42",
			projectChat:   strPtr("-2002"),
			projectThread: strPtr("7"),
			wantChat:      "-2002",
			wantThread:    7,
		},
		{
			name:         "invalid global thread is omitted",
			globalChat:   "-1001",
			globalThread: "not-a-number",
			wantChat:     "-1001",
			wantInvalid:  true,
		},
		{
			name:          "invalid project thread is omitted",
			globalChat:    "-1001",
			globalThread:  "42",
			projectThread: strPtr("abc"),
			wantChat:      "-1001",
			wantInvalid:   true,
		},
		{
			name:          "whitespace-only project chat is treated as unset",
			globalChat:    "-1001",
			globalThread:  "42",
			projectChat:   strPtr("   "),
			projectThread: strPtr(" 99 "),
			wantChat:      "-1001",
			wantThread:    99,
		},
		{
			name:         "leading zeros are canonicalized",
			globalChat:   "-1001",
			globalThread: "042",
			wantChat:     "-1001",
			wantThread:   42,
		},
		{
			name:         "zero thread is omitted",
			globalChat:   "-1001",
			globalThread: "0",
			wantChat:     "-1001",
			wantInvalid:  true,
		},
		{
			name:         "negative thread is omitted",
			globalChat:   "-1001",
			globalThread: "-5",
			wantChat:     "-1001",
			wantInvalid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := resolveTelegramDestination(
				tt.globalChat,
				tt.globalThread,
				tt.projectChat,
				tt.projectThread,
			)
			assert.Equal(t, tt.wantChat, dest.chatID)
			assert.Equal(t, tt.wantThread, dest.threadID)
			assert.Equal(t, tt.wantInvalid, dest.invalidThread)
		})
	}
}

func TestRenderTelegramAlert_Payload(t *testing.T) {
	base := Alert{
		Name:   "Deploy",
		Author: "Ann",
		Task: alertTask{
			ID:      "12",
			URL:     "https://example.test/task/12",
			Result:  "success",
			Desc:    "ok",
			Version: "1.0",
		},
	}

	t.Run("without thread id", func(t *testing.T) {
		body, err := renderTelegramAlert(base, telegramDestination{chatID: "-1001"})
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "-1001", payload["chat_id"])
		assert.Equal(t, "HTML", payload["parse_mode"])
		assert.Equal(
			t,
			"<code>Deploy</code>\n#12 <b>success</b> <code>1.0</code> - ok\nby Ann\nhttps://example.test/task/12",
			payload["text"],
		)
		_, hasThread := payload["message_thread_id"]
		assert.False(t, hasThread)
		assert.NotContains(t, string(body), "message_thread_id")
	})

	t.Run("with thread id as json number", func(t *testing.T) {
		body, err := renderTelegramAlert(base, telegramDestination{chatID: "-1001", threadID: 42})
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "-1001", payload["chat_id"])
		assert.Equal(t, float64(42), payload["message_thread_id"])
	})

	t.Run("escapes html in text", func(t *testing.T) {
		alert := base
		alert.Name = `Deploy <prod> & "qa"`

		body, err := renderTelegramAlert(alert, telegramDestination{chatID: "-1001"})
		require.NoError(t, err)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		text, ok := payload["text"].(string)
		require.True(t, ok)
		assert.Contains(t, text, `Deploy &lt;prod&gt; &amp; &#34;qa&#34;`)
		assert.NotContains(t, text, `<prod>`)
	})
}
