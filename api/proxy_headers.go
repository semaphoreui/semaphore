package api

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/semaphoreui/semaphore/services/audit"
)

// ProxyHeadersMiddleware applies the client address, host and scheme from
// X-Forwarded-For, X-Real-IP, X-Forwarded-Host and X-Forwarded-Proto, but only
// when the peer is one of the trusted proxies. With no trusted proxies the
// headers are ignored.
func ProxyHeadersMiddleware(trusted []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if len(trusted) == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !audit.FromTrustedProxy(r.RemoteAddr, trusted) {
				next.ServeHTTP(w, r)
				return
			}

			r.RemoteAddr = audit.ClientIP(r.RemoteAddr, r.Header, trusted)
			if host := r.Header.Get("X-Forwarded-Host"); host != "" {
				r.Host = host
			}
			if proto := strings.ToLower(r.Header.Get("X-Forwarded-Proto")); proto == "http" || proto == "https" {
				r.URL.Scheme = proto
			}

			next.ServeHTTP(w, r)
		})
	}
}
