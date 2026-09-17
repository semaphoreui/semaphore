package api

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/api/helpers"
)

const (
	maxForwardedHeaderBytes = 4096
	maxForwardedHops        = 32
	maxUserAgentBytes       = 1024
)

func AuditRequestContextMiddleware(trustedProxies []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			peer, ok := requestPeerIP(r.RemoteAddr)
			securityContext := helpers.AuditRequestContext{
				RequestID: uuid.NewString(),
				UserAgent: normalizeUserAgent(r.UserAgent()),
			}
			if ok {
				securityContext.SourceIP = peer.String()
			}
			if ok && isTrustedProxy(peer, trustedProxies) {
				if source, ok := forwardedSourceIP(r, trustedProxies); ok {
					securityContext.SourceIP = source.String()
				}
			}

			w.Header().Set("X-Request-ID", securityContext.RequestID)
			next.ServeHTTP(w, helpers.SetAuditRequestContext(r, securityContext))
		})
	}
}

func requestPeerIP(remoteAddr string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		remoteAddr = host
	}
	addr, err := netip.ParseAddr(remoteAddr)
	return addr, err == nil
}

func isTrustedProxy(addr netip.Addr, trustedProxies []netip.Prefix) bool {
	for _, prefix := range trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func forwardedSourceIP(r *http.Request, trustedProxies []netip.Prefix) (netip.Addr, bool) {
	if values := r.Header.Values("X-Forwarded-For"); len(values) > 0 {
		return sourceFromXForwardedFor(values, trustedProxies)
	}
	if values := r.Header.Values("X-Real-IP"); len(values) > 0 {
		return sourceFromXRealIP(values)
	}
	return netip.Addr{}, false
}

func sourceFromXForwardedFor(values []string, trustedProxies []netip.Prefix) (netip.Addr, bool) {
	if len(values) > maxForwardedHops {
		return netip.Addr{}, false
	}

	length := 0
	hops := 0
	for _, value := range values {
		length += len(value)
		if length > maxForwardedHeaderBytes {
			return netip.Addr{}, false
		}
		hops++
		for i := range value {
			if value[i] == ',' {
				hops++
			}
		}
		if hops > maxForwardedHops {
			return netip.Addr{}, false
		}
	}

	addresses := make([]netip.Addr, 0, maxForwardedHops)
	for _, value := range values {
		for {
			part, remaining, hasMore := strings.Cut(value, ",")
			if len(addresses) == maxForwardedHops {
				return netip.Addr{}, false
			}
			addr, err := netip.ParseAddr(strings.TrimSpace(part))
			if err != nil {
				return netip.Addr{}, false
			}
			addresses = append(addresses, addr)
			if !hasMore {
				break
			}
			value = remaining
		}
	}

	for i := len(addresses) - 1; i >= 0; i-- {
		if !isTrustedProxy(addresses[i], trustedProxies) {
			return addresses[i], true
		}
	}
	return netip.Addr{}, false
}

func sourceFromXRealIP(values []string) (netip.Addr, bool) {
	if len(values) != 1 || len(values[0]) > maxForwardedHeaderBytes {
		return netip.Addr{}, false
	}
	addr, err := netip.ParseAddr(values[0])
	return addr, err == nil
}

func normalizeUserAgent(value string) string {
	value = strings.TrimSpace(strings.ToValidUTF8(value, "?"))
	if len(value) <= maxUserAgentBytes {
		return value
	}

	limit := 0
	for limit < len(value) {
		_, size := utf8.DecodeRuneInString(value[limit:])
		if limit+size > maxUserAgentBytes {
			break
		}
		limit += size
	}
	return value[:limit]
}
