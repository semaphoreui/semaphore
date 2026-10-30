// Package apiclient is a Go client for the Semaphore UI REST API.
//
// It is the transport layer of the `semaphore client` commands and is written
// so that another program can embed it: it knows nothing about the server
// configuration (util.Config), keeps no global state and takes everything it
// needs through Options. Authentication is an API token sent as
// `Authorization: Bearer <token>`; the session-cookie flow used by Login exists
// only to obtain such a token.
package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// DefaultTimeout bounds a single request when Options.Timeout is zero.
const DefaultTimeout = 30 * time.Second

// Options configures a Client.
type Options struct {
	// BaseURL is the address of the Semaphore UI instance, e.g.
	// "https://semaphore.example.com" or "https://host/semaphore". A trailing
	// slash or a trailing "/api" is tolerated and stripped.
	BaseURL string
	// Token is the API token. Empty when the client is only going to Login.
	Token string
	// CACertFile is a PEM bundle used to verify the server certificate in
	// addition to the system roots. Preferred over Insecure.
	CACertFile string
	// Insecure disables server certificate verification entirely.
	Insecure bool
	// Timeout bounds the dial, the TLS handshake and the wait for response
	// headers of every request. Zero means DefaultTimeout.
	Timeout time.Duration
	// UserAgent is sent with every request; empty means "semaphore-apiclient".
	UserAgent string
	// Debug, when not nil, receives a trace of every request and response.
	// Authorization and cookie headers are masked before they are written.
	Debug io.Writer
}

// Client talks to one Semaphore UI instance.
type Client struct {
	baseURL   *url.URL
	token     string
	userAgent string
	http      *http.Client
	debug     io.Writer
}

// New builds a Client from opts. It validates the URL and loads the CA bundle
// but performs no network request.
func New(opts Options) (*Client, error) {
	base, err := normalizeBaseURL(opts.BaseURL)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if opts.Insecure {
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // explicit user choice, warned about by the CLI
	}
	if opts.CACertFile != "" {
		pool, err := loadCertPool(opts.CACertFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	dialer := &net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}

	// Cookies matter only between Login and Logout; a token-authenticated
	// request never receives one. The jar is in memory and dies with the client.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = "semaphore-apiclient"
	}

	return &Client{
		baseURL:   base,
		token:     opts.Token,
		userAgent: userAgent,
		debug:     opts.Debug,
		http: &http.Client{
			Jar: jar,
			Transport: &http.Transport{
				TLSClientConfig:       tlsConfig,
				DialContext:           dialer.DialContext,
				TLSHandshakeTimeout:   timeout,
				ResponseHeaderTimeout: timeout,
				IdleConnTimeout:       90 * time.Second,
				ExpectContinueTimeout: time.Second,
				Proxy:                 http.ProxyFromEnvironment,
			},
			// A redirect could carry the bearer token to another host, so the
			// client never follows one; the user must give the final URL.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return errors.New("redirects are not followed; use the final server URL")
			},
		},
	}, nil
}

// BaseURL returns the normalized server address without a trailing slash.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}

// Token returns the API token the client authenticates with.
func (c *Client) Token() string {
	return c.token
}

// SetToken replaces the API token, e.g. right after Login created one.
func (c *Client) SetToken(token string) {
	c.token = token
}

// URL builds an absolute API URL for an API path such as "/project/1/tasks"
// or "/tasks/last?count=10" (a query string is kept as a query string).
func (c *Client) URL(apiPath string) string {
	path, query, _ := strings.Cut(apiPath, "?")
	u := *c.baseURL
	u.Path = strings.TrimSuffix(u.Path, "/") + "/api" + "/" + strings.TrimPrefix(path, "/")
	u.RawQuery = query
	return u.String()
}

// Do sends a JSON request and decodes a JSON response into out (which may be
// nil). in may be nil, an io.Reader sent as-is, or a value marshalled to JSON.
func (c *Client) Do(ctx context.Context, method, apiPath string, in, out any) error {
	_, err := c.Call(ctx, method, apiPath, in, out)
	return err
}

// Call is Do that also returns the response headers; the task list uses it to
// read the X-Has-Next pagination header.
func (c *Client) Call(ctx context.Context, method, apiPath string, in, out any) (http.Header, error) {
	resp, err := c.DoRaw(ctx, method, apiPath, in)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.Header, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.Header, fmt.Errorf("read response: %w", err)
	}
	c.trace("< body %s", redactJSON(body))

	if len(bytes.TrimSpace(body)) == 0 {
		return resp.Header, nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return resp.Header, fmt.Errorf("decode response from %s %s: %w", method, apiPath, err)
	}
	return resp.Header, nil
}

// DoRaw sends a request and returns the response with its body still open.
// A non-2xx status is turned into an *APIError and the body is consumed.
// The caller must close the body of a successful response.
func (c *Client) DoRaw(ctx context.Context, method, apiPath string, in any) (*http.Response, error) {
	var body io.Reader
	contentType := ""
	switch v := in.(type) {
	case nil:
	case io.Reader:
		body = v
	default:
		buf, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(buf)
		contentType = "application/json"
		c.trace("> body %s", redactJSON(buf))
	}

	req, err := http.NewRequestWithContext(ctx, method, c.URL(apiPath), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	c.trace("> %s %s", req.Method, req.URL)
	c.traceHeaders("> ", req.Header)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &TransportError{Op: method + " " + apiPath, Err: err}
	}

	c.trace("< %s", resp.Status)
	c.traceHeaders("< ", resp.Header)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		return nil, newAPIError(resp)
	}
	return resp, nil
}

// NormalizeURL returns the canonical form of a server address — the one the
// client really talks to, without a trailing slash, an "/api" suffix, a query
// or a fragment — so that two spellings of the same server compare equal. An
// address it cannot parse is returned trimmed, so that comparing two of them
// still says whether they are the same text.
func NormalizeURL(raw string) string {
	u, err := normalizeBaseURL(raw)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return u.String()
}

func normalizeBaseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("server URL is not set")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid server URL %q: scheme must be http or https", raw)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid server URL %q: host is missing", raw)
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.Path = strings.TrimSuffix(u.Path, "/api")
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}

func loadCertPool(file string) (*x509.CertPool, error) {
	pem, err := readFile(file)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("CA certificate %s contains no PEM certificates", file)
	}
	return pool, nil
}
