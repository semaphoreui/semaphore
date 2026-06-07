package debuglog

import (
	log "github.com/sirupsen/logrus"
)

// FilteringFormatter wraps an inner logrus.Formatter and suppresses DEBUG-level
// entries whose `context` namespace is not enabled by Filter. Entries at any
// other level pass through to the inner formatter unchanged.
//
// Suppression is done by returning an empty payload: logrus writes the formatter
// result verbatim, so returning nil produces no output (not even a newline).
type FilteringFormatter struct {
	Inner  log.Formatter
	Filter *Filter
}

// NewFilteringFormatter wraps inner with debug-namespace filtering driven by
// filter. If inner is nil, a default logrus.TextFormatter is used.
func NewFilteringFormatter(inner log.Formatter, filter *Filter) *FilteringFormatter {
	if inner == nil {
		inner = &log.TextFormatter{}
	}
	return &FilteringFormatter{Inner: inner, Filter: filter}
}

// Format implements logrus.Formatter.
func (f *FilteringFormatter) Format(entry *log.Entry) ([]byte, error) {
	if entry.Level == log.DebugLevel {
		ns, _ := entry.Data["context"].(string)
		if f.Filter == nil || !f.Filter.Enabled(ns) {
			return nil, nil
		}
	}
	return f.Inner.Format(entry)
}
