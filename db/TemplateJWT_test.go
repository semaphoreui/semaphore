package db

import (
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetConfig(maxTTL string) {
	util.Config = &util.ConfigType{
		JWT: &util.JWTConfig{
			MaxTTL: maxTTL,
		},
	}
}

func TestTemplateJWTParams_Validate(t *testing.T) {
	resetConfig("1h")

	tests := []struct {
		name    string
		params  *TemplateJWTParams
		wantErr string
	}{
		{
			"nil", nil, "",
		},
		{
			"disabled",
			&TemplateJWTParams{
				Enabled: false,
				TTL:     "999h",
			},
			"",
		},
		{
			"ok",
			&TemplateJWTParams{
				Enabled:  true,
				Audience: []string{"a"},
				TTL:      "30m",
			},
			"",
		},
		{
			"empty audience entry",
			&TemplateJWTParams{
				Enabled:  true,
				Audience: []string{""},
			},
			"audience entries must not be empty",
		},
		{
			"invalid ttl",
			&TemplateJWTParams{
				Enabled: true,
				TTL:     "nope",
			},
			"invalid JWT TTL",
		},
		{
			"non-positive ttl",
			&TemplateJWTParams{
				Enabled: true,
				TTL:     "0s",
			},
			"must be positive",
		},
		{
			"ttl exceeds max",
			&TemplateJWTParams{
				Enabled: true,
				TTL:     "2h",
			},
			"exceeds configured maximum",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestTemplateJWTParams_Validate_AudienceCap(t *testing.T) {
	resetConfig("1h")

	aud := make([]string, maxJWTAudienceEntries+1)
	for i := range aud {
		aud[i] = "x"
	}
	p := &TemplateJWTParams{
		Enabled:  true,
		Audience: aud,
	}
	assert.ErrorContains(t, p.Validate(), "at most")
}

func TestTemplateJWTParams_ScanValue(t *testing.T) {
	original := &TemplateJWTParams{
		Enabled:  true,
		Audience: []string{"a", "b"},
		TTL:      "30m",
	}

	v, err := original.Value()
	require.NoError(t, err)
	require.NotNil(t, v)

	var restored TemplateJWTParams
	require.NoError(t, restored.Scan(v))
	assert.Equal(t, *original, restored)

	var zero TemplateJWTParams
	require.NoError(t, zero.Scan(nil))
	assert.Equal(t, TemplateJWTParams{}, zero)
}
