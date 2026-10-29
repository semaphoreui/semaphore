package util

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

type SubscriptionConfig struct {
	// Key is a subscription key or token that can be set via config.
	// When this is set, subscription activation from the web interface is disabled.
	Key       string `json:"key,omitempty" db:"-" env:"SEMAPHORE_SUBSCRIPTION_KEY,sensitive"`
	KeyFile   string `json:"key_file,omitempty" db:"-" env:"SEMAPHORE_SUBSCRIPTION_KEY_FILE"`
	ServerURL string `json:"server_url,omitempty" env:"SEMAPHORE_SUBSCRIPTION_SERVER_URL" default:"https://portal.semaphoreui.com/billing"`

	// Proxy sends the requests to the subscription server through an HTTP proxy.
	// It affects only those requests: tasks and other outgoing connections are
	// not routed through it.
	Proxy *SubscriptionProxyConfig `json:"proxy,omitempty"`
}

// SubscriptionProxyConfig is the proxy used to reach the subscription server.
// An installation that cannot change the environment of the server process
// configures the proxy here instead of through HTTP_PROXY.
type SubscriptionProxyConfig struct {
	// URL is the proxy endpoint, for example http://proxy.example.com:3128.
	// Schemes http, https, socks5 and socks5h are supported. When empty, the
	// proxy variables of the server process are used (HTTP_PROXY, HTTPS_PROXY,
	// NO_PROXY), as for every other outgoing request.
	URL string `json:"url,omitempty" env:"SEMAPHORE_SUBSCRIPTION_PROXY_URL"`
	// Username authenticates against the proxy. It overrides a user name given
	// in URL, and keeps the credentials out of the logged proxy address.
	Username string `json:"username,omitempty" env:"SEMAPHORE_SUBSCRIPTION_PROXY_USERNAME"`
	// Password belongs to Username. It overrides a password given in URL.
	Password string `json:"password,omitempty" env:"SEMAPHORE_SUBSCRIPTION_PROXY_PASSWORD,sensitive"`
}

// IsConfigured reports whether a proxy address is set. Credentials alone do not
// configure a proxy.
func (c *SubscriptionProxyConfig) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.URL) != ""
}

// ProxyURL returns the proxy address with the configured credentials applied,
// or nil when no proxy is configured.
func (c *SubscriptionProxyConfig) ProxyURL() (*url.URL, error) {
	if !c.IsConfigured() {
		return nil, nil
	}

	proxyURL, err := url.Parse(strings.TrimSpace(c.URL))
	if err != nil {
		// url.Error prints the URL it failed on, and that URL may carry
		// credentials. Keep the reason, drop the URL: the error ends up in the
		// log and in an API response.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("subscription.proxy.url is not a valid URL: %w", err)
	}

	// Schemes http.Transport can dial a proxy with.
	proxySchemes := []string{"http", "https", "socks5", "socks5h"}

	scheme := strings.ToLower(proxyURL.Scheme)
	if !slices.Contains(proxySchemes, scheme) {
		return nil, fmt.Errorf(
			"subscription.proxy.url must start with one of %s, got %q",
			strings.Join(proxySchemes, ", "), proxyURL.Scheme)
	}
	proxyURL.Scheme = scheme

	if proxyURL.Hostname() == "" {
		return nil, errors.New("subscription.proxy.url has no host")
	}

	if c.Username != "" {
		proxyURL.User = url.UserPassword(c.Username, c.Password)
	}

	return proxyURL, nil
}

// Validate checks the proxy settings at config load, so a typo in the proxy
// address is reported at startup and not on the first call to the portal.
func (c *SubscriptionConfig) Validate() error {
	if c == nil {
		return nil
	}
	_, err := c.Proxy.ProxyURL()
	return err
}

// NewHTTPClient builds the client talking to the subscription server. Keep the
// returned client for the lifetime of its owner instead of building one per
// request: a new client keeps no connection pool and no proxy connection.
func (c *SubscriptionConfig) NewHTTPClient(timeout time.Duration) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if c != nil && c.Proxy.IsConfigured() {
		proxyURL, err := c.Proxy.ProxyURL()
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

// SubscriptionClient hands out the HTTP client for the subscription server. It
// builds the client on first use — the configuration is already loaded and
// validated by then — and reuses it afterwards, so the connection to the proxy
// and to the portal is pooled. Keep one per owner, not one per request.
type SubscriptionClient struct {
	timeout time.Duration
	once    sync.Once
	client  *http.Client
	err     error
}

func NewSubscriptionClient(timeout time.Duration) *SubscriptionClient {
	return &SubscriptionClient{timeout: timeout}
}

// Client returns the shared client, or the error the proxy settings are wrong
// with.
func (c *SubscriptionClient) Client() (*http.Client, error) {
	c.once.Do(func() {
		var subscription *SubscriptionConfig
		if Config != nil {
			subscription = Config.Subscription
		}
		c.client, c.err = subscription.NewHTTPClient(c.timeout)
	})
	return c.client, c.err
}
