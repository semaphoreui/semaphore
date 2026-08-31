package util

import (
	"math"
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
		{"unset field", &RunnerConfig{}, time.Second},
		{"zero falls back", &RunnerConfig{CheckIntervalSeconds: 0}, time.Second},
		{"negative falls back", &RunnerConfig{CheckIntervalSeconds: -5}, time.Second},
		{"configured", &RunnerConfig{CheckIntervalSeconds: 10}, 10 * time.Second},
		{"large", &RunnerConfig{CheckIntervalSeconds: 300}, 5 * time.Minute},
		{"max representable", &RunnerConfig{CheckIntervalSeconds: maxRunnerCheckIntervalSec},
			time.Duration(maxRunnerCheckIntervalSec) * time.Second},
		{"overflows falls back", &RunnerConfig{CheckIntervalSeconds: maxRunnerCheckIntervalSec + 1}, time.Second},
		{"far past overflow falls back", &RunnerConfig{CheckIntervalSeconds: math.MaxInt64}, time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &ConfigType{Runner: tt.runner}
			assert.Equal(t, tt.expected, conf.RunnerCheckInterval())
		})
	}
}

// The result must always be safe for time.NewTicker.
func TestRunnerCheckIntervalNeverPanicsTicker(t *testing.T) {
	for _, sec := range []int{0, -1, 1, 300, maxRunnerCheckIntervalSec, maxRunnerCheckIntervalSec + 1, math.MaxInt64} {
		conf := &ConfigType{Runner: &RunnerConfig{CheckIntervalSeconds: sec}}
		d := conf.RunnerCheckInterval()

		assert.Positive(t, d, "interval must stay positive for CheckIntervalSeconds=%d", sec)
		assert.NotPanics(t, func() {
			time.NewTicker(d).Stop()
		}, "time.NewTicker must accept the interval for CheckIntervalSeconds=%d", sec)
	}
}

// The default tag must match the constant.
func TestRunnerCheckIntervalDefaultMatchesTag(t *testing.T) {
	conf := &ConfigType{Runner: &RunnerConfig{}}
	loadDefaultsToObject(conf)

	assert.Equal(t, defaultRunnerCheckIntervalSec, conf.Runner.CheckIntervalSeconds)
	assert.Equal(t, time.Second, conf.RunnerCheckInterval())
}
