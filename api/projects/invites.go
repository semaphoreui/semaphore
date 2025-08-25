package projects

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"text/template"

	"github.com/semaphoreui/semaphore/api/helpers"
	emailTemplates "github.com/semaphoreui/semaphore/api/templates"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/semaphoreui/semaphore/util/mailer"
	log "github.com/sirupsen/logrus"
)

// InviteMiddleware ensures an invite exists and loads it to the context
func InviteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		inviteID, err := helpers.GetIntParam("invite_id", w, r)
		if err != nil {
			return
		}

		invite, err := helpers.Store(r).GetProjectInvite(project.ID, inviteID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "projectInvite", invite)
		next.ServeHTTP(w, r)
	})
}

// GetInvites returns all invites for a project
func GetInvites(w http.ResponseWriter, r *http.Request) {
	// get single invite if invite ID specified in the request
	if invite := helpers.GetFromContext(r, "projectInvite"); invite != nil {
		helpers.WriteJSON(w, http.StatusOK, invite.(db.ProjectInvite))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)
	invites, err := helpers.Store(r).GetProjectInvites(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, invites)
}

// CreateInvite creates a new project invitation
func CreateInvite(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.UserFromContext(r)

	var request struct {
		UserID    *int               `json:"user_id,omitempty"`
		Email     *string            `json:"email,omitempty"`
		Role      db.ProjectUserRole `json:"role" binding:"required"`
		ExpiresAt *time.Time         `json:"expires_at,omitempty"`
	}

	if !helpers.Bind(w, r, &request) {
		return
	}

	if !request.Role.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate that either user_id or email is provided, but not both
	if (request.UserID == nil && request.Email == nil) || (request.UserID != nil && request.Email != nil) {
		helpers.WriteErrorStatus(w, "Either user_id or email must be provided, but not both", http.StatusBadRequest)
		return
	}

	// If user_id is provided, check if user exists
	if request.UserID != nil {
		_, err := helpers.Store(r).GetUser(*request.UserID)
		if err != nil {
			helpers.WriteError(w, fmt.Errorf("user not found"))
			return
		}

		// Check if user is already a member of the project
		_, err = helpers.Store(r).GetProjectUser(project.ID, *request.UserID)
		if err == nil {
			helpers.WriteErrorStatus(w, "User is already a member of this project", http.StatusConflict)
			return
		}
	}

	// Generate secure token for the invite
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		helpers.WriteError(w, fmt.Errorf("failed to generate invite token"))
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Set default expiration if not provided (7 days from now)
	expiresAt := request.ExpiresAt
	if expiresAt == nil {
		defaultExpiry := time.Now().Add(7 * 24 * time.Hour)
		expiresAt = &defaultExpiry
	}

	invite := db.ProjectInvite{
		ProjectID:     project.ID,
		UserID:        request.UserID,
		Email:         request.Email,
		Role:          request.Role,
		Status:        db.ProjectInvitePending,
		Token:         token,
		InviterUserID: user.ID,
		Created:       time.Now(),
		ExpiresAt:     expiresAt,
	}

	newInvite, err := helpers.Store(r).CreateProjectInvite(invite)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if newInvite.Email != nil {
		// Render email via template
		inviterName := user.Username
		if user.Name != "" {
			inviterName = user.Name
		}

		var body bytes.Buffer
		data := struct {
			InviterName string
			ProjectName string
			Role        db.ProjectUserRole
			Token       string
			ExpiresAt   string
			WebURL      string
			AcceptURL   string
		}{
			InviterName: inviterName,
			ProjectName: project.Name,
			Role:        newInvite.Role,
			Token:       newInvite.Token,
			ExpiresAt:   "",
			WebURL:      util.GetPublicHost(),
			AcceptURL:   "",
		}

		if newInvite.ExpiresAt != nil {
			data.ExpiresAt = newInvite.ExpiresAt.Format(time.RFC1123)
		}

		// Optionally construct a direct accept URL if we decide to support one later
		// data.AcceptURL = fmt.Sprintf("%s/accept?token=%s", data.WebURL, newInvite.Token)

		tpl, err := template.ParseFS(emailTemplates.FS, "invite.tmpl")
		if err == nil {
			_ = tpl.Execute(&body, data)
		}

		if body.Len() == 0 {
			// Fallback minimal body
			body.WriteString(fmt.Sprintf("Invitation to join %s as %s. Token: %s", data.ProjectName, data.Role, data.Token))
		}

		subject := fmt.Sprintf("Invitation to join project %s", project.Name)

		if err := mailer.Send(
			util.Config.EmailSecure,
			util.Config.EmailTls,
			util.Config.EmailHost,
			util.Config.EmailPort,
			util.Config.EmailUsername,
			util.Config.EmailPassword,
			util.Config.EmailSender,
			*newInvite.Email,
			subject,
			body.String(),
		); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"email":   *newInvite.Email,
				"context": "project_invite",
			}).Error("failed to send project invitation email")
		} else {
			log.WithFields(log.Fields{
				"email":   *newInvite.Email,
				"context": "project_invite",
			}).Info("project invitation email sent")
		}
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      user.ID,
		ProjectID:   project.ID,
		ObjectType:  "project_invite",
		ObjectID:    newInvite.ID,
		Description: fmt.Sprintf("Project invitation created for %s with role %s", getInviteTarget(newInvite), newInvite.Role),
	})

	helpers.WriteJSON(w, http.StatusCreated, newInvite)
}

