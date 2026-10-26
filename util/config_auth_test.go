package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigType_MaxSessionLife(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConfigType
		expected time.Duration
	}{
		{"nil config", nil, 0},
		{"auth section not configured", &ConfigType{}, 0},
		{"zero hours means unlimited", &ConfigType{Auth: &AuthConfig{MaxSessionLifeHours: 0}}, 0},
		{"negative hours means unlimited", &ConfigType{Auth: &AuthConfig{MaxSessionLifeHours: -5}}, 0},
		{"configured hours", &ConfigType{Auth: &AuthConfig{MaxSessionLifeHours: 12}}, 12 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.MaxSessionLife())
		})
	}
}

func TestAuthConfig_LoadFromEnvironment(t *testing.T) {
	t.Setenv("SEMAPHORE_AUTH_MAX_SESSION_LIFE_HOURS", "36")

	cfg := &ConfigType{}
	_, err := loadEnvironmentToObject(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.Auth)
	assert.Equal(t, 36, cfg.Auth.MaxSessionLifeHours)
	assert.Equal(t, 36*time.Hour, cfg.MaxSessionLife())
}

func TestAuthConfig_ValidateRejectsNegative(t *testing.T) {
	assert.Error(t, validate(&AuthConfig{MaxSessionLifeHours: -1}))
	assert.NoError(t, validate(&AuthConfig{MaxSessionLifeHours: 0}))
	assert.NoError(t, validate(&AuthConfig{MaxSessionLifeHours: 720}))
}

func TestJWTAuthConfig_Validate(t *testing.T) {
	valid := JWTAuthConfig{
		Enabled: true,
		Header:  "X-Pomerium-Jwt-Assertion",
		JWKSURL: "https://auth.example.com/.well-known/pomerium/jwks.json",
	}

	tests := []struct {
		name    string
		mutate  func(c *JWTAuthConfig)
		wantErr string
	}{
		{"valid", func(c *JWTAuthConfig) {}, ""},
		{"valid http", func(c *JWTAuthConfig) { c.JWKSURL = "http://192.0.2.10:8080/jwks.json" }, ""},
		{"disabled skips checks", func(c *JWTAuthConfig) { c.Enabled = false; c.Header = ""; c.JWKSURL = "" }, ""},
		{"missing header", func(c *JWTAuthConfig) { c.Header = "" }, "header is not configured"},
		{"missing jwks url", func(c *JWTAuthConfig) { c.JWKSURL = "" }, "jwks_url is not configured"},
		{"no scheme", func(c *JWTAuthConfig) { c.JWKSURL = "auth.example.com/jwks.json" }, "must be an http or https URL"},
		{"no host", func(c *JWTAuthConfig) { c.JWKSURL = "https:///jwks.json" }, "must be an http or https URL"},
		{"other scheme", func(c *JWTAuthConfig) { c.JWKSURL = "file:///etc/jwks.json" }, "must be an http or https URL"},
		{"unparsable", func(c *JWTAuthConfig) { c.JWKSURL = "https://auth.example.com/%zz" }, "must be an http or https URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			tt.mutate(&c)
			err := c.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
