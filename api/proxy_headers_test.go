package api

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

type seenRequest struct {
	remoteAddr string
	host       string
	scheme     string
}

func serveProxied(trusted []netip.Prefix, remoteAddr string) seenRequest {
	var seen seenRequest
	handler := ProxyHeadersMiddleware(trusted)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = seenRequest{remoteAddr: r.RemoteAddr, host: r.Host, scheme: r.URL.Scheme}
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = remoteAddr
	req.Host = "semaphore.internal"
	req.Header.Set("X-Forwarded-For", "198.51.100.7, 10.40.0.2")
	req.Header.Set("X-Forwarded-Host", "semaphore.example.com")
	req.Header.Set("X-Forwarded-Proto", "HTTPS")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	return seen
}

func TestProxyHeadersMiddleware(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.40.0.0/16")}

	tests := []struct {
		name       string
		trusted    []netip.Prefix
		remoteAddr string
		want       seenRequest
	}{
		{
			"trusted proxy",
			trusted,
			"10.40.0.1:4321",
			seenRequest{remoteAddr: "198.51.100.7", host: "semaphore.example.com", scheme: "https"},
		},
		{
			"untrusted peer",
			trusted,
			"203.0.113.9:4321",
			seenRequest{remoteAddr: "203.0.113.9:4321", host: "semaphore.internal"},
		},
		{
			"no trusted proxies",
			nil,
			"10.40.0.1:4321",
			seenRequest{remoteAddr: "10.40.0.1:4321", host: "semaphore.internal"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, serveProxied(tt.trusted, tt.remoteAddr))
		})
	}
}

func TestProxyHeadersMiddleware_IgnoresUnknownScheme(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.40.0.0/16")}

	var scheme string
	handler := ProxyHeadersMiddleware(trusted)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme = r.URL.Scheme
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.40.0.1:4321"
	req.Header.Set("X-Forwarded-Proto", "javascript")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Empty(t, scheme)
}

func TestProxyHeadersMiddleware_AccessLogRemote(t *testing.T) {
	hook := captureAccessLog(t)
	trusted := []netip.Prefix{netip.MustParsePrefix("10.40.0.0/16")}
	handler := ProxyHeadersMiddleware(trusted)(AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	for _, tt := range []struct{ remoteAddr, want string }{
		{"10.40.0.1:4321", "198.51.100.7"},
		{"203.0.113.9:4321", "203.0.113.9:4321"},
	} {
		hook.Reset()
		req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		req.RemoteAddr = tt.remoteAddr
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		handler.ServeHTTP(httptest.NewRecorder(), req)

		if assert.Len(t, hook.Entries, 1) {
			assert.Equal(t, tt.want, hook.LastEntry().Data["remote"], tt.remoteAddr)
		}
	}
}
