package audit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExporter struct {
	started      bool
	stopped      bool
	err          error
	destinations []string
}

func (e *fakeExporter) Start() error {
	e.started = true
	return e.err
}

func (e *fakeExporter) Stop() {
	e.stopped = true
}

func (e *fakeExporter) DestinationIDs() []string {
	return e.destinations
}

func requestIDHeader(s *Service) string {
	w := httptest.NewRecorder()
	s.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).
		ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	return w.Header().Get("X-Request-ID")
}

func TestStartService_Disabled(t *testing.T) {
	store := &fakeStore{}
	exporter := &fakeExporter{}

	s, err := StartService(store, &util.AuditConfig{}, "", exporter)
	require.NoError(t, err)

	assert.Equal(t, Nop{}, s.Recorder())
	s.Stop()
	assert.False(t, exporter.started)
	assert.False(t, exporter.stopped)
	assert.Empty(t, store.rows)
	assert.Empty(t, requestIDHeader(s))
}

func TestStartService_RecordsStart(t *testing.T) {
	store := &fakeStore{}
	exporter := &fakeExporter{}

	s, err := StartService(store, &util.AuditConfig{Enabled: true, InstanceID: "test-instance"}, "node-a", exporter)
	require.NoError(t, err)

	assert.True(t, exporter.started)
	require.Len(t, store.rows, 1)
	envelope, err := EnvelopeFromRow(store.rows[0])
	require.NoError(t, err)
	assert.Equal(t, "audit.lifecycle", envelope.EventCode)
	assert.Equal(t, "start", envelope.Action)
	assert.Equal(t, SystemActor(ComponentServer), envelope.Actor)
	assert.Equal(t, "node-a", envelope.NodeID)
	assert.JSONEq(t, `{"destinations":[]}`, string(envelope.Metadata))
	assert.NotEmpty(t, requestIDHeader(s))

	s.Stop()
	assert.True(t, exporter.stopped)
}

func TestStartService_ExporterError(t *testing.T) {
	store := &fakeStore{}

	_, err := StartService(store, &util.AuditConfig{Enabled: true, InstanceID: "test-instance"}, "", &fakeExporter{err: errors.New("bad ca")})

	assert.ErrorContains(t, err, "bad ca")
	assert.Empty(t, store.rows)
}
