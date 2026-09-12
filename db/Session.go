package db

import "time"

// SessionInactivityTimeout is how long a session may stay unused before it is
// rejected. It applies regardless of the configured absolute session lifetime.
const SessionInactivityTimeout = 7 * 24 * time.Hour

type SessionVerificationMethod int

const (
	SessionVerificationNone SessionVerificationMethod = iota
	SessionVerificationTotp
	SessionVerificationEmail
)

// Session is a connection to the API
type Session struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	Created    time.Time `db:"created" json:"created"`
	LastActive time.Time `db:"last_active" json:"last_active"`
	IP         string    `db:"ip" json:"ip"`
	UserAgent  string    `db:"user_agent" json:"user_agent"`
	Expired    bool      `db:"expired" json:"expired"`

	VerificationMethod SessionVerificationMethod `db:"verification_method" json:"verification_method"`
	Verified           bool                      `db:"verified" json:"verified"`
}

func (s *Session) IsVerified() bool {
	if s.VerificationMethod == SessionVerificationNone {
		return true
	}
	return s.Verified
}

// IsExpiredAt reports whether the session must no longer be accepted at the
// given moment. A session is expired when it was revoked, when it has been
// unused for longer than maxInactivity, or, if maxLife is positive, when more
// than maxLife has passed since it was created. maxLife <= 0 disables the
// absolute lifetime limit.
func (s *Session) IsExpiredAt(now time.Time, maxLife time.Duration, maxInactivity time.Duration) bool {
	if s.Expired {
		return true
	}

	if now.Sub(s.LastActive) > maxInactivity {
		return true
	}

	if maxLife > 0 && now.Sub(s.Created) > maxLife {
		return true
	}

	return false
}
