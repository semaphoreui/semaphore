package jwt

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudience_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		audience Audience
		expected string
	}{
		{"empty", Audience{}, "null"},
		{"nil", nil, "null"},
		{"single", Audience{"semaphore"}, `"semaphore"`},
		{"multiple", Audience{"a", "b"}, `["a","b"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.audience)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(b))
		})
	}
}

func TestAudience_IsZero(t *testing.T) {
	assert.True(t, Audience(nil).IsZero())
	assert.True(t, Audience{}.IsZero())
	assert.False(t, Audience{"x"}.IsZero())
}
