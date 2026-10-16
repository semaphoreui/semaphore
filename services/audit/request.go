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
				IP:        ClientIP(r.RemoteAddr, r.Header, trusted),
				UserAgent: TruncateName(r.UserAgent(), maxUserAgentBytes),
			}
			w.Header().Set("X-Request-ID", info.ID)
			next.ServeHTTP(w, r.WithContext(WithRequest(r.Context(), info)))
		})
	}
}

// ClientIP returns the client address of a request. Forwarded headers are
// used only when the peer is in trusted.
func ClientIP(remoteAddr string, header http.Header, trusted []netip.Prefix) string {
	peer, ok := parsePeer(remoteAddr)
	if !ok {
		return TruncateName(remoteAddr, maxIPBytes)
	}
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

// FromTrustedProxy reports whether the peer address is in trusted.
func FromTrustedProxy(remoteAddr string, trusted []netip.Prefix) bool {
	peer, ok := parsePeer(remoteAddr)
	return ok && isTrusted(peer, trusted)
}

func parsePeer(remoteAddr string) (netip.Addr, bool) {
	addrPort, err := netip.ParseAddrPort(remoteAddr)
	peer := addrPort.Addr()
	if err != nil {
		if peer, err = netip.ParseAddr(remoteAddr); err != nil {
			return netip.Addr{}, false
		}
	}
	return peer.Unmap().WithZone(""), true
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
