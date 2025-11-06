//go:build !windows
// +build !windows

package cmd

import (
	"log/syslog"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

func initSyslog(conf *util.SyslogConfig) {
	if conf.Enabled {
		hook, err := lSyslog.NewSyslogHook(conf.Network, conf.Address, syslog.LOG_DEBUG, conf.Tag)
		if err == nil {
			log.AddHook(hook)
		} else {
			log.WithError(err).Fatal("Failed to create syslog hook")
		}
	}
}
