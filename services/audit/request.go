package audit

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/google/uuid"
)

const (
	maxForwardedHeaderBytes = 4096
	maxForwardedHops        = 32
)

// RequestMiddleware resolves the caller's address against the trusted proxies
// and tags the request with an id.
func RequestMiddleware(trusted []netip.Prefix) func(http.Handler) http.Handler {
	return requestMiddleware(trusted, true)
}

// ClientAddrMiddleware resolves the caller's address without tagging the
// request. It must run whether or not audit recording is enabled: it is
// installed outside `handlers.ProxyHeaders`, which overwrites RemoteAddr from
// X-Forwarded-For without validating it, so it is the only place the real peer
// is still visible. Features that record the caller's address — the integration
// call history among them — would otherwise store a value the caller chose.
func ClientAddrMiddleware(trusted []netip.Prefix) func(http.Handler) http.Handler {
	return requestMiddleware(trusted, false)
}

func requestMiddleware(trusted []netip.Prefix, tagRequest bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := RequestInfo{
				IP:        clientIP(r.RemoteAddr, r.Header, trusted),
				UserAgent: TruncateName(r.UserAgent(), maxUserAgentBytes),
			}

			if tagRequest {
				info.ID = uuid.NewString()
				w.Header().Set("X-Request-ID", info.ID)
			}

			next.ServeHTTP(w, r.WithContext(WithRequest(r.Context(), info)))
		})
	}
}

func clientIP(remoteAddr string, header http.Header, trusted []netip.Prefix) string {
	addrPort, err := netip.ParseAddrPort(remoteAddr)
	peer := addrPort.Addr()
	if err != nil {
		if peer, err = netip.ParseAddr(remoteAddr); err != nil {
			return TruncateName(remoteAddr, maxIPBytes)
		}
	}
	peer = peer.Unmap().WithZone("")
	if !isTrusted(peer, trusted) {
		return peer.String()
	}

	if values := header.Values("X-Forwarded-For"); len(values) > 0 {
		if addr, ok := fromForwardedFor(values, trusted); ok {
			return addr.String()
		}
		return peer.String()
	}

	if values := header.Values("X-Real-IP"); len(values) == 1 {
		if realIP := strings.TrimSpace(values[0]); realIP != "" && len(realIP) <= maxIPBytes {
			if addr, err := netip.ParseAddr(realIP); err == nil {
				return addr.Unmap().WithZone("").String()
			}
		}
	}

	return peer.String()
}

func fromForwardedFor(values []string, trusted []netip.Prefix) (netip.Addr, bool) {
	size := 0
	for _, value := range values {
		size += len(value)
	}
	if size > maxForwardedHeaderBytes {
		return netip.Addr{}, false
	}

	var hops []string
	for _, value := range values {
		hops = append(hops, strings.Split(value, ",")...)
	}
	if len(hops) > maxForwardedHops {
		return netip.Addr{}, false
	}

	var addr netip.Addr
	for i := len(hops) - 1; i >= 0; i-- {
		parsed, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
		if err != nil {
			return netip.Addr{}, false
		}
		addr = parsed.Unmap().WithZone("")
		if !isTrusted(addr, trusted) {
			return addr, true
		}
	}
	return addr, true
}

func isTrusted(addr netip.Addr, trusted []netip.Prefix) bool {
	for _, prefix := range trusted {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
