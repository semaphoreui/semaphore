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
	// OnSuccess / OnError mirror the template's suppress_* flags.
	OnSuccess bool `json:"on_success"`
	OnError   bool `json:"on_error"`
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
	return !s.Instance && len(s.AlertIDs) == 0
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
