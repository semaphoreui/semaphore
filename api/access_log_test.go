package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestAccessLogMiddleware(t *testing.T) {
	hook := test.NewLocal(log.StandardLogger())

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
	hook := test.NewLocal(log.StandardLogger())

	handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/ping", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if len(hook.Entries) != 0 {
		t.Errorf("expected no log entries for /api/ping, got %d", len(hook.Entries))
	}
}
