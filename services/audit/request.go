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

func RequestMiddleware(trusted []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := RequestInfo{
				ID:        uuid.NewString(),
				IP:        clientIP(r.RemoteAddr, r.Header, trusted),
				UserAgent: TruncateName(r.UserAgent(), maxUserAgentBytes),
			}
			w.Header().Set("X-Request-ID", info.ID)
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
