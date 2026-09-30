package audit

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("fd00::/8")}
	manyHops := strings.TrimSuffix(strings.Repeat("198.51.100.7,", 40), ",")

	tests := []struct {
		name       string
		remoteAddr string
		xff        []string
		realIP     string
		want       string
	}{
		{"untrusted peer ignores headers", "203.0.113.9:1234", []string{"1.2.3.4"}, "5.6.7.8", "203.0.113.9"},
		{"client before trusted proxy", "10.0.0.1:80", []string{"198.51.100.7, 10.0.0.2"}, "", "198.51.100.7"},
		{"right-most untrusted hop", "10.0.0.1:80", []string{"1.1.1.1, 198.51.100.7"}, "", "198.51.100.7"},
		{"all hops trusted", "10.0.0.1:80", []string{"10.0.0.3, 10.0.0.2"}, "", "10.0.0.3"},
		{"several header lines", "10.0.0.1:80", []string{"198.51.100.7", "10.0.0.2"}, "", "198.51.100.7"},
		{"malformed hop", "10.0.0.1:80", []string{"198.51.100.7, garbage"}, "", "10.0.0.1"},
		{"hop with port", "10.0.0.1:80", []string{"198.51.100.7:443"}, "", "10.0.0.1"},
		{"too many hops", "10.0.0.1:80", []string{manyHops}, "", "10.0.0.1"},
		{"oversized header", "10.0.0.1:80", []string{strings.Repeat(" ", 5000) + "198.51.100.7"}, "", "10.0.0.1"},
		{"x-real-ip", "10.0.0.1:80", nil, "198.51.100.7", "198.51.100.7"},
		{"invalid x-real-ip", "10.0.0.1:80", nil, "nope", "10.0.0.1"},
		{"xff wins over x-real-ip", "10.0.0.1:80", []string{"198.51.100.7"}, "5.6.7.8", "198.51.100.7"},
		{"ipv6 peer", "[2001:db8::1]:443", nil, "", "2001:db8::1"},
		{"ipv6 trusted proxy", "[fd00::1]:443", []string{"2001:db8::9"}, "", "2001:db8::9"},
		{"ipv4-mapped peer", "[::ffff:10.0.0.1]:80", []string{"198.51.100.7"}, "", "198.51.100.7"},
		{"ipv4-mapped hop", "10.0.0.1:80", []string{"::ffff:198.51.100.7"}, "", "198.51.100.7"},
		{"unparsable peer", "@", []string{"198.51.100.7"}, "", "@"},
		{"zoned peer", "[fe80::1%eth0]:443", nil, "", "fe80::1"},
		{"zoned hop", "10.0.0.1:80", []string{"fe80::9%eth0"}, "", "fe80::9"},
		{"zoned x-real-ip", "10.0.0.1:80", nil, "fe80::7%eth0", "fe80::7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			for _, v := range tt.xff {
				header.Add("X-Forwarded-For", v)
			}
			if tt.realIP != "" {
				header.Set("X-Real-IP", tt.realIP)
			}
			assert.Equal(t, tt.want, clientIP(tt.remoteAddr, header, trusted))
		})
	}
}

func TestClientIP_RepeatedRealIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	header := http.Header{}
	header.Add("X-Real-IP", "198.51.100.66")
	header.Add("X-Real-IP", "203.0.113.9")

	assert.Equal(t, "10.0.0.1", clientIP("10.0.0.1:80", header, trusted))
}

func TestClientIP_NoTrustedProxies(t *testing.T) {
	assert.Equal(t, "10.0.0.1", clientIP("10.0.0.1:80", http.Header{"X-Forwarded-For": {"1.2.3.4"}}, nil))
}

func TestRequestMiddleware_UserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		want      string
	}{
		{"plain", "curl/8.5", "curl/8.5"},
		{"invalid utf-8", "a\xffb", "a�b"},
		{"too long", strings.Repeat("x", 2000), strings.Repeat("x", 1024)},
		{"cut on a rune boundary", strings.Repeat("x", 1022) + "€", strings.Repeat("x", 1022)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got RequestInfo
			inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got, _ = RequestFrom(r.Context())
			})
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("User-Agent", tt.userAgent)

			RequestMiddleware(nil)(inner).ServeHTTP(httptest.NewRecorder(), r)

			assert.Equal(t, tt.want, got.UserAgent)
		})
	}
}

func TestRequestMiddleware(t *testing.T) {
	var got RequestInfo
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		got, ok = RequestFrom(r.Context())
		require.True(t, ok)
	})

	r := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	r.RemoteAddr = "203.0.113.9:1234"
	r.Header.Set("User-Agent", "curl/8.5")
	r.Header.Set("X-Request-ID", "client-chosen")
	w := httptest.NewRecorder()

	RequestMiddleware(nil)(inner).ServeHTTP(w, r)

	_, err := uuid.Parse(got.ID)
	assert.NoError(t, err)
	assert.Equal(t, got.ID, w.Header().Get("X-Request-ID"))
	assert.Equal(t, "203.0.113.9", got.IP)
	assert.Equal(t, "curl/8.5", got.UserAgent)
}

func TestRequestMiddleware_OutsideProxyHeaders(t *testing.T) {
	var got RequestInfo
	var rewritten string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = RequestFrom(r.Context())
		rewritten = r.RemoteAddr
	})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.9:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")

	RequestMiddleware(nil)(handlers.ProxyHeaders(inner)).ServeHTTP(httptest.NewRecorder(), r)

	assert.Equal(t, "1.2.3.4", rewritten)
	assert.Equal(t, "203.0.113.9", got.IP)
}
