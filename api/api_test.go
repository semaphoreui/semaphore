package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	_ "github.com/snikch/goodman/transaction"
)

func TestApiPing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/ping", nil)
	rr := httptest.NewRecorder()

	r := Route()

	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Response code should be 200 %d", rr.Code)
	}
}
