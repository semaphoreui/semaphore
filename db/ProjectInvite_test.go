package db

import (
	"testing"
	"time"
)

func TestProjectInviteStatus_IsValid(t *testing.T) {
	tests := []struct {
		status ProjectInviteStatus
		valid  bool
	}{
		{ProjectInvitePending, true},
		{ProjectInviteAccepted, true},
		{ProjectInviteDeclined, true},
		{ProjectInviteExpired, true},
		{ProjectInviteStatus("invalid"), false},
		{ProjectInviteStatus(""), false},
	}

	for _, test := range tests {
		if test.status.IsValid() != test.valid {
			t.Errorf("Status %q: expected valid=%v, got %v", test.status, test.valid, test.status.IsValid())
		}
	}
}

func TestProjectInvite_EmailBasedInvite(t *testing.T) {
	email := "test@example.com"
	invite := ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         &email,
		Role:          ProjectManager,
		Status:        ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
	}

	if invite.UserID != nil {
		t.Error("Email-based invite should not have UserID")
	}

	if invite.Email == nil || *invite.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", invite.Email)
	}

	if invite.Status != ProjectInvitePending {
		t.Errorf("Expected status 'pending', got %s", invite.Status)
	}
}

func TestProjectInvite_UserBasedInvite(t *testing.T) {
	userID := 42
	invite := ProjectInvite{
		ID:            1,
		ProjectID:     1,
		UserID:        &userID,
		Role:          ProjectTaskRunner,
		Status:        ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
	}

	if invite.Email != nil {
		t.Error("User-based invite should not have Email")
	}

	if invite.UserID == nil || *invite.UserID != 42 {
		t.Errorf("Expected user_id 42, got %v", invite.UserID)
	}

	if invite.Role != ProjectTaskRunner {
		t.Errorf("Expected role 'task_runner', got %s", invite.Role)
	}
}

func TestProjectInvite_WithExpiration(t *testing.T) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	email := "test@example.com"

	invite := ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         &email,
		Role:          ProjectManager,
		Status:        ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
		ExpiresAt:     &expiresAt,
	}

	if invite.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}

	if invite.AcceptedAt != nil {
		t.Error("AcceptedAt should be nil for pending invite")
	}
}

func TestProjectInvite_AcceptedInvite(t *testing.T) {
	acceptedAt := time.Now()
	email := "test@example.com"

	invite := ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         &email,
		Role:          ProjectManager,
		Status:        ProjectInviteAccepted,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now().Add(-1 * time.Hour),
		AcceptedAt:    &acceptedAt,
	}

	if invite.Status != ProjectInviteAccepted {
		t.Errorf("Expected status 'accepted', got %s", invite.Status)
	}

	if invite.AcceptedAt == nil {
		t.Error("AcceptedAt should be set for accepted invite")
	}
}

func TestProjectInviteWithUser_Structure(t *testing.T) {
	email := "test@example.com"
	invite := ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         &email,
		Role:          ProjectManager,
		Status:        ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
	}

	invitedByUser := User{
		ID:       1,
		Username: "admin",
		Email:    "admin@example.com",
		Name:     "Administrator",
	}

	inviteWithUser := ProjectInviteWithUser{
		ProjectInvite: invite,
		InvitedByUser: &invitedByUser,
		User:          nil, // Email-based invite
	}

	if inviteWithUser.ProjectInvite.ID != invite.ID {
		t.Error("ProjectInvite should be embedded correctly")
	}

	if inviteWithUser.InvitedByUser == nil {
		t.Error("InvitedByUser should be set")
	}

	if inviteWithUser.InvitedByUser.Username != "admin" {
		t.Errorf("Expected inviter username 'admin', got %s", inviteWithUser.InvitedByUser.Username)
	}

	if inviteWithUser.User != nil {
		t.Error("User should be nil for email-based invite")
	}
}
