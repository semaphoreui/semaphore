package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// AlertType identifies a delivery channel (telegram, slack, email, ...).
// The set of supported values is owned by the alerting channel registry,
// the db layer only stores the string.
type AlertType string

// AlertEvent is a task life-cycle moment that may produce a notification.
type AlertEvent string

const (
	AlertEventSuccess             AlertEvent = "success"
	AlertEventError               AlertEvent = "error"
	AlertEventWaitingConfirmation AlertEvent = "waiting_confirmation"
)

// Alert modes decide where a template or a schedule sends notifications.
const (
	// AlertModeDefault (templates): server channels enabled for the project
	// plus project alerts marked as default.
	AlertModeDefault = "default"
	// AlertModeInherit (schedules): use whatever the template resolves to.
	AlertModeInherit = "inherit"
	// AlertModeIDs (both): only the explicitly linked alerts. Empty means silent.
	AlertModeIDs = "ids"
)

// AlertEvents is the list of events an alert subscribes to. It is stored
// as a comma-separated string and exposed as a JSON array. An empty list
// means "use the channel defaults".
type AlertEvents []AlertEvent

func (e *AlertEvents) Scan(value any) error {
	if value == nil {
		*e = nil
		return nil
	}
	var raw string
	switch v := value.(type) {
	case []byte:
		raw = string(v)
	case string:
		raw = v
	default:
		return errors.New("unsupported type for AlertEvents")
	}
	*e = ParseAlertEvents(raw)
	return nil
}

func (e AlertEvents) Value() (driver.Value, error) {
	if len(e) == 0 {
		return nil, nil
	}
	return e.String(), nil
}

func (e AlertEvents) String() string {
	parts := make([]string, 0, len(e))
	for _, ev := range e {
		parts = append(parts, string(ev))
	}
	return strings.Join(parts, ",")
}

func (e AlertEvents) Contains(event AlertEvent) bool {
	for _, ev := range e {
		if ev == event {
			return true
		}
	}
	return false
}

// ParseAlertEvents parses the comma-separated storage format, dropping blanks
// and duplicates while keeping order.
func ParseAlertEvents(raw string) AlertEvents {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	seen := make(map[AlertEvent]bool)
	var out AlertEvents
	for _, part := range strings.Split(raw, ",") {
		ev := AlertEvent(strings.TrimSpace(part))
		if ev == "" || seen[ev] {
			continue
		}
		seen[ev] = true
		out = append(out, ev)
	}
	return out
}

// Alert is a named notification destination that belongs to a project:
// a channel type, the destination fields the channel needs, the events it
// listens to and an optional custom message template.
type Alert struct {
	ID        int       `db:"id" json:"id" backup:"-"`
	ProjectID int       `db:"project_id" json:"project_id" backup:"-"`
	Name      string    `db:"name" json:"name"`
	Type      AlertType `db:"type" json:"type"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	// IsDefault marks the alert as one of the project defaults: templates in
	// AlertModeDefault send it without linking it explicitly.
	IsDefault bool `db:"is_default" json:"is_default"`

	// Events overrides the channel default event list when not empty.
	Events AlertEvents `db:"events" json:"events"`

	// Destination fields. Which of them apply is declared by the channel.
	ChatID     *string `db:"chat_id" json:"chat_id,omitempty"`
	ThreadID   *string `db:"thread_id" json:"thread_id,omitempty"`
	URL        *string `db:"url" json:"url,omitempty"`
	Token      *string `db:"token" json:"token,omitempty" backup:"-"`
	Recipients *string `db:"recipients" json:"recipients,omitempty"`

	// Body is a custom Go text/template (html/template for email). Empty
	// means the channel's built-in template.
	Body *string `db:"body" json:"body,omitempty"`
}

func (a Alert) GetID() int {
	return a.ID
}

func (a Alert) GetName() string {
	return a.Name
}

// MarshalJSON never echoes the token: it is write-only through the API.
// Backup and export use reflection over db/backup tags, not this method.
func (a Alert) MarshalJSON() ([]byte, error) {
	type alertJSON Alert
	out := alertJSON(a)
	out.Token = nil
	if out.Events == nil {
		out.Events = AlertEvents{}
	}
	return json.Marshal(out)
}

// Validate checks the channel-agnostic invariants. Channel-specific rules
// (required destination fields, URL policy) live in the alerting registry.
func (a *Alert) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return common_errors.NewValidationError("alert name can not be empty")
	}
	if strings.TrimSpace(string(a.Type)) == "" {
		return common_errors.NewValidationError("alert type can not be empty")
	}
	for _, ev := range a.Events {
		switch ev {
		case AlertEventSuccess, AlertEventError, AlertEventWaitingConfirmation:
		default:
			return common_errors.NewValidationError("unknown alert event: " + string(ev))
		}
	}
	return nil
}

// Normalize trims strings and turns blank optional values into NULL so the
// UI's empty inputs never end up stored as "".
func (a *Alert) Normalize() {
	a.Name = strings.TrimSpace(a.Name)
	a.Type = AlertType(strings.ToLower(strings.TrimSpace(string(a.Type))))
	a.ChatID = trimToNil(a.ChatID)
	a.ThreadID = trimToNil(a.ThreadID)
	a.URL = trimToNil(a.URL)
	a.Token = trimToNil(a.Token)
	a.Recipients = trimToNil(a.Recipients)
	a.Body = trimToNil(a.Body)
}

func trimToNil(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// ValidateAlertMode rejects modes outside the allowed set for the owner
// (templates accept default|ids, schedules accept inherit|ids).
func ValidateAlertMode(mode string, allowed ...string) error {
	for _, a := range allowed {
		if mode == a {
			return nil
		}
	}
	return common_errors.NewValidationError("alert_mode must be one of: " + strings.Join(allowed, ", "))
}
