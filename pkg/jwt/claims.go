package jwt

import (
	"encoding/json"
	"time"
)

// Audience encodes the JWT "aud" claim. Per RFC 7519 §4.1.3 the value may be
// either a single string or a JSON array of strings. If we have only one audience,
// we encode it as a string, otherwise as an array. An empty audience is encoded as JSON null.
type Audience []string

// MarshalJSON implements json.Marshaler.
func (a Audience) MarshalJSON() ([]byte, error) {
	switch len(a) {
	case 0:
		return []byte("null"), nil
	case 1:
		return json.Marshal(a[0])
	default:
		return json.Marshal([]string(a))
	}
}

// IsZero lets `omitempty` skip an empty audience claim.
func (a Audience) IsZero() bool { return len(a) == 0 }

// TaskClaims is the JWT payload issued by Semaphore for a single task run.
type TaskClaims struct {
	// Registered claims
	Issuer    string   `json:"iss,omitempty"`
	Subject   string   `json:"sub,omitempty"`
	Audience  Audience `json:"aud,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	JWTID     string   `json:"jti,omitempty"`

	// Semaphore-specific claims
	TaskID     int  `json:"task_id"`
	ProjectID  int  `json:"project_id"`
	TemplateID int  `json:"template_id"`
	UserID     *int `json:"user_id,omitempty"`
}

// TaskInfo bundles the data needed to mint a TaskClaims set.
type TaskInfo struct {
	TaskID     int
	ProjectID  int
	TemplateID int
	UserID     *int

	Audience Audience
	TTL      time.Duration
}

// SignerOptions controls token issuance defaults.
type SignerOptions struct {
	Issuer     string
	DefaultTTL time.Duration
	MaxTTL     time.Duration
}
