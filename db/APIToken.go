package db

import "time"

// APIToken is given to a user to allow API access
type APIToken struct {
	ID        string     `db:"id" json:"id"`
	Created   time.Time  `db:"created" json:"created"`
	Expired   bool       `db:"expired" json:"expired"`
	ExpiresAt *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	UserID    int        `db:"user_id" json:"user_id"`
	Name      string     `db:"name" json:"name"`
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
