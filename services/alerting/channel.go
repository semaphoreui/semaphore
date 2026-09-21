// Package alerting delivers task notifications.
//
// A Channel knows how to talk to one messenger (Telegram, Slack, ...). The
// Registry lists the channels the server supports. The Service decides which
// destinations a task must notify (server-wide config channels and/or project
// alerts), renders the message and hands it to the channel.
//
// Adding a channel means implementing Channel in one file and registering it
// in NewRegistry: the API, the UI form and the sender pick it up from there.
package alerting

import (
	"context"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// BodyFormat selects the template engine used for the message body.
type BodyFormat string

const (
	// BodyFormatText renders with text/template (JSON payloads, chat text).
	BodyFormatText BodyFormat = "text"
	// BodyFormatHTML renders with html/template (e-mail).
	BodyFormatHTML BodyFormat = "html"
)

// Field names a destination attribute a channel understands. They map 1:1 to
// the columns on db.Alert so the UI can build the form from the channel list.
type Field struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	// Secret fields are write-only: the API never returns them.
	Secret bool `json:"secret"`
}

const (
	FieldChatID     = "chat_id"
	FieldThreadID   = "thread_id"
	FieldURL        = "url"
	FieldToken      = "token"
	FieldRecipients = "recipients"
)

// Destination is where one message goes. It is built either from a project
// alert (user supplied, untrusted) or from the server config (admin supplied,
// trusted).
type Destination struct {
	ChatID     string
	ThreadID   string
	URL        string
	Token      string
	Recipients []string

	// Trusted destinations come from config.json. Only untrusted (project)
	// destinations are subject to the outbound URL policy in http.go.
	Trusted bool
}

// DestinationFromAlert copies the destination columns of a project alert.
func DestinationFromAlert(alert db.Alert) Destination {
	return Destination{
		ChatID:     strValue(alert.ChatID),
		ThreadID:   strValue(alert.ThreadID),
		URL:        strValue(alert.URL),
		Token:      strValue(alert.Token),
		Recipients: splitRecipients(strValue(alert.Recipients)),
	}
}

// Message is a rendered notification ready to be sent.
type Message struct {
	Subject string
	Body    string
	Payload Payload
}

// Channel is one messenger implementation.
type Channel interface {
	Type() db.AlertType
	Title() string
	// Icon is the Material Design icon shown by the UI.
	Icon() string
	Fields() []Field
	// DefaultEvents are used when an alert does not override its events.
	DefaultEvents() db.AlertEvents
	BodyFormat() BodyFormat
	// DefaultBody is the built-in message template.
	DefaultBody() string
	// StatusColor lets channels with colored attachments pick a color.
	StatusColor(status task_logger.TaskStatus) string

	// Validate checks a user supplied destination.
	Validate(dest Destination) error
	// InstanceDestination returns the server-wide destination configured in
	// config.json for this project, ok=false when the channel is not enabled
	// or not configured there.
	InstanceDestination(cfg *util.ConfigType, project db.Project) (dest Destination, ok bool)
	// ServerReady reports whether the server side prerequisites exist
	// (SMTP server, bot token). Channels without prerequisites return nil.
	ServerReady(cfg *util.ConfigType) error

	Send(ctx context.Context, cfg *util.ConfigType, dest Destination, msg Message) error
}

// EventForStatus maps a task status to the alert event it produces.
func EventForStatus(status task_logger.TaskStatus) (db.AlertEvent, bool) {
	switch status {
	case task_logger.TaskSuccessStatus:
		return db.AlertEventSuccess, true
	case task_logger.TaskFailStatus:
		return db.AlertEventError, true
	case task_logger.TaskWaitingConfirmation:
		return db.AlertEventWaitingConfirmation, true
	default:
		return "", false
	}
}

// AllEvents lists every event in display order.
func AllEvents() db.AlertEvents {
	return db.AlertEvents{
		db.AlertEventSuccess,
		db.AlertEventError,
		db.AlertEventWaitingConfirmation,
	}
}

func strValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func splitRecipients(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == ' '
	}) {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
