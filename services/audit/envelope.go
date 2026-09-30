package audit

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	SchemaVersion     = "1"
	TimestampLayout   = "2006-01-02T15:04:05.000000Z"
	MaxNameBytes      = 128
	MaxLoginNameBytes = 64
	MaxMetadataBytes  = 8192

	// Sizes of the audit_event columns.
	maxIPBytes        = 64
	maxUserAgentBytes = 1024
	maxNodeIDBytes    = 255
)

type Envelope struct {
	EventID       string          `json:"event_id"`
	Seq           int64           `json:"seq"`
	Timestamp     string          `json:"timestamp"`
	SchemaVersion string          `json:"schema_version"`
	Category      string          `json:"category"`
	EventCode     string          `json:"event_code"`
	Type          Type            `json:"type"`
	Action        string          `json:"action"`
	Outcome       Outcome         `json:"outcome"`
	Reason        Reason          `json:"reason"`
	Actor         Actor           `json:"actor"`
	Source        *Source         `json:"source,omitempty"`
	Target        *Target         `json:"target,omitempty"`
	Scope         *Scope          `json:"scope,omitempty"`
	RequestID     string          `json:"request_id,omitempty"`
	InstanceID    string          `json:"instance_id"`
	NodeID        string          `json:"node_id,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
}

type Actor struct {
	Type             ActorType  `json:"type"`
	ID               string     `json:"id,omitempty"`
	Name             string     `json:"name,omitempty"`
	Auth             AuthMethod `json:"auth,omitempty"`
	TokenFingerprint string     `json:"token_fingerprint,omitempty"`
}

type Source struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent,omitempty"`
}

type Target struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Scope struct {
	ProjectID string `json:"project_id"`
}

func UserTarget(id int, name string) *Target {
	return &Target{Type: TargetUser, ID: strconv.Itoa(id), Name: name}
}

// PostgreSQL rejects NUL, so control characters are dropped.
func TruncateName(s string, max int) string {
	s = strings.ToValidUTF8(s, "�")
	s = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' {
			return -1
		}
		return r
	}, s)
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