// AcceptInvite accepts a project invitation using token
func AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if !helpers.Bind(w, r, &request) {
		return
	}

	invite, err := helpers.Store(r).GetProjectInviteByToken(request.Token)
	if err != nil {
		helpers.WriteErrorStatus(w, "Invalid or expired invitation token", http.StatusNotFound)
		return
	}

	// Check if invite is still valid
	if invite.Status != db.ProjectInvitePending {
		helpers.WriteErrorStatus(w, "Invitation is no longer valid", http.StatusBadRequest)
		return
	}

	if invite.ExpiresAt != nil && time.Now().After(*invite.ExpiresAt) {
		helpers.WriteErrorStatus(w, "Invitation has expired", http.StatusBadRequest)
		return
	}

	currentUser := helpers.UserFromContext(r)

	// If invite is for a specific user, verify it matches
	if invite.UserID != nil && *invite.UserID != currentUser.ID {
		helpers.WriteErrorStatus(w, "This invitation is not for your account", http.StatusForbidden)
		return
	}

	// If invite is by email, verify email matches
	if invite.Email != nil && *invite.Email != currentUser.Email {
		helpers.WriteErrorStatus(w, "This invitation is not for your email address", http.StatusForbidden)
		return
	}

	// Check if user is already a member of the project
	_, err = helpers.Store(r).GetProjectUser(invite.ProjectID, currentUser.ID)
	if err == nil {
		helpers.WriteErrorStatus(w, "You are already a member of this project", http.StatusConflict)
		return
	}

	// Create project user
	_, err = helpers.Store(r).CreateProjectUser(db.ProjectUser{
		ProjectID: invite.ProjectID,
		UserID:    currentUser.ID,
		Role:      invite.Role,
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Update invite status
	now := time.Now()
	invite.Status = db.ProjectInviteAccepted
	invite.AcceptedAt = &now
	invite.UserID = &currentUser.ID // Set user ID if it was an email invite

	err = helpers.Store(r).UpdateProjectInvite(invite)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      currentUser.ID,
		ProjectID:   invite.ProjectID,
		ObjectType:  "project_invite",
		ObjectID:    invite.ID,
		Description: fmt.Sprintf("Project invitation accepted by %s", currentUser.Username),
	})

	var result struct {
		ProjectID int `json:"project_id"`
	}

	result.ProjectID = invite.ProjectID

	helpers.WriteJSON(w, http.StatusOK, result)
}

// UpdateInvite updates an existing project invitation
func UpdateInvite(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	invite := helpers.GetFromContext(r, "projectInvite").(db.ProjectInvite)
	user := helpers.UserFromContext(r)

	var request struct {
		Status db.ProjectInviteStatus `json:"status"`
	}

	if !helpers.Bind(w, r, &request) {
		return
	}

	if !request.Status.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only allow certain status transitions
	if invite.Status != db.ProjectInvitePending && request.Status != db.ProjectInviteExpired {
		helpers.WriteErrorStatus(w, "Cannot modify non-pending invitations", http.StatusBadRequest)
		return
	}

	invite.Status = request.Status

	err := helpers.Store(r).UpdateProjectInvite(invite)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      user.ID,
		ProjectID:   project.ID,
		ObjectType:  "project_invite",
		ObjectID:    invite.ID,
		Description: fmt.Sprintf("Project invitation status changed to %s", request.Status),
	})

	w.WriteHeader(http.StatusNoContent)
}

// DeleteInvite removes a project invitation
func DeleteInvite(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	invite := helpers.GetFromContext(r, "projectInvite").(db.ProjectInvite)
	user := helpers.UserFromContext(r)

	err := helpers.Store(r).DeleteProjectInvite(project.ID, invite.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      user.ID,
		ProjectID:   project.ID,
		ObjectType:  "project_invite",
		ObjectID:    invite.ID,
		Description: fmt.Sprintf("Project invitation deleted for %s", getInviteTarget(invite)),
	})

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to get invite target (user or email)
func getInviteTarget(invite db.ProjectInvite) string {
	if invite.Email != nil {
		return *invite.Email
	}
	return fmt.Sprintf("User ID %d", *invite.UserID)
}
