package tasks

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func TestHasSurveySecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"empty object", "{}", false},
		{"invalid json", "not-json", false},
		{"json array", `["a"]`, false},
		{"one secret", `{"passwd":"123456"}`, true},
		{"several secrets", `{"a":"1","b":"2"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, hasSurveySecrets(tt.input))
		})
	}
}

func TestTaskSecretExpireAt(t *testing.T) {
	tests := []struct {
		name           string
		maxDurationSec int
		expectedTTL    time.Duration
	}{
		{"unlimited task duration", 0, 24 * time.Hour},
		{"limited task duration", 3600, time.Hour + time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.Config = &util.ConfigType{MaxTaskDurationSec: tt.maxDurationSec}

			expireAt := taskSecretExpireAt()

			ttl := expireAt.Sub(tz.Now())
			assert.InDelta(t, tt.expectedTTL.Seconds(), ttl.Seconds(), 5)
		})
	}
}
