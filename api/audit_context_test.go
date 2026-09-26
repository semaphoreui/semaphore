package api

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditRequestContextMiddlewareSourceIP(t *testing.T) {
	trustedProxy := netip.MustParsePrefix("10.0.0.0/8")
	tests := []struct {
		name       string
		remoteAddr string
		headers    http.Header
		wantIP     string
	}{
		{
			name:       "direct request uses TCP peer",
			remoteAddr: "198.51.100.10:42000",
			wantIP:     "198.51.100.10",
		},
		{
			name:       "direct IPv6 request uses TCP peer",
			remoteAddr: "[2001:db8::10]:42000",
			wantIP:     "2001:db8::10",
		},
		{
			name:       "untrusted peer cannot spoof forwarded headers",
			remoteAddr: "198.51.100.10:42000",
			headers: http.Header{
				"X-Forwarded-For": {"203.0.113.4"},
				"X-Real-Ip":       {"203.0.113.5"},
			},
			wantIP: "198.51.100.10",
		},
		{
			name:       "trusted proxy resolves one hop XFF address",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": {"198.51.100.11"},
			},
			wantIP: "198.51.100.11",
		},
		{
			name:       "mixed XFF chain resolves closest untrusted address",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": {"198.51.100.11, 203.0.113.3, 10.0.0.1"},
			},
			wantIP: "203.0.113.3",
		},
		{
			name:       "all trusted XFF addresses fall back to TCP peer",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": {"10.0.0.1"},
			},
			wantIP: "10.0.0.2",
		},
		{
			name:       "trusted proxy ignores multi element Forwarded header",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"Forwarded": {"for=198.51.100.12, for=10.0.0.1"},
			},
			wantIP: "10.0.0.2",
		},
		{
			name:       "trusted proxy accepts valid X Real IP address",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Real-Ip": {"198.51.100.13"},
				"Forwarded": {"for=203.0.113.6"},
			},
			wantIP: "198.51.100.13",
		},
		{
			name:       "malformed XFF falls back to TCP peer",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": {"not-an-ip"},
			},
			wantIP: "10.0.0.2",
		},
		{
			name:       "repeated XFF fields fall back to TCP peer",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": make([]string, maxForwardedHops+1),
			},
			wantIP: "10.0.0.2",
		},
		{
			name:       "oversized XFF value falls back to TCP peer",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": {strings.Repeat("1", maxForwardedHeaderBytes+1)},
			},
			wantIP: "10.0.0.2",
		},
		{
			name:       "too many XFF hops fall back to TCP peer",
			remoteAddr: "10.0.0.2:42000",
			headers: http.Header{
				"X-Forwarded-For": {
					strings.Repeat("198.51.100.11,", maxForwardedHops) + "198.51.100.11",
				},
			},
			wantIP: "10.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
			request.RemoteAddr = tt.remoteAddr
			request.Header = tt.headers.Clone()

			var got helpers.AuditRequestContext
			handler := AuditRequestContextMiddleware([]netip.Prefix{trustedProxy})(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					var ok bool
					got, ok = helpers.AuditRequestContextFrom(r)
					require.True(t, ok)
				},
			))
			handler.ServeHTTP(httptest.NewRecorder(), request)

			assert.Equal(t, tt.wantIP, got.SourceIP)
		})
	}
}

func TestAuditRequestContextMiddlewareGeneratesRequestID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	request.RemoteAddr = "198.51.100.10:42000"
	request.Header.Set("X-Request-ID", uuid.NewString())
	request.Header.Set("User-Agent", "  "+strings.Repeat("é", maxUserAgentBytes)+"  ")

	var got helpers.AuditRequestContext
	handler := AuditRequestContextMiddleware(nil)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var ok bool
			got, ok = helpers.AuditRequestContextFrom(r)
			require.True(t, ok)
		},
	))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	assert.NotEqual(t, request.Header.Get("X-Request-ID"), got.RequestID)
	assert.Equal(t, got.RequestID, response.Header().Get("X-Request-ID"))
	parsedID, err := uuid.Parse(got.RequestID)
	require.NoError(t, err)
	assert.Equal(t, uuid.RFC4122, parsedID.Variant())
	assert.Equal(t, uuid.Version(4), parsedID.Version())
	assert.Equal(t, strings.Repeat("é", maxUserAgentBytes/2), got.UserAgent)
}
