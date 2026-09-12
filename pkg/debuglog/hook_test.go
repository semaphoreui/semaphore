package debuglog

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingHook struct {
	entries []*log.Entry
}

func (h *recordingHook) Levels() []log.Level {
	return log.AllLevels
}

func (h *recordingHook) Fire(entry *log.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
}

func TestFilteringHook_SuppressesNonMatchingDebug(t *testing.T) {
	inner := &recordingHook{}
	hook := NewFilteringHook(inner, Parse("runner"))

	err := hook.Fire(&log.Entry{Level: log.DebugLevel, Data: log.Fields{"context": "task_pool"}})

	require.NoError(t, err)
	assert.Empty(t, inner.entries)
}

func TestFilteringHook_PassesMatchingDebug(t *testing.T) {
	inner := &recordingHook{}
	hook := NewFilteringHook(inner, Parse("runner"))
	entry := &log.Entry{Level: log.DebugLevel, Data: log.Fields{"context": "runner"}}

	err := hook.Fire(entry)

	require.NoError(t, err)
	require.Len(t, inner.entries, 1)
	assert.Same(t, entry, inner.entries[0])
}

func TestFilteringHook_NeverTouchesHigherLevels(t *testing.T) {
	inner := &recordingHook{}
	hook := NewFilteringHook(inner, Parse("runner"))
	entry := &log.Entry{Level: log.InfoLevel, Data: log.Fields{"context": "task_pool"}}

	err := hook.Fire(entry)

	require.NoError(t, err)
	require.Len(t, inner.entries, 1)
	assert.Same(t, entry, inner.entries[0])
}

func TestFilteringHook_DelegatesLevels(t *testing.T) {
	inner := &recordingHook{}
	hook := NewFilteringHook(inner, Parse("runner"))

	assert.Equal(t, log.AllLevels, hook.Levels())
}
