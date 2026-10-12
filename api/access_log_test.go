package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

// captureAccessLog turns on debug logging and records entries from the
// standard logger. The previous level and hooks are restored when the test
// ends so other tests in the package see an unchanged logger.
func captureAccessLog(t *testing.T) *test.Hook {
	t.Helper()

	logger := log.StandardLogger()
	prevLevel := logger.GetLevel()
	prevHooks := logger.ReplaceHooks(make(log.LevelHooks))
	t.Cleanup(func() {
		logger.SetLevel(prevLevel)
		logger.ReplaceHooks(prevHooks)
	})

	logger.SetLevel(log.DebugLevel)
	return test.NewLocal(logger)
}

func TestAccessLogMiddleware(t *testing.T) {
	hook := captureAccessLog(t)

	handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if len(hook.Entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(hook.Entries))
	}

	entry := hook.LastEntry()
	if entry.Level != log.DebugLevel {
		t.Errorf("expected debug level, got %v", entry.Level)
	}
	for _, field := range []string{"method", "path", "status", "size", "duration", "remote"} {
		if _, ok := entry.Data[field]; !ok {
			t.Errorf("missing expected log field %q", field)
		}
	}

	if entry.Data["method"] != "GET" {
		t.Errorf("expected method GET, got %v", entry.Data["method"])
	}
	if entry.Data["path"] != "/api/projects" {
		t.Errorf("expected path /api/projects, got %v", entry.Data["path"])
	}
	if entry.Data["status"] != http.StatusOK {
		t.Errorf("expected status %d, got %v", http.StatusOK, entry.Data["status"])
	}
}

func TestAccessLogMiddleware_SkipsPing(t *testing.T) {
	hook := captureAccessLog(t)

	handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, path := range []string{"/api/ping", "/semaphore/api/ping"} {
		hook.Reset()
		req := httptest.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if len(hook.Entries) != 0 {
			t.Errorf("expected no log entries for %s, got %d", path, len(hook.Entries))
		}
	}
}

// The websocket endpoint needs http.Hijacker from the wrapped ResponseWriter.
func TestAccessLogMiddleware_PreservesHijacker(t *testing.T) {
	captureAccessLog(t)

	var hijackable bool
	handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hijackable = w.(http.Hijacker)
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/ws")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if !hijackable {
		t.Error("ResponseWriter passed to the handler does not implement http.Hijacker")
	}
}
