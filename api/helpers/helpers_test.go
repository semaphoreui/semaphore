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
	err := os.Setenv("DEBUG_DELAY", delay.String())

	if err != nil {
		panic("Failed to set DEBUG_DELAY environment variable: " + err.Error())
	}

	return func() {
		if originalDelay == "" {
			err = os.Unsetenv("DEBUG_DELAY")
		} else {
			err = os.Setenv("DEBUG_DELAY", originalDelay)
		}

		if err != nil {
			panic("Failed to unset DEBUG_DELAY environment variable: " + err.Error())
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
	_, ok := GetIntParam("test_id", w, r)
	if !ok {
		return
	}

	w.WriteHeader(200)
}
