//go:build !windows
// +build !windows

package cmd

import (
	srslog "github.com/RackSec/srslog"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func initSyslog(conf *util.SyslogConfig) {
	if conf.Enabled {
		writer, err := srslog.Dial(conf.Network, conf.Address, srslog.LOG_DEBUG, conf.Tag)
		if err != nil {
			log.WithError(err).Fatal("Failed to connect to syslog")
			return
		}

		writer.SetFormatter(srslog.RFC5424Formatter)

		log.AddHook(&rfc5424Hook{writer: writer})
		log.Info("Syslog logging enabled (RFC 5424)")
	}
}

type rfc5424Hook struct {
	writer *srslog.Writer
}

func (h *rfc5424Hook) Levels() []log.Level {
	return log.AllLevels
}

func (h *rfc5424Hook) Fire(entry *log.Entry) error {
	msg, err := entry.String()
	if err != nil {
		return err
	}

	switch entry.Level {
	case log.PanicLevel, log.FatalLevel:
		return h.writer.Crit(msg)
	case log.ErrorLevel:
		return h.writer.Err(msg)
	case log.WarnLevel:
		return h.writer.Warning(msg)
	case log.InfoLevel:
		return h.writer.Info(msg)
	case log.DebugLevel, log.TraceLevel:
		return h.writer.Debug(msg)
	default:
		return h.writer.Info(msg)
	}
}
