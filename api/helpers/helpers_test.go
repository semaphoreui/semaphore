package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// SetTestDelay sets a delay for testing slow network conditions.
// Cleanup is handled by testing.T.Setenv.
func SetTestDelay(t *testing.T, delay time.Duration) {
	t.Helper()
	t.Setenv("DEBUG_DELAY", delay.String())
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

func TestGetIntParamInvalidXHR(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test/abc", nil)
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/test/{test_id}", mockParam)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Response code should be %d %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetIntParamInvalidHTML(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test/abc", nil)
	req.Header.Set("Accept", "text/html")
	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/test/{test_id}", mockParam)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("Response code should be %d %d", http.StatusFound, rr.Code)
	}

	if rr.Header().Get("Location") != "/404" {
		t.Errorf("Location header should be /404, got %s", rr.Header().Get("Location"))
	}
}

func mockParam(w http.ResponseWriter, r *http.Request) {
	_, ok := GetIntParamOrAbort("test_id", w, r)
	if !ok {
		return
	}

	w.WriteHeader(200)
}
