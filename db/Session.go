package db

import "time"

type SessionVerificationMethod int

const (
	SessionVerificationNone SessionVerificationMethod = iota
	SessionVerificationTotp
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
