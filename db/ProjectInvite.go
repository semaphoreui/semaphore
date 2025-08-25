package db

import (
	"time"
)

type ProjectInviteStatus string

const (
	ProjectInvitePending  ProjectInviteStatus = "pending"
	ProjectInviteAccepted ProjectInviteStatus = "accepted"
	ProjectInviteDeclined ProjectInviteStatus = "declined"
	ProjectInviteExpired  ProjectInviteStatus = "expired"
)

func (s ProjectInviteStatus) IsValid() bool {
	switch s {
	case ProjectInvitePending, ProjectInviteAccepted, ProjectInviteDeclined, ProjectInviteExpired:
		return true
	default:
		return false
	}
}

type ProjectInvite struct {
	ID            int                 `db:"id" json:"id" backup:"-"`
	ProjectID     int                 `db:"project_id" json:"project_id"`
	UserID        *int                `db:"user_id" json:"user_id,omitempty"` // Can be null for email invites
	Email         *string             `db:"email" json:"email,omitempty"`     // For email-based invites
	Role          ProjectUserRole     `db:"role" json:"role"`
	Status        ProjectInviteStatus `db:"status" json:"status"`
	Token         string              `db:"token" json:"-"`                         // Secret token for accepting invite
	InviterUserID int                 `db:"inviter_user_id" json:"inviter_user_id"` // User who created the invite
	Created       time.Time           `db:"created" json:"created" backup:"-"`
	ExpiresAt     *time.Time          `db:"expires_at" json:"expires_at,omitempty"`
	AcceptedAt    *time.Time          `db:"accepted_at" json:"accepted_at,omitempty"`
}

type ProjectInviteWithUser struct {
	ProjectInvite
	InvitedByUser *User `json:"inviter_user,omitempty"`
	User          *User `json:"user,omitempty"`
}
