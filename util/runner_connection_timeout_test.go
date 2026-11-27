package util

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerConnectionTimeout(t *testing.T) {
	tests := []struct {
		name     string
		runner   *RunnerConfig
		expected time.Duration
	}{
		{"no runner section", nil, 3 * time.Second},
		{"no connection section", &RunnerConfig{}, 3 * time.Second},
		{"unset field", &RunnerConfig{Connection: &RunnerConnectionConfig{}}, 3 * time.Second},
		{"zero falls back",
			&RunnerConfig{Connection: &RunnerConnectionConfig{TimeoutSeconds: 0}}, 3 * time.Second},
		{"negative falls back",
			&RunnerConfig{Connection: &RunnerConnectionConfig{TimeoutSeconds: -5}}, 3 * time.Second},
		{"configured",
			&RunnerConfig{Connection: &RunnerConnectionConfig{TimeoutSeconds: 30}}, 30 * time.Second},
		{"largest int",
			&RunnerConfig{Connection: &RunnerConnectionConfig{TimeoutSeconds: math.MaxInt}},
			time.Duration(clampTimeout(math.MaxInt)) * time.Second},
	}

	if intHoldsOverflowingInterval {
		tests = append(tests,
			struct {
				name     string
				runner   *RunnerConfig
				expected time.Duration
			}{"max representable",
				&RunnerConfig{Connection: &RunnerConnectionConfig{TimeoutSeconds: int(maxIntervalSec)}},
				time.Duration(maxRunnerConnectionTimeoutSec) * time.Second},
			struct {
				name     string
				runner   *RunnerConfig
				expected time.Duration
			}{"overflows falls back",
				&RunnerConfig{Connection: &RunnerConnectionConfig{TimeoutSeconds: int(maxIntervalSec) + 1}},
				3 * time.Second},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &ConfigType{Runner: tt.runner}
			assert.Equal(t, tt.expected, conf.RunnerConnectionTimeout())
		})
	}
}

func clampTimeout(sec int) int64 {
	v := int64(sec)
	if v > 0 && v <= maxRunnerConnectionTimeoutSec {
		return v
	}
	return int64(defaultRunnerConnectionTimeoutSec)
}

// The result must always be a usable, positive deadline.
func TestRunnerConnectionTimeoutAlwaysPositive(t *testing.T) {
	secs := []int{0, -1, 1, 300, math.MaxInt}
	if intHoldsOverflowingInterval {
		secs = append(secs, int(maxIntervalSec), int(maxIntervalSec)+1)
	}

	for _, sec := range secs {
		conf := &ConfigType{Runner: &RunnerConfig{
			Connection: &RunnerConnectionConfig{TimeoutSeconds: sec},
		}}

		assert.Positive(t, conf.RunnerConnectionTimeout(),
			"timeout must stay positive for TimeoutSeconds=%d", sec)
	}
}

// The default tag must match the constant.
func TestRunnerConnectionTimeoutDefaultMatchesTag(t *testing.T) {
	conf := &ConfigType{Runner: &RunnerConfig{Connection: &RunnerConnectionConfig{}}}
	require.NoError(t, loadDefaultsToObject(conf))

	assert.Equal(t, defaultRunnerConnectionTimeoutSec, conf.Runner.Connection.TimeoutSeconds)
	assert.Equal(t, 3*time.Second, conf.RunnerConnectionTimeout())
}

// A stuck request must not outlive the poll cycle that started it, otherwise
// the next cycle is skipped and the heartbeat stops.
func TestRunnerConnectionTimeoutBelowCheckInterval(t *testing.T) {
	assert.Less(t, defaultRunnerConnectionTimeoutSec, defaultRunnerCheckIntervalSec)
}
