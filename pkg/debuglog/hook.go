package debuglog

import log "github.com/sirupsen/logrus"

// FilteringHook wraps a logrus Hook and suppresses DEBUG-level entries whose
// `context` namespace is not enabled by Filter before they reach the wrapped
// hook. Entries at any other level pass through unchanged.
type FilteringHook struct {
	Inner  log.Hook
	Filter *Filter
}

// NewFilteringHook wraps inner with debug-namespace filtering driven by filter.
func NewFilteringHook(inner log.Hook, filter *Filter) *FilteringHook {
	return &FilteringHook{Inner: inner, Filter: filter}
}

// Levels implements logrus.Hook.
func (h *FilteringHook) Levels() []log.Level {
	if h.Inner == nil {
		return nil
	}
	return h.Inner.Levels()
}

// Fire implements logrus.Hook.
func (h *FilteringHook) Fire(entry *log.Entry) error {
	if h.Inner == nil {
		return nil
	}

	if entry.Level == log.DebugLevel {
		ns, _ := entry.Data["context"].(string)
		if h.Filter == nil || !h.Filter.Enabled(ns) {
			return nil
		}
	}

	return h.Inner.Fire(entry)
}
