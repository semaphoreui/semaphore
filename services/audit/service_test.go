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

// The caller's address must be resolved even with audit recording off: other
// features store it, and `handlers.ProxyHeaders` has already overwritten
// RemoteAddr from X-Forwarded-For by the time a handler runs.
func TestWrap_ResolvesClientAddrWhenDisabled(t *testing.T) {
	s, err := StartService(&fakeStore{}, &util.AuditConfig{}, "", &fakeExporter{})
	require.NoError(t, err)

	var got RequestInfo
	var ok bool

	handler := s.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, ok = RequestFrom(r.Context())
	}))

	r := httptest.NewRequest(http.MethodPost, "/api/integrations/some-alias", nil)
	r.RemoteAddr = "198.51.100.7:51234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")

	handler.ServeHTTP(httptest.NewRecorder(), r)

	require.True(t, ok, "the request info must be in the context")
	assert.Equal(t, "198.51.100.7", got.IP, "an untrusted peer cannot forge its address")
	assert.Empty(t, got.ID, "the request id stays an audit concern")
}

func TestWrap_HonoursForwardedForFromATrustedProxy(t *testing.T) {
	s, err := StartService(&fakeStore{},
		&util.AuditConfig{TrustedProxyCIDRs: []string{"198.51.100.0/24"}}, "", &fakeExporter{})
	require.NoError(t, err)

	var got RequestInfo
	handler := s.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = RequestFrom(r.Context())
	}))

	r := httptest.NewRequest(http.MethodPost, "/api/integrations/some-alias", nil)
	r.RemoteAddr = "198.51.100.7:51234"
	r.Header.Set("X-Forwarded-For", "203.0.113.9")

	handler.ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, "203.0.113.9", got.IP)
}
