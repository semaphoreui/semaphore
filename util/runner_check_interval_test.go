package util

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// intHoldsOverflowingInterval reports whether an int is wide enough to express a
// value past the duration limit. On 32-bit targets (386, arm) it is not, so
// those cases cannot occur there and are skipped.
const intHoldsOverflowingInterval = int64(math.MaxInt) > maxRunnerCheckIntervalSec

// A variable, not a constant: int(maxIntervalSec) would be a constant
// conversion and fail to compile on 32-bit even inside a guarded branch.
var maxIntervalSec = maxRunnerCheckIntervalSec

func TestRunnerCheckInterval(t *testing.T) {
	tests := []struct {
		name     string
		runner   *RunnerConfig
		expected time.Duration
	}{
		{"unset runner section", nil, time.Second},
		{"unset field", &RunnerConfig{}, time.Second},
		{"zero falls back", &RunnerConfig{CheckIntervalSeconds: 0}, time.Second},
		{"negative falls back", &RunnerConfig{CheckIntervalSeconds: -5}, time.Second},
		{"configured", &RunnerConfig{CheckIntervalSeconds: 10}, 10 * time.Second},
		{"large", &RunnerConfig{CheckIntervalSeconds: 300}, 5 * time.Minute},
		{"largest int", &RunnerConfig{CheckIntervalSeconds: math.MaxInt},
			time.Duration(clampInterval(math.MaxInt)) * time.Second},
	}

	if intHoldsOverflowingInterval {
		tests = append(tests,
			struct {
				name     string
				runner   *RunnerConfig
				expected time.Duration
			}{"max representable", &RunnerConfig{CheckIntervalSeconds: int(maxIntervalSec)},
				time.Duration(maxRunnerCheckIntervalSec) * time.Second},
			struct {
				name     string
				runner   *RunnerConfig
				expected time.Duration
			}{"overflows falls back", &RunnerConfig{CheckIntervalSeconds: int(maxIntervalSec) + 1},
				time.Second},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &ConfigType{Runner: tt.runner}
			assert.Equal(t, tt.expected, conf.RunnerCheckInterval())
		})
	}
}

func clampInterval(sec int) int64 {
	v := int64(sec)
	if v > 0 && v <= maxRunnerCheckIntervalSec {
		return v
	}
	return int64(defaultRunnerCheckIntervalSec)
}

// The result must always be safe for time.NewTicker.
func TestRunnerCheckIntervalNeverPanicsTicker(t *testing.T) {
	secs := []int{0, -1, 1, 300, math.MaxInt}
	if intHoldsOverflowingInterval {
		secs = append(secs, int(maxIntervalSec), int(maxIntervalSec)+1)
	}

	for _, sec := range secs {
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
