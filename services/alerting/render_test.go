package alerting

import (
	"encoding/json"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func samplePayload() Payload {
	return Payload{
		Name:   "Deploy <prod>",
		Author: "alice",
		Color:  "good",
		Task: PayloadTask{
			ID:      "42",
			URL:     "https://semaphore.example/project/1/templates/2?t=42",
			Result:  task_logger.TaskSuccessStatus.Format(),
			Desc:    "hotfix",
			Version: "1.2.3",
			Status:  task_logger.TaskSuccessStatus,
		},
		Chat:    PayloadChat{ID: "123"},
		Project: PayloadProject{ID: 1, Name: "Homelab"},
	}
}

func TestRender_BuiltinTemplatesProduceValidOutput(t *testing.T) {
	registry := NewRegistry()
	payload := samplePayload()

	for _, channel := range registry.List() {
		t.Run(string(channel.Type()), func(t *testing.T) {
			out, err := Render(channel, "", payload)
			require.NoError(t, err)
			assert.Contains(t, out, "42")

			switch channel.Type() {
			case TypeTelegram, TypeEmail:
				// plain text / html bodies
			default:
				var doc map[string]any
				assert.NoError(t, json.Unmarshal([]byte(out), &doc), "channel %s must render JSON: %s", channel.Type(), out)
			}
		})
	}
}

func TestRender_CustomBody(t *testing.T) {
	slack, err := NewRegistry().Get(TypeSlack)
	require.NoError(t, err)

	out, err := Render(slack, `{"text": "{{ .Project.Name }} / {{ .Name }} #{{ .Task.ID }}"}`, samplePayload())
	require.NoError(t, err)
	assert.Equal(t, `{"text": "Homelab / Deploy <prod> #42"}`, out)
}

func TestRender_EmailEscapesHTML(t *testing.T) {
	email, err := NewRegistry().Get(TypeEmail)
	require.NoError(t, err)

	out, err := Render(email, "", samplePayload())
	require.NoError(t, err)
	assert.Contains(t, out, "Deploy &lt;prod&gt;")
}

func TestRender_Errors(t *testing.T) {
	slack, err := NewRegistry().Get(TypeSlack)
	require.NoError(t, err)

	_, err = Render(slack, "{{ .Missing.Field }}", samplePayload())
	assert.ErrorContains(t, err, "can not render")

	_, err = Render(slack, "{{ if }}", samplePayload())
	assert.ErrorContains(t, err, "invalid message template")

	_, err = Render(slack, "{{ \"\" }}", samplePayload())
	assert.ErrorContains(t, err, "empty")
}

func TestValidateBody(t *testing.T) {
	slack, err := NewRegistry().Get(TypeSlack)
	require.NoError(t, err)

	assert.NoError(t, ValidateBody(slack, ""))
	assert.NoError(t, ValidateBody(slack, "{{ .Name }}"))
	assert.Error(t, ValidateBody(slack, "{{ .Name "))

	email, err := NewRegistry().Get(TypeEmail)
	require.NoError(t, err)
	assert.Error(t, ValidateBody(email, "{{ end }}"))
}

func TestEventForStatus(t *testing.T) {
	tests := []struct {
		status task_logger.TaskStatus
		event  db.AlertEvent
		ok     bool
	}{
		{task_logger.TaskSuccessStatus, db.AlertEventSuccess, true},
		{task_logger.TaskFailStatus, db.AlertEventError, true},
		{task_logger.TaskWaitingConfirmation, db.AlertEventWaitingConfirmation, true},
		{task_logger.TaskRunningStatus, "", false},
		{task_logger.TaskStoppedStatus, "", false},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			event, ok := EventForStatus(tt.status)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.event, event)
		})
	}
}
