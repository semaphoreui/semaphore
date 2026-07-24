package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturingLogWriter struct {
	event pro_interfaces.EventLogRecord
}

func (c *capturingLogWriter) WriteEventLog(event pro_interfaces.EventLogRecord) error {
	c.event = event
	return nil
}

func (c *capturingLogWriter) WriteTaskLog(pro_interfaces.TaskLogRecord) error { return nil }
func (c *capturingLogWriter) WriteResult(any) error                           { return nil }

func TestEventLog_FillsAuditFields(t *testing.T) {
	store := sql.CreateTestStore()
	logWriter := &capturingLogWriter{}

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "auditor", Name: "Auditor", Email: "auditor@example.com",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/project/1/keys", nil)
	req.Header.Set("X-Real-IP", "203.0.113.7")
	req.Header.Set("User-Agent", "test-agent/1.0")
	req = SetContextValue(req, "store", store)
	req = SetContextValue(req, "log_writer", logWriter)

	EventLog(req, EventLogCreate, EventLogItem{
		UserID:      user.ID,
		ObjectType:  db.EventKey,
		ObjectID:    2,
		Description: "Access Key test\r\ncreated",
	})

	events, err := store.GetAllEvents(db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	evt := events[0]
	require.NotNil(t, evt.Action)
	assert.Equal(t, "create", *evt.Action)
	require.NotNil(t, evt.IP)
	assert.Equal(t, "203.0.113.7", *evt.IP)
	require.NotNil(t, evt.UserAgent)
	assert.Equal(t, "test-agent/1.0", *evt.UserAgent)
	require.NotNil(t, evt.Description)
	assert.Equal(t, "Access Key test  created", *evt.Description)

	assert.Equal(t, "create", logWriter.event.Action)
	assert.Equal(t, "203.0.113.7", logWriter.event.IP)
	assert.Equal(t, "test-agent/1.0", logWriter.event.UserAgent)
	require.NotNil(t, logWriter.event.ObjectType)
	assert.Equal(t, "key", *logWriter.event.ObjectType)
	require.NotNil(t, logWriter.event.ObjectID)
	assert.Equal(t, 2, *logWriter.event.ObjectID)
}

func TestExtractClientIP_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		realIP     string
		remoteAddr string
		expected   string
	}{
		{"X-Real-IP wins", "198.51.100.1", "10.0.0.1:1234", "198.51.100.1"},
		{"RemoteAddr host:port", "", "10.0.0.1:1234", "10.0.0.1"},
		{"RemoteAddr without port", "", "10.0.0.1", "10.0.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			assert.Equal(t, tt.expected, extractClientIP(req))
		})
	}
}
