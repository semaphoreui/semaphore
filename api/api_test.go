package api

import (
	"github.com/semaphoreui/semaphore/util"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestSwaggerEndpoint(t *testing.T) {
	util.Config = &util.ConfigType{
		Debugging: &util.DebuggingConfig{},
	}

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
	)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "swagger with trailing slash",
			path:           "/swagger/",
			expectedStatus: 200,
			expectedBody:   "Swagger UI",
		},
		{
			name:           "swagger without trailing slash",
			path:           "/swagger",
			expectedStatus: 200,
			expectedBody:   "Swagger UI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("Expected response to contain '%s', got: %s", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
