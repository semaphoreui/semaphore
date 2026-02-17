//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

func initSyslog(conf *util.SyslogConfig) {
	if !conf.Enabled {
		return
	}

	hook, err := lSyslog.NewSyslogHook(conf.Network, conf.Address, syslog.LOG_DEBUG, conf.Tag)
	if err != nil {
		log.WithError(err).Fatal("Failed to create syslog hook")
		return
	}

	switch conf.Format {
	case util.SyslogDefault:
		log.AddHook(hook)
		log.Info("Syslog logging enabled")
	case util.SyslogRFC5424:
		log.AddHook(&rfc5424Hook{
			writer: hook.Writer,
			tag:    conf.Tag,
		})
		log.Info("Syslog logging enabled (RFC 5424)")
	}
}

type rfc5424Hook struct {
	writer *syslog.Writer
	tag    string
}

func (h *rfc5424Hook) Levels() []log.Level {
	return log.AllLevels
}

func (h *rfc5424Hook) Fire(entry *log.Entry) error {
	msg := h.formatMessage(entry)

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

func (h *rfc5424Hook) formatMessage(entry *log.Entry) string {
	hostname, _ := os.Hostname()

	sd := "-"
	if len(entry.Data) > 0 {
		var pairs []string
		for k, v := range entry.Data {
			pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, escapeSDValue(fmt.Sprintf("%v", v))))
		}
		sd = fmt.Sprintf("[%s@0 %s]", h.tag, strings.Join(pairs, " "))
	}

	return fmt.Sprintf("1 %s %s %s %d - %s %s",
		entry.Time.Format(time.RFC3339),
		hostname,
		h.tag,
		os.Getpid(),
		sd,
		entry.Message,
	)
}

func escapeSDValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, `]`, `\]`)
	return v
}
