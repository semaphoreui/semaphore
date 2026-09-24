package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/util"
)

func TestApiPing(t *testing.T) {
	util.Config = &util.ConfigType{
		Debugging: &util.DebuggingConfig{},
	}

	req, _ := http.NewRequest("GET", "/api/ping", nil)
	rr := httptest.NewRecorder()

	r := Route(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Response code should be 200 %d", rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type should be text/plain; charset=utf-8, got %s", ct)
	}

	if body := rr.Body.String(); body != "pong" {
		t.Errorf("Response body should be pong, got %s", body)
	}
}
