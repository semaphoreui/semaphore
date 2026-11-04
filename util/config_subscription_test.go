package util

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionProxyConfig_IsConfigured(t *testing.T) {
	var nilConfig *SubscriptionProxyConfig
	assert.False(t, nilConfig.IsConfigured())
	assert.False(t, (&SubscriptionProxyConfig{}).IsConfigured())
	assert.False(t, (&SubscriptionProxyConfig{URL: "  "}).IsConfigured())
	assert.False(t, (&SubscriptionProxyConfig{Username: "user"}).IsConfigured())
	assert.True(t, (&SubscriptionProxyConfig{URL: "http://proxy:3128"}).IsConfigured())
}

func TestSubscriptionProxyConfig_ProxyURL(t *testing.T) {
	tests := []struct {
		name     string
		proxy    *SubscriptionProxyConfig
		expected string
	}{
		{"not configured", nil, ""},
		{"empty url", &SubscriptionProxyConfig{}, ""},
		{"http", &SubscriptionProxyConfig{URL: "http://proxy.example.com:3128"}, "http://proxy.example.com:3128"},
		{"socks5", &SubscriptionProxyConfig{URL: "socks5://proxy.example.com:1080"}, "socks5://proxy.example.com:1080"},
		{"surrounding spaces", &SubscriptionProxyConfig{URL: " http://proxy:3128 "}, "http://proxy:3128"},
		{"upper case scheme", &SubscriptionProxyConfig{URL: "HTTP://proxy:3128"}, "http://proxy:3128"},
		{
			"credentials from config",
			&SubscriptionProxyConfig{URL: "http://proxy:3128", Username: "user", Password: "p4ss"},
			"http://user:p4ss@proxy:3128",
		},
		{
			"credentials in url are kept",
			&SubscriptionProxyConfig{URL: "http://user:p4ss@proxy:3128"},
			"http://user:p4ss@proxy:3128",
		},
		{
			"config credentials override the url",
			&SubscriptionProxyConfig{URL: "http://old:old@proxy:3128", Username: "user", Password: "p4ss"},
			"http://user:p4ss@proxy:3128",
		},
		{
			"empty password",
			&SubscriptionProxyConfig{URL: "http://proxy:3128", Username: "user"},
			"http://user:@proxy:3128",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyURL, err := tt.proxy.ProxyURL()
			require.NoError(t, err)

			if tt.expected == "" {
				assert.Nil(t, proxyURL)
				return
			}

			require.NotNil(t, proxyURL)
			assert.Equal(t, tt.expected, proxyURL.String())
		})
	}
}

func TestSubscriptionProxyConfig_ProxyURLErrors(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		contains string
	}{
		{"no scheme", "proxy.example.com:3128", "must start with one of"},
		{"unsupported scheme", "ftp://proxy:3128", "must start with one of"},
		{"no host", "http://", "has no host"},
		{"not an url", "http://proxy:3128/%zz", "not a valid URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := &SubscriptionProxyConfig{URL: tt.url}

			_, err := proxy.ProxyURL()
			assert.ErrorContains(t, err, tt.contains)

			err = (&SubscriptionConfig{Proxy: proxy}).Validate()
			assert.ErrorContains(t, err, tt.contains)
		})
	}
}

func TestSubscriptionConfig_Validate(t *testing.T) {
	var nilConfig *SubscriptionConfig
	assert.NoError(t, nilConfig.Validate())
	assert.NoError(t, (&SubscriptionConfig{}).Validate())
	assert.NoError(t, (&SubscriptionConfig{
		Proxy: &SubscriptionProxyConfig{URL: "https://proxy:8443", Username: "user", Password: "p4ss"},
	}).Validate())
}

