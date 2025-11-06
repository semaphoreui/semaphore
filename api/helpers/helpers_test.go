package helpers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// SetTestDelay sets a delay for testing slow network conditions
func SetTestDelay(delay time.Duration) func() {
	originalDelay := os.Getenv("DEBUG_DELAY")
	os.Setenv("DEBUG_DELAY", delay.String())
	return func() {
		if originalDelay == "" {
			os.Unsetenv("DEBUG_DELAY")
		} else {
			os.Setenv("DEBUG_DELAY", originalDelay)
		}
	}
}

func TestGetIntParam(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test/123", nil)
	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/test/{test_id}", mockParam)
	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Response code should be 200 %d", rr.Code)
	}
}

func mockParam(w http.ResponseWriter, r *http.Request) {
	_, err := GetIntParam("test_id", w, r)
	if err != nil {
		return
	}

	w.WriteHeader(200)
}
