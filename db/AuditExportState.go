package db

import "time"

type AuditExportState struct {
	DestinationID   string     `db:"destination_id" json:"destination_id"`
	CursorSeq       int64      `db:"cursor_seq" json:"cursor_seq"`
	OwnerID         *string    `db:"owner_id" json:"owner_id,omitempty"`
	LeaseUntil      *time.Time `db:"lease_until" json:"lease_until,omitempty"`
	LeaseGeneration int64      `db:"lease_generation" json:"lease_generation"`
	LastAttemptAt   *time.Time `db:"last_attempt_at" json:"last_attempt_at,omitempty"`
	LastSuccessAt   *time.Time `db:"last_success_at" json:"last_success_at,omitempty"`
	LastError       *string    `db:"last_error" json:"last_error,omitempty"`
}
