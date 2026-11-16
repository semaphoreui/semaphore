package db

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"
)

// APITokenType says where a user token is accepted and on which routes it is
// listed, issued and revoked.
type APITokenType string

const (
	// APITokenTypeAPI is accepted by the REST API.
	APITokenTypeAPI APITokenType = "api"
	// APITokenTypeMCP is accepted by the MCP endpoint and the REST API: MCP
	// tools call the REST API with the token of the MCP request.
	APITokenTypeMCP APITokenType = "mcp"
)

// APIToken is given to a user to allow API access
type APIToken struct {
	ID        string     `db:"id" json:"id"`
	Created   time.Time  `db:"created" json:"created"`
	Expired   bool       `db:"expired" json:"expired"`
	ExpiresAt *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	UserID    int        `db:"user_id" json:"user_id"`
	Name      string     `db:"name" json:"name"`
	// Type is read from the row; each token route serves one type, so it is
	// not sent to clients.
	Type APITokenType `db:"type" json:"-"`
}

// NewUserTokenID returns a new random value for a user token of any type.
// The value is the token itself and the primary key of its row in user__token:
// 32 random bytes in lower-case URL-safe base64, 44 characters.
func NewUserTokenID() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return strings.ToLower(base64.URLEncoding.EncodeToString(raw)), nil
}

// IsExpiredAt reports whether the token is revoked or past its expiry.
func (t APIToken) IsExpiredAt(now time.Time) bool {
	if t.Expired {
		return true
	}
	if t.ExpiresAt != nil && !t.ExpiresAt.After(now) {
		return true
	}
	return false
}
