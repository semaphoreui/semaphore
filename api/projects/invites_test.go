package projects

import (
	"fmt"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
)

// Mock store for testing
type mockInviteStore struct {
	projects       map[int]db.Project
	users          map[int]db.User
	projectUsers   map[string]db.ProjectUser
	invites        map[int]db.ProjectInvite
	invitesByToken map[string]db.ProjectInvite
	nextInviteID   int
}

func newMockInviteStore() *mockInviteStore {
	return &mockInviteStore{
		projects:       make(map[int]db.Project),
		users:          make(map[int]db.User),
		projectUsers:   make(map[string]db.ProjectUser),
		invites:        make(map[int]db.ProjectInvite),
		invitesByToken: make(map[string]db.ProjectInvite),
		nextInviteID:   1,
	}
}

func (m *mockInviteStore) GetProject(projectID int) (db.Project, error) {
	if project, exists := m.projects[projectID]; exists {
		return project, nil
	}
	return db.Project{}, db.ErrNotFound
}

func (m *mockInviteStore) GetUser(userID int) (db.User, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return db.User{}, db.ErrNotFound
}

func (m *mockInviteStore) GetProjectUser(projectID, userID int) (db.ProjectUser, error) {
	key := fmt.Sprintf("%d-%d", projectID, userID)
	if projectUser, exists := m.projectUsers[key]; exists {
		return projectUser, nil
	}
	return db.ProjectUser{}, db.ErrNotFound
}

func (m *mockInviteStore) CreateProjectUser(projectUser db.ProjectUser) (db.ProjectUser, error) {
	key := fmt.Sprintf("%d-%d", projectUser.ProjectID, projectUser.UserID)
	m.projectUsers[key] = projectUser
	return projectUser, nil
}

func (m *mockInviteStore) GetProjectInvites(projectID int, params db.RetrieveQueryParams) ([]db.ProjectInviteWithUser, error) {
	var invites []db.ProjectInviteWithUser
	for _, invite := range m.invites {
		if invite.ProjectID == projectID {
			inviteWithUser := db.ProjectInviteWithUser{
				ProjectInvite: invite,
			}
			if invitedByUser, exists := m.users[invite.InviterUserID]; exists {
				inviteWithUser.InvitedByUser = &invitedByUser
			}
			if invite.UserID != nil {
				if user, exists := m.users[*invite.UserID]; exists {
					inviteWithUser.User = &user
				}
			}
			invites = append(invites, inviteWithUser)
		}
	}
	return invites, nil
}

func (m *mockInviteStore) CreateProjectInvite(invite db.ProjectInvite) (db.ProjectInvite, error) {
	invite.ID = m.nextInviteID
	m.nextInviteID++
	m.invites[invite.ID] = invite
	m.invitesByToken[invite.Token] = invite
	return invite, nil
}

func (m *mockInviteStore) GetProjectInvite(projectID, inviteID int) (db.ProjectInvite, error) {
	if invite, exists := m.invites[inviteID]; exists && invite.ProjectID == projectID {
		return invite, nil
	}
	return db.ProjectInvite{}, db.ErrNotFound
}

func (m *mockInviteStore) GetProjectInviteByToken(token string) (db.ProjectInvite, error) {
	if invite, exists := m.invitesByToken[token]; exists {
		return invite, nil
	}
	return db.ProjectInvite{}, db.ErrNotFound
}

func (m *mockInviteStore) UpdateProjectInvite(invite db.ProjectInvite) error {
	if _, exists := m.invites[invite.ID]; exists {
		m.invites[invite.ID] = invite
		m.invitesByToken[invite.Token] = invite
		return nil
	}
	return db.ErrNotFound
}

func (m *mockInviteStore) DeleteProjectInvite(projectID, inviteID int) error {
	if invite, exists := m.invites[inviteID]; exists && invite.ProjectID == projectID {
		delete(m.invites, inviteID)
		delete(m.invitesByToken, invite.Token)
		return nil
	}
	return db.ErrNotFound
}

// Test database repository functionality

func TestMockStore_GetProjectInvites(t *testing.T) {
	store := newMockInviteStore()

	// Add test invites
	invite1 := db.ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         stringPtr("user1@example.com"),
		Role:          db.ProjectManager,
		Status:        db.ProjectInvitePending,
		Token:         "token1",
		InviterUserID: 1,
		Created:       time.Now(),
	}
	store.CreateProjectInvite(invite1)

	invite2 := db.ProjectInvite{
		ID:            2,
		ProjectID:     1,
		UserID:        intPtr(2),
		Role:          db.ProjectTaskRunner,
		Status:        db.ProjectInvitePending,
		Token:         "token2",
		InviterUserID: 1,
		Created:       time.Now(),
	}
	store.users[2] = db.User{ID: 2, Username: "user2", Email: "user2@example.com"}
	store.CreateProjectInvite(invite2)

	// Test getting invites
	invites, err := store.GetProjectInvites(1, db.RetrieveQueryParams{})
	if err != nil {
		t.Errorf("Failed to get invites: %v", err)
	}

	if len(invites) != 2 {
		t.Errorf("Expected 2 invites, got %d", len(invites))
	}

	// Find email-based invite
	var emailInvite *db.ProjectInviteWithUser
	var userInvite *db.ProjectInviteWithUser

	for i := range invites {
		if invites[i].Email != nil {
			emailInvite = &invites[i]
		}
		if invites[i].UserID != nil {
			userInvite = &invites[i]
		}
	}

	// Verify email invite
	if emailInvite == nil {
		t.Error("Expected to find email-based invite")
	} else if *emailInvite.Email != "user1@example.com" {
		t.Errorf("Expected email invite 'user1@example.com', got %v", *emailInvite.Email)
	}

	// Verify user invite
	if userInvite == nil {
		t.Error("Expected to find user-based invite")
	} else if *userInvite.UserID != 2 {
		t.Errorf("Expected user invite user_id 2, got %v", *userInvite.UserID)
	}
}

