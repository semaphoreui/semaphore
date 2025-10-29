//go:build windows
// +build windows

package cmd

import (
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// initSyslog is disabled on Windows because the standard syslog package is not supported.
func initSyslog(conf *util.SyslogConfig) {
	if conf != nil && conf.Enabled {
		log.Warn("Syslog is not supported on Windows. The syslog log channel will be disabled.")
	}
	// no-op on Windows
}