func TestSubscriptionConfig_NewHTTPClient(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "https://portal.semaphoreui.com/billing", nil)
	require.NoError(t, err)

	t.Run("without proxy keeps the environment proxy", func(t *testing.T) {
		t.Setenv("HTTPS_PROXY", "http://env-proxy:3128")

		client, err := (&SubscriptionConfig{}).NewHTTPClient(7 * time.Second)
		require.NoError(t, err)
		assert.Equal(t, 7*time.Second, client.Timeout)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		require.NotNil(t, transport.Proxy)

		proxyURL, err := transport.Proxy(request)
		require.NoError(t, err)
		require.NotNil(t, proxyURL)
		assert.Equal(t, "http://env-proxy:3128", proxyURL.String())
	})

	t.Run("configured proxy wins over the environment", func(t *testing.T) {
		t.Setenv("HTTPS_PROXY", "http://env-proxy:3128")

		config := &SubscriptionConfig{Proxy: &SubscriptionProxyConfig{
			URL:      "http://proxy:3128",
			Username: "user",
			Password: "p4ss",
		}}

		client, err := config.NewHTTPClient(time.Second)
		require.NoError(t, err)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		require.NotNil(t, transport.Proxy)

		proxyURL, err := transport.Proxy(request)
		require.NoError(t, err)
		require.NotNil(t, proxyURL)
		assert.Equal(t, "http://user:p4ss@proxy:3128", proxyURL.String())
	})

	t.Run("invalid proxy is reported", func(t *testing.T) {
		config := &SubscriptionConfig{Proxy: &SubscriptionProxyConfig{URL: "ftp://proxy:3128"}}

		client, err := config.NewHTTPClient(time.Second)
		assert.Nil(t, client)
		assert.ErrorContains(t, err, "must start with one of")
	})

	t.Run("nil config builds a client", func(t *testing.T) {
		var config *SubscriptionConfig

		client, err := config.NewHTTPClient(time.Second)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestSubscriptionProxyConfig_FromEnvironment(t *testing.T) {
	t.Setenv("SEMAPHORE_SUBSCRIPTION_PROXY_URL", "http://proxy.example.com:3128")
	t.Setenv("SEMAPHORE_SUBSCRIPTION_PROXY_USERNAME", "user")
	t.Setenv("SEMAPHORE_SUBSCRIPTION_PROXY_PASSWORD", "p4ss")

	config := &ConfigType{}
	sensitive, err := loadEnvironmentToObject(config)
	require.NoError(t, err)
	assert.Contains(t, sensitive, "SEMAPHORE_SUBSCRIPTION_PROXY_PASSWORD")
	require.NoError(t, loadDefaultsToObject(config))

	require.NotNil(t, config.Subscription)
	require.NotNil(t, config.Subscription.Proxy)
	assert.Equal(t, "http://proxy.example.com:3128", config.Subscription.Proxy.URL)
	assert.Equal(t, "user", config.Subscription.Proxy.Username)
	assert.Equal(t, "p4ss", config.Subscription.Proxy.Password)
	assert.Equal(t, "https://portal.semaphoreui.com/billing", config.Subscription.ServerURL)
	assert.NoError(t, config.Subscription.Validate())
}

func TestSubscriptionClient_GetReusesTheClient(t *testing.T) {
	Config = &ConfigType{Subscription: &SubscriptionConfig{
		Proxy: &SubscriptionProxyConfig{URL: "http://proxy:3128"},
	}}
	defer func() { Config = nil }()

	holder := NewSubscriptionClient(5 * time.Second)

	first, err := holder.Client()
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, 5*time.Second, first.Timeout)

	second, err := holder.Client()
	require.NoError(t, err)
	assert.Same(t, first, second)
}

func TestSubscriptionClient_GetReportsInvalidProxy(t *testing.T) {
	Config = &ConfigType{Subscription: &SubscriptionConfig{
		Proxy: &SubscriptionProxyConfig{URL: "ftp://proxy:3128"},
	}}
	defer func() { Config = nil }()

	client, err := NewSubscriptionClient(time.Second).Client()
	assert.Nil(t, client)
	assert.ErrorContains(t, err, "must start with one of")
}
