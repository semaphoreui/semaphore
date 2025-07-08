package api

import (
	"github.com/semaphoreui/semaphore/util"
	"net/http"
	"net/http/httptest"
	"testing"
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
	)

	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Response code should be 200 %d", rr.Code)
	}
}
