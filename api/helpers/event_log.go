package helpers

import (
	"net"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	log "github.com/sirupsen/logrus"
)

type EventLogItem struct {
	IntegrationID int
	UserID        int
	ProjectID     int

	ObjectType  db.EventObjectType
	ObjectID    int
	Description string
}

type EventLogType string

const (
	EventLogCreate EventLogType = "create"
	EventLogUpdate EventLogType = "update"
	EventLogDelete EventLogType = "delete"

	EventLogLoginSuccess EventLogType = "login_success"
	EventLogLoginFail    EventLogType = "login_fail"
	EventLogLogout       EventLogType = "logout"
)

// extractClientIP returns the client address, preferring the reverse-proxy
// header used elsewhere in the codebase (see createSession in api/login.go).
func extractClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// sanitizeLogValue strips CR/LF so a value cannot inject extra lines into
// newline-delimited log files (OWASP log injection).
func sanitizeLogValue(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}

func EventLog(r *http.Request, action EventLogType, item EventLogItem) {
	actionStr := string(action)
	ip := extractClientIP(r)
	userAgent := r.Header.Get("user-agent")
	description := sanitizeLogValue(item.Description)

	event := db.Event{
		ObjectType:  &item.ObjectType,
		ObjectID:    &item.ObjectID,
		Description: &description,
		Action:      &actionStr,
		IP:          &ip,
		UserAgent:   &userAgent,
	}

	if item.IntegrationID > 0 {
		event.IntegrationID = &item.IntegrationID
	}

	if item.UserID > 0 {
		event.UserID = &item.UserID
	}

	if item.ProjectID > 0 {
		event.ProjectID = &item.ProjectID
	}

	logFields := event.ToFields()
	logFields["action"] = string(action)

	if _, err := Store(r).CreateEvent(event); err != nil {
		log.WithFields(logFields).Error("Failed to store event")
	}

	logWriterVal, ok := GetOkFromContext(r, "log_writer")
	if !ok {
		return
	}
	logWriter := logWriterVal.(pro_interfaces.LogWriteService)

	objectType := string(item.ObjectType)

	if err := logWriter.WriteEventLog(pro_interfaces.EventLogRecord{
		Action:        string(action),
		ProjectID:     event.ProjectID,
		UserID:        event.UserID,
		IntegrationID: event.IntegrationID,
		ObjectType:    &objectType,
		ObjectID:      event.ObjectID,
		Description:   event.Description,
		IP:            ip,
		UserAgent:     userAgent,
	}); err != nil {
		log.WithFields(logFields).Error("Failed to store event in log file")
	}
}
