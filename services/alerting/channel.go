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
	// BodyFormatJSON renders with text/template after JSON-escaping dynamic
	// string fields in the payload.
	BodyFormatJSON BodyFormat = "json"
	// BodyFormatHTML renders with html/template (e-mail).
	BodyFormatHTML BodyFormat = "html"
)

// FieldKind tells the UI which input to render for a field.
type FieldKind string

const (
	FieldKindText   FieldKind = "text"
	FieldKindNumber FieldKind = "number"
	FieldKindBool   FieldKind = "bool"
)

// Field names a destination attribute a channel understands. Column fields
// (chat_id, thread_id, url, recipients) live on db.Alert; every other name
// is stored in db.Alert.Params.
type Field struct {
	Name     string    `json:"name"`
	Kind     FieldKind `json:"kind"`
	Required bool      `json:"required"`
	// Override fields belong to the "use my own server" group: they are only
	// meaningful when the alert brings its own secret instead of the
	// server-wide one (an e-mail alert with its own SMTP server).
	Override bool `json:"override"`
}

const (
	FieldChatID     = "chat_id"
	FieldThreadID   = "thread_id"
	FieldURL        = "url"
	FieldRecipients = "recipients"

	FieldSMTPHost   = "smtp_host"
	FieldSMTPPort   = "smtp_port"
	FieldSMTPSender = "smtp_sender"
	FieldSMTPSecure = "smtp_secure"
	FieldSMTPTLS    = "smtp_tls"
)

// SecretSpec describes the secret a channel needs and where it may come from.
type SecretSpec struct {
	// KeyType is the access key type that carries the secret: string for a
	// single token, login_password for credentials.
	KeyType db.AccessKeyType `json:"key_type"`
	// Label names the secret for the UI ("Bot token", "SMTP credentials").
	Label string `json:"label"`
	// Optional secrets may be left out even when overriding the server
	// settings (an SMTP server without authentication).
	Optional bool `json:"optional"`
}

// Destination is where one message goes. It is built either from a project
// alert (user supplied, untrusted) or from the server config (admin supplied,
// trusted).
type Destination struct {
	ChatID     string
	ThreadID   string
	URL        string
	Recipients []string

	// Params carries the override fields declared by the channel.
	Params map[string]any

	// Secret is the decrypted access key the alert points at, nil when the
	// server-wide secret applies.
	Secret *db.AccessKey

	// Trusted destinations come from config.json. Only untrusted (project)
	// destinations are subject to the outbound URL policy in http.go.
	Trusted bool
}

// DestinationFromAlert copies the destination columns of a project alert.
// The secret is attached separately by the service once decrypted.
func DestinationFromAlert(alert db.Alert) Destination {
	return Destination{
		ChatID:     strValue(alert.ChatID),
		ThreadID:   strValue(alert.ThreadID),
		URL:        strValue(alert.URL),
		Recipients: splitRecipients(strValue(alert.Recipients)),
		Params:     alert.Params,
	}
}

// Param returns a params value as a trimmed string.
func (d Destination) Param(name string) string {
	return db.Alert{Params: d.Params}.ParamString(name)
}

// ParamBool returns a params value as a bool.
func (d Destination) ParamBool(name string) bool {
	return db.Alert{Params: d.Params}.ParamBool(name)
}

// Value returns a field by name, whether it is a column or a param.
func (d Destination) Value(name string) string {
	switch name {
	case FieldChatID:
		return d.ChatID
	case FieldThreadID:
		return d.ThreadID
	case FieldURL:
		return d.URL
	case FieldRecipients:
		return strings.Join(d.Recipients, ",")
	default:
		return d.Param(name)
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
	// Secret describes the secret the channel needs, nil when it needs none.
	Secret() *SecretSpec

	// Validate checks a user supplied destination, including whether the
	// server-wide secret may be used when the alert brings none.
	Validate(cfg *util.ConfigType, dest Destination) error
	// InstanceDestination returns the server-wide destination configured in
	// config.json for this project, ok=false when the channel is not enabled
	// or not configured there.
	InstanceDestination(cfg *util.ConfigType, project db.Project) (dest Destination, ok bool)
	// ServerEnabled reports whether the server-wide toggle enables this channel
	// at all, even when more project-specific data is still required.
	ServerEnabled(cfg *util.ConfigType) bool
	// ServerReady reports whether the server-wide secret exists (SMTP
	// server, bot token, Gotify pair). Channels without a secret return nil.
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
