package alerting

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

const (
	sendTimeout      = 15 * time.Second
	maxResponseBytes = 64 * 1024
)

// ValidateWebhookURL enforces the outbound policy for user supplied
// destinations: http(s) only, a host must be present, and hosts that point
// back at the server itself or at cloud metadata endpoints are rejected.
// Private (RFC 1918) ranges stay allowed because self-hosted Gotify or
// Rocket.Chat instances commonly live there.
func ValidateWebhookURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return common_errors.NewValidationError("URL can not be empty")
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return common_errors.NewValidationError("invalid URL")
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return common_errors.NewValidationError("URL must use http or https")
	}

	host := strings.ToLower(u.Hostname())
	if hostForbidden(host) {
		return common_errors.NewValidationError("URL host is not allowed")
	}
	if ip := net.ParseIP(host); ip != nil && ipForbidden(ip) {
		return common_errors.NewValidationError("URL host is not allowed")
	}
	return nil
}

func hostForbidden(host string) bool {
	return host == "localhost" ||
		strings.HasSuffix(host, ".localhost") ||
		host == "metadata" ||
		host == "metadata.google.internal" ||
		strings.HasSuffix(host, ".metadata.google.internal")
}

// ipForbidden rejects addresses a project webhook must never reach:
// loopback, link-local (cloud metadata lives on 169.254.169.254) and
// unspecified.
func ipForbidden(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// newHTTPClient returns the client used for a destination. Untrusted
// destinations resolve the host first and only dial addresses that pass
// ipForbidden, closing the DNS-rebinding gap between validation and dial.
// Redirects are never followed so a webhook can not bounce to a forbidden
// address either.
func newHTTPClient(trusted bool) *http.Client {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if !trusted {
		transport.Proxy = nil
		transport.DialContext = guardedDialContext
	}
	return &http.Client{
		Timeout:   sendTimeout,
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func guardedDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}

	dialer := net.Dialer{Timeout: 10 * time.Second}
	var lastErr error
	for _, ip := range ips {
		if ipForbidden(ip) {
			continue
		}
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr == nil {
		lastErr = common_errors.NewValidationError("URL host is not allowed")
	}
	return nil, lastErr
}

// postJSON sends body to dest.URL (or to the explicit url when given) and
// treats any of okCodes as success.
func postJSON(
	ctx context.Context,
	dest Destination,
	target string,
	body string,
	headers map[string]string,
	okCodes ...int,
) error {
	if target == "" {
		return common_errors.NewValidationError("URL is empty")
	}
	if !dest.Trusted {
		if err := ValidateWebhookURL(target); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := newHTTPClient(dest.Trusted).Do(req)
	if err != nil {
		return common_errors.NewUserError(err)
	}
	defer resp.Body.Close() //nolint:errcheck
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBytes))

	for _, code := range okCodes {
		if resp.StatusCode == code {
			return nil
		}
	}
	return common_errors.NewUserErrorS(fmt.Sprintf("unexpected response code %d", resp.StatusCode))
}
