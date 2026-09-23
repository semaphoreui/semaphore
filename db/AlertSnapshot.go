package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// AlertSnapshot is the alerting decision frozen on a task when it is created.
// Every HA node that later reports a status change reads the same snapshot,
// so edits to templates, schedules or alerts made while a task runs never
// change what is sent for it.
type AlertSnapshot struct {
	// Instance is true when the task should also go to the channels
	// configured server-wide in config.json (legacy "allow alerts for this
	// project" behaviour).
	Instance bool `json:"instance"`
	// AlertIDs are the project alerts to send.
	AlertIDs []int `json:"alert_ids"`
	// Deliveries is the effective delivery state frozen when the task was
	// created. Older tasks may still only carry Instance/AlertIDs.
	Deliveries []AlertDeliverySnapshot `json:"deliveries,omitempty"`
	// OnSuccess / OnError mirror the template's suppress_* flags.
	OnSuccess bool `json:"on_success"`
	OnError   bool `json:"on_error"`
}

type AlertDeliverySnapshot struct {
	Key        string            `json:"key"`
	Name       string            `json:"name"`
	Type       AlertType         `json:"type"`
	Events     AlertEvents       `json:"events,omitempty"`
	ChatID     string            `json:"chat_id,omitempty"`
	ThreadID   string            `json:"thread_id,omitempty"`
	URL        string            `json:"url,omitempty"`
	Recipients []string          `json:"recipients,omitempty"`
	KeyID      *int              `json:"key_id,omitempty"`
	Params     MapStringAnyField `json:"params,omitempty"`
	Body       string            `json:"body,omitempty"`
	Trusted    bool              `json:"trusted,omitempty"`
}

// Allows reports whether the template-level flags let the event through.
// Events without a template flag (waiting_confirmation) are always allowed.
func (s AlertSnapshot) Allows(event AlertEvent) bool {
	switch event {
	case AlertEventSuccess:
		return s.OnSuccess
	case AlertEventError:
		return s.OnError
	default:
		return true
	}
}

func (s AlertSnapshot) IsEmpty() bool {
	return len(s.Deliveries) == 0 && !s.Instance && len(s.AlertIDs) == 0
}

func (s *AlertSnapshot) Scan(value any) error {
	if value == nil {
		*s = AlertSnapshot{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return errors.New("unsupported type for AlertSnapshot")
	}
	if len(raw) == 0 {
		*s = AlertSnapshot{}
		return nil
	}
	return json.Unmarshal(raw, s)
}

func (s AlertSnapshot) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
