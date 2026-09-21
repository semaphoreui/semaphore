package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// AlertSnapshot is the resolved alerting decision stored on a task at creation
// time so every HA node sends the same alerts even if bindings change later.
type AlertSnapshot struct {
	AlertIDs  []int `json:"alert_ids"`
	OnSuccess bool  `json:"on_success"`
	OnError   bool  `json:"on_error"`
}

func (s *AlertSnapshot) Scan(value any) error {
	if value == nil {
		*s = AlertSnapshot{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*s = AlertSnapshot{}
			return nil
		}
		return json.Unmarshal(v, s)
	case string:
		if v == "" {
			*s = AlertSnapshot{}
			return nil
		}
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported type for AlertSnapshot")
	}
}

func (s *AlertSnapshot) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
