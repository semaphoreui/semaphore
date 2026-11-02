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

func TestConfigType_SessionInactivityTimeout(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConfigType
		expected time.Duration
	}{
		{"nil config", nil, DefaultSessionInactivityTimeout},
		{"auth section not configured", &ConfigType{}, DefaultSessionInactivityTimeout},
		{"zero hours means built-in default", &ConfigType{Auth: &AuthConfig{SessionInactivityHours: 0}}, DefaultSessionInactivityTimeout},
		{"negative hours means built-in default", &ConfigType{Auth: &AuthConfig{SessionInactivityHours: -5}}, DefaultSessionInactivityTimeout},
		{"configured hours", &ConfigType{Auth: &AuthConfig{SessionInactivityHours: 8}}, 8 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.SessionInactivityTimeout())
		})
	}
}

func TestAuthConfig_LoadFromEnvironment(t *testing.T) {
	t.Setenv("SEMAPHORE_AUTH_MAX_SESSION_LIFE_HOURS", "36")
	t.Setenv("SEMAPHORE_AUTH_SESSION_INACTIVITY_HOURS", "8")

	cfg := &ConfigType{}
	_, err := loadEnvironmentToObject(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.Auth)
	assert.Equal(t, 36, cfg.Auth.MaxSessionLifeHours)
	assert.Equal(t, 36*time.Hour, cfg.MaxSessionLife())
	assert.Equal(t, 8, cfg.Auth.SessionInactivityHours)
	assert.Equal(t, 8*time.Hour, cfg.SessionInactivityTimeout())
}

func TestAuthConfig_ValidateRejectsNegative(t *testing.T) {
	assert.Error(t, validate(&AuthConfig{MaxSessionLifeHours: -1}))
	assert.NoError(t, validate(&AuthConfig{MaxSessionLifeHours: 0}))
	assert.NoError(t, validate(&AuthConfig{MaxSessionLifeHours: 720}))
	assert.Error(t, validate(&AuthConfig{SessionInactivityHours: -1}))
	assert.NoError(t, validate(&AuthConfig{SessionInactivityHours: 0}))
	assert.NoError(t, validate(&AuthConfig{SessionInactivityHours: 48}))
}
