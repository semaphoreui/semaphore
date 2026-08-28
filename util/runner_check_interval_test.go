package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunnerCheckInterval(t *testing.T) {
	tests := []struct {
		name     string
		runner   *RunnerConfig
		expected time.Duration
	}{
		{"unset runner section", nil, time.Second},
		{"unset field", &RunnerConfig{}, time.Second},
		// loadDefaultsToObject skips non-zero fields, so a configured 0 is
		// indistinguishable from unset; fall back rather than busy-loop.
		{"zero falls back", &RunnerConfig{CheckIntervalSeconds: 0}, time.Second},
		{"negative falls back", &RunnerConfig{CheckIntervalSeconds: -5}, time.Second},
		{"configured", &RunnerConfig{CheckIntervalSeconds: 10}, 10 * time.Second},
		{"large", &RunnerConfig{CheckIntervalSeconds: 300}, 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &ConfigType{Runner: tt.runner}
			assert.Equal(t, tt.expected, conf.RunnerCheckInterval())
		})
	}
}

// The default tag must match the constant, or config-file and no-config runs
// would poll at different rates.
func TestRunnerCheckIntervalDefaultMatchesTag(t *testing.T) {
	conf := &ConfigType{Runner: &RunnerConfig{}}
	loadDefaultsToObject(conf)

	assert.Equal(t, defaultRunnerCheckIntervalSec, conf.Runner.CheckIntervalSeconds)
	assert.Equal(t, time.Second, conf.RunnerCheckInterval())
}