func TestMockStore_CreateProjectInvite(t *testing.T) {
	store := newMockInviteStore()

	invite := db.ProjectInvite{
		ProjectID:     1,
		Email:         stringPtr("newuser@example.com"),
		Role:          db.ProjectManager,
		Status:        db.ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
	}

	createdInvite, err := store.CreateProjectInvite(invite)
	if err != nil {
		t.Errorf("Failed to create invite: %v", err)
	}

	if createdInvite.ID == 0 {
		t.Error("Expected invite ID to be set")
	}

	if createdInvite.Email == nil || *createdInvite.Email != "newuser@example.com" {
		t.Errorf("Expected email 'newuser@example.com', got %v", createdInvite.Email)
	}

	if createdInvite.Role != db.ProjectManager {
		t.Errorf("Expected role 'manager', got %s", createdInvite.Role)
	}

	if createdInvite.Status != db.ProjectInvitePending {
		t.Errorf("Expected status 'pending', got %s", createdInvite.Status)
	}

	// Verify it can be retrieved by token
	retrievedInvite, err := store.GetProjectInviteByToken("test-token")
	if err != nil {
		t.Errorf("Failed to get invite by token: %v", err)
	}

	if retrievedInvite.ID != createdInvite.ID {
		t.Errorf("Expected invite ID %d, got %d", createdInvite.ID, retrievedInvite.ID)
	}
}

func TestMockStore_UpdateProjectInvite(t *testing.T) {
	store := newMockInviteStore()

	// Create test invite
	invite := db.ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         stringPtr("test@example.com"),
		Role:          db.ProjectManager,
		Status:        db.ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
	}
	store.CreateProjectInvite(invite)

	// Update invite status
	invite.Status = db.ProjectInviteAccepted
	now := time.Now()
	invite.AcceptedAt = &now

	err := store.UpdateProjectInvite(invite)
	if err != nil {
		t.Errorf("Failed to update invite: %v", err)
	}

	// Verify update
	updatedInvite := store.invites[1]
	if updatedInvite.Status != db.ProjectInviteAccepted {
		t.Errorf("Expected status 'accepted', got %s", updatedInvite.Status)
	}

	if updatedInvite.AcceptedAt == nil {
		t.Error("Expected AcceptedAt to be set")
	}
}

func TestMockStore_DeleteProjectInvite(t *testing.T) {
	store := newMockInviteStore()

	// Create test invite
	invite := db.ProjectInvite{
		ID:            1,
		ProjectID:     1,
		Email:         stringPtr("test@example.com"),
		Role:          db.ProjectManager,
		Status:        db.ProjectInvitePending,
		Token:         "test-token",
		InviterUserID: 1,
		Created:       time.Now(),
	}
	store.CreateProjectInvite(invite)

	// Verify invite exists
	if _, exists := store.invites[1]; !exists {
		t.Error("Invite should exist before deletion")
	}

	// Delete invite
	err := store.DeleteProjectInvite(1, 1)
	if err != nil {
		t.Errorf("Failed to delete invite: %v", err)
	}

	// Verify invite was deleted
	if _, exists := store.invites[1]; exists {
		t.Error("Invite should not exist after deletion")
	}

	// Verify token was also removed
	if _, exists := store.invitesByToken["test-token"]; exists {
		t.Error("Invite token should not exist after deletion")
	}
}

func TestMockStore_GetProjectUser(t *testing.T) {
	store := newMockInviteStore()

	// Add project user
	projectUser := db.ProjectUser{
		ProjectID: 1,
		UserID:    2,
		Role:      db.ProjectManager,
	}
	store.CreateProjectUser(projectUser)

	// Test retrieval
	retrievedUser, err := store.GetProjectUser(1, 2)
	if err != nil {
		t.Errorf("Failed to get project user: %v", err)
	}

	if retrievedUser.Role != db.ProjectManager {
		t.Errorf("Expected role 'manager', got %s", retrievedUser.Role)
	}

	// Test non-existent user
	_, err = store.GetProjectUser(1, 999)
	if err != db.ErrNotFound {
		t.Error("Expected ErrNotFound for non-existent project user")
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Test ProjectInvite model validation
func TestProjectInviteStatus_IsValid(t *testing.T) {
	validStatuses := []db.ProjectInviteStatus{
		db.ProjectInvitePending,
		db.ProjectInviteAccepted,
		db.ProjectInviteDeclined,
		db.ProjectInviteExpired,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Errorf("Status %s should be valid", status)
		}
	}

	invalidStatus := db.ProjectInviteStatus("invalid")
	if invalidStatus.IsValid() {
		t.Error("Invalid status should not be valid")
	}
}

// Test helper function
func TestGetInviteTarget(t *testing.T) {
	// Test email-based invite
	emailInvite := db.ProjectInvite{
		Email: stringPtr("test@example.com"),
	}
	target := getInviteTarget(emailInvite)
	if target != "test@example.com" {
		t.Errorf("Expected 'test@example.com', got %s", target)
	}

	// Test user-based invite
	userInvite := db.ProjectInvite{
		UserID: intPtr(42),
	}
	target = getInviteTarget(userInvite)
	expected := "User ID 42"
	if target != expected {
		t.Errorf("Expected '%s', got %s", expected, target)
	}
}
