//go:build !pro

package helpers

import log "github.com/sirupsen/logrus"

// AppendEventToLog opens (or creates) the log file in append mode and writes the Event in key=value format.
func appendEventToLog(event log.Fields) error {
	return nil
}
