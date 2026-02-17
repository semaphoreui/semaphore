//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"log/syslog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

var localSyslogPaths = []string{"/dev/log", "/var/run/syslog", "/var/run/log"}

func initSyslog(conf *util.SyslogConfig) {
	if !conf.Enabled {
		return
	}

	switch conf.Format {
	case util.SyslogRFC5424:
		hook, err := newRFC5424Hook(conf.Network, conf.Address, conf.Tag)
		if err != nil {
			log.WithError(err).Fatal("Failed to create syslog hook")
			return
		}
		log.AddHook(hook)
		log.Info("Syslog logging enabled (RFC 5424)")
	default:
		hook, err := lSyslog.NewSyslogHook(conf.Network, conf.Address, syslog.LOG_DEBUG, conf.Tag)
		if err != nil {
			log.WithError(err).Fatal("Failed to create syslog hook")
			return
		}
		log.AddHook(hook)
		log.Info("Syslog logging enabled")
	}
}

type rfc5424Hook struct {
	conn     net.Conn
	tag      string
	hostname string
	mu       sync.Mutex
}

func newRFC5424Hook(network, address, tag string) (*rfc5424Hook, error) {
	var conn net.Conn
	var err error

	if network != "" && address != "" {
		conn, err = net.Dial(network, address)
	} else {
		for _, path := range localSyslogPaths {
			conn, err = net.Dial("unixgram", path)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()

	return &rfc5424Hook{
		conn:     conn,
		tag:      tag,
		hostname: hostname,
	}, nil
}

var levelToSeverity = map[log.Level]syslog.Priority{
	log.PanicLevel: syslog.LOG_CRIT,
	log.FatalLevel: syslog.LOG_CRIT,
	log.ErrorLevel: syslog.LOG_ERR,
	log.WarnLevel:  syslog.LOG_WARNING,
	log.InfoLevel:  syslog.LOG_INFO,
	log.DebugLevel: syslog.LOG_DEBUG,
	log.TraceLevel: syslog.LOG_DEBUG,
}

func (h *rfc5424Hook) Levels() []log.Level {
	return log.AllLevels
}

func (h *rfc5424Hook) Fire(entry *log.Entry) error {
	severity, ok := levelToSeverity[entry.Level]
	if !ok {
		severity = syslog.LOG_INFO
	}
	pri := syslog.LOG_USER | severity

	sd := "-"
	if len(entry.Data) > 0 {
		var pairs []string
		for k, v := range entry.Data {
			pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, escapeSDValue(fmt.Sprintf("%v", v))))
		}
		sd = fmt.Sprintf("[%s@0 %s]", h.tag, strings.Join(pairs, " "))
	}

	// RFC 5424: <PRI>VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID SP STRUCTURED-DATA [SP MSG]
	msg := fmt.Sprintf("<%d>1 %s %s %s %d - %s %s",
		pri,
		entry.Time.Format(time.RFC3339),
		h.hostname,
		h.tag,
		os.Getpid(),
		sd,
		entry.Message,
	)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := fmt.Fprintln(h.conn, msg)
	return err
}

func escapeSDValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, `]`, `\]`)
	return v
}
