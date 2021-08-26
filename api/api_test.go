package api

import (
	//_ "github.com/snikch/goodman/hooks"
	//_ "github.com/snikch/goodman/transaction"
	"net/http"
	"net/http/httptest"
	"testing"
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
