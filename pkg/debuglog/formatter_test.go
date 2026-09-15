package debuglog

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogger returns a logrus logger wired with the FilteringFormatter at
// DEBUG level, writing into the returned buffer.
func newTestLogger(spec string) (*log.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := log.New()
	logger.SetOutput(buf)
	logger.SetLevel(log.DebugLevel)
	logger.SetFormatter(NewFilteringFormatter(&log.TextFormatter{DisableColors: true}, Parse(spec)))
	return logger, buf
}

func TestFilteringFormatter_SuppressesNonMatchingDebug(t *testing.T) {
	logger, buf := newTestLogger("runner")

	logger.WithField("context", "task_pool").Debug("should be hidden")

	assert.Empty(t, buf.String())
}

func TestFilteringFormatter_PassesMatchingDebug(t *testing.T) {
	logger, buf := newTestLogger("runner")

	logger.WithField("context", "runner").Debug("should appear")

	assert.Contains(t, buf.String(), "should appear")
	assert.Contains(t, buf.String(), "context=runner")
}

func TestFilteringFormatter_NeverTouchesHigherLevels(t *testing.T) {
	// Even a namespace that is NOT enabled must still emit at Info and above.
	logger, buf := newTestLogger("runner")

	logger.WithField("context", "task_pool").Info("info stays")
	logger.WithField("context", "task_pool").Warn("warn stays")
	logger.WithField("context", "task_pool").Error("error stays")

	out := buf.String()
	assert.Contains(t, out, "info stays")
	assert.Contains(t, out, "warn stays")
	assert.Contains(t, out, "error stays")
}

func TestFilteringFormatter_ContextlessDebug(t *testing.T) {
	t.Run("narrow filter hides contextless", func(t *testing.T) {
		logger, buf := newTestLogger("runner")
		logger.Debug("no context here")
		assert.Empty(t, buf.String())
	})

	t.Run("global filter shows contextless", func(t *testing.T) {
		logger, buf := newTestLogger("*")
		logger.Debug("no context here")
		assert.Contains(t, buf.String(), "no context here")
	})
}

func TestFilteringFormatter_NilFilterSuppressesDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New()
	logger.SetOutput(buf)
	logger.SetLevel(log.DebugLevel)
	logger.SetFormatter(NewFilteringFormatter(&log.TextFormatter{DisableColors: true}, nil))

	logger.WithField("context", "runner").Debug("hidden")
	logger.WithField("context", "runner").Info("shown")

	assert.NotContains(t, buf.String(), "hidden")
	assert.Contains(t, buf.String(), "shown")
}

func TestFilteringFormatter_SuppressedEntryWritesNoBytes(t *testing.T) {
	logger, buf := newTestLogger("runner")

	logger.WithField("context", "task_pool").Debug("x")

	// Suppressed debug entries must produce zero output, not a bare newline.
	require.Equal(t, 0, buf.Len())
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name      string
		level     log.Level
		formatter log.Formatter
		namespace string
		expected  bool
	}{
		{"debug without filter", log.DebugLevel, &log.TextFormatter{}, "runner", true},
		{"info without filter", log.InfoLevel, &log.TextFormatter{}, "runner", false},
		{"matching filter", log.DebugLevel, NewFilteringFormatter(nil, Parse("runner")), "runner", true},
		{"non-matching filter", log.DebugLevel, NewFilteringFormatter(nil, Parse("task_pool")), "runner", false},
		{"excluded namespace", log.DebugLevel, NewFilteringFormatter(nil, Parse("*,-runner")), "runner", false},
		{"nil filter", log.DebugLevel, NewFilteringFormatter(nil, nil), "runner", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New()
			logger.SetLevel(tt.level)
			logger.SetFormatter(tt.formatter)
			assert.Equal(t, tt.expected, Enabled(logger, tt.namespace))
		})
	}
}
