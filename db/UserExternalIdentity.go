package db

import "time"

// UserExternalIdentity links a Semaphore user to an identity at an external
// auth provider. Provider is "ldap" or a key of the oidc_providers config map.
// ExternalUID is the provider's stable user ID: the OIDC "sub" claim or the
// LDAP entry DN. Matching by this pair (instead of by email) prevents account
// takeover via reused/unverified emails.
type UserExternalIdentity struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	Provider    string    `db:"provider" json:"provider"`
	ExternalUID string    `db:"external_uid" json:"external_uid"`
	Created     time.Time `db:"created" json:"created"`
}
