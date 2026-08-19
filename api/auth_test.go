package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsStateChangingMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{http.MethodGet, false},
		{http.MethodHead, false},
		{http.MethodOptions, false},
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, true},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			assert.Equal(t, tt.expected, isStateChangingMethod(tt.method))
		})
	}
}

func TestRequestOriginHost(t *testing.T) {
	t.Run("from Origin header", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/users/1/password", nil)
		r.Header.Set("Origin", "https://semaphore.example.com")

		host, ok := requestOriginHost(r)
		assert.True(t, ok)
		assert.Equal(t, "semaphore.example.com", host)
	})

	t.Run("falls back to Referer", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/users/1/password", nil)
		r.Header.Set("Referer", "https://semaphore.example.com/project/1")

		host, ok := requestOriginHost(r)
		assert.True(t, ok)
		assert.Equal(t, "semaphore.example.com", host)
	})

	t.Run("Origin takes precedence over Referer", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/users/1/password", nil)
		r.Header.Set("Origin", "https://attacker.com")
		r.Header.Set("Referer", "https://semaphore.example.com/")

		host, ok := requestOriginHost(r)
		assert.True(t, ok)
		assert.Equal(t, "attacker.com", host)
	})

	t.Run("no headers", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/users/1/password", nil)

		_, ok := requestOriginHost(r)
		assert.False(t, ok)
	})
}

// newRecordingHandler returns an http.Handler that records whether it was
// called, used to assert that the middleware did or did not pass the request
// through.
func newRecordingHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusNoContent)
	})
}

func TestCsrfProtectionMiddleware(t *testing.T) {
	orig := util.WebHostURL
	defer func() { util.WebHostURL = orig }()

	webHost, err := url.Parse("https://semaphore.example.com")
	require.NoError(t, err)
	util.WebHostURL = webHost

	tests := []struct {
		name         string
		method       string
		host         string
		origin       string
		referer      string
		authHeader   string
		wantStatus   int
		wantForwPass bool
	}{
		{
			name:         "safe method is always allowed",
			method:       http.MethodGet,
			host:         "semaphore.example.com",
			origin:       "https://attacker.com",
			wantStatus:   http.StatusNoContent,
			wantForwPass: true,
		},
		{
			name:         "same origin POST is allowed",
			method:       http.MethodPost,
			host:         "semaphore.example.com",
			origin:       "https://semaphore.example.com",
			wantStatus:   http.StatusNoContent,
			wantForwPass: true,
		},
		{
			name:         "cross origin POST is blocked",
			method:       http.MethodPost,
			host:         "semaphore.example.com",
			origin:       "https://attacker.com",
			wantStatus:   http.StatusForbidden,
			wantForwPass: false,
		},
		{
			name:         "cross origin DELETE is blocked",
			method:       http.MethodDelete,
			host:         "semaphore.example.com",
			origin:       "https://attacker.com:1337",
			wantStatus:   http.StatusForbidden,
			wantForwPass: false,
		},
		{
			name:         "cross origin via Referer is blocked",
			method:       http.MethodPost,
			host:         "semaphore.example.com",
			referer:      "https://attacker.com/evil",
			wantStatus:   http.StatusForbidden,
			wantForwPass: false,
		},
		{
			name:         "missing origin and referer is allowed",
			method:       http.MethodPost,
			host:         "semaphore.example.com",
			wantStatus:   http.StatusNoContent,
			wantForwPass: true,
		},
		{
			name:         "bearer token bypasses origin check",
			method:       http.MethodPost,
			host:         "semaphore.example.com",
			origin:       "https://attacker.com",
			authHeader:   "Bearer sometoken",
			wantStatus:   http.StatusNoContent,
			wantForwPass: true,
		},
		{
			name:         "origin matching request host is allowed",
			method:       http.MethodPost,
			host:         "internal-proxy:3000",
			origin:       "http://internal-proxy:3000",
			wantStatus:   http.StatusNoContent,
			wantForwPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/api/users/1/password", nil)
			r.Host = tt.host
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				r.Header.Set("Referer", tt.referer)
			}
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}

			var forwarded bool
			w := httptest.NewRecorder()

			csrfProtectionMiddleware(newRecordingHandler(&forwarded)).ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantForwPass, forwarded)
		})
	}
}
