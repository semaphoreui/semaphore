// Package jwt provides JWT issuance for Semaphore task executions so that
// playbooks can authenticate to external systems (e.g. HashiCorp Vault) using
// the JWT auth method. Tokens are signed with RS256 and the corresponding
// public key is published at the /.well-known/jwks.json endpoint.
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
	TaskID       int    `json:"task_id"`
	ProjectID    int    `json:"project_id"`
	ProjectName  string `json:"project_name,omitempty"`
	TemplateID   int    `json:"template_id"`
	TemplateName string `json:"template_name,omitempty"`
	UserID       *int   `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
}

// TaskInfo bundles the data needed to mint a TaskClaims set.
type TaskInfo struct {
	TaskID       int
	ProjectID    int
	ProjectName  string
	TemplateID   int
	TemplateName string
	UserID       *int
	Username     string

	Audience Audience
	TTL      time.Duration
}

// SignerOptions controls token issuance defaults.
type SignerOptions struct {
	Issuer   string
	Audience Audience
	TTL      time.Duration
	MaxTTL   time.Duration
}
