package projects

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
)

// UserMiddleware ensures a user exists and loads it to the context
func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		userID, ok := helpers.GetIntParamOrAbort("user_id", w, r)
		if !ok {
			return
		}

		membership, err := helpers.Store(r).GetProjectUser(project.ID, userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "projectUser", user)
		r = helpers.SetContextValue(r, "projectMembership", membership)
		next.ServeHTTP(w, r)
	})
}

type projUser struct {
	ID       int                `json:"id"`
	Username string             `json:"username"`
	Name     string             `json:"name"`
	Role     db.ProjectUserRole `json:"role"`
}

// GetUsers returns all users in a project
func GetUsers(w http.ResponseWriter, r *http.Request) {

	// get single user if user ID specified in the request
	if user := helpers.GetFromContext(r, "projectUser"); user != nil {
		helpers.WriteJSON(w, http.StatusOK, user.(db.User))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)
	users, err := helpers.Store(r).GetProjectUsers(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var result = make([]projUser, 0)

	for _, user := range users {
		result = append(result, projUser{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Role:     user.Role,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

// AddUser adds a user to a projects team in the database
func AddUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	var projectUser struct {
		UserID int                `json:"user_id" binding:"required"`
		Role   db.ProjectUserRole `json:"role"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	if !projectUser.Role.IsValid() {
		_, err := helpers.Store(r).GetProjectOrGlobalRoleBySlug(project.ID, string(projectUser.Role))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	_, err := helpers.Store(r).CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    projectUser.UserID,
		Role:      projectUser.Role,
	})

	if err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventUser,
		ObjectID:    projectUser.UserID,
		Description: fmt.Sprintf("User ID %d added to team", projectUser.UserID),
	})

	target := audit.UserTarget(projectUser.UserID, "")
	if added, err := helpers.Store(r).GetUser(projectUser.UserID); err == nil {
		target.Name = added.Username
	}
	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.IAMMembershipAdd,
		Target:    target,
		ProjectID: project.ID,
		Metadata:  audit.MembershipMetadata{Role: audit.TruncateName(string(projectUser.Role), audit.MaxNameBytes)},
	})

	w.WriteHeader(http.StatusNoContent)
}

// removeUser removes a user from a project team
func removeUser(targetUser db.User, role db.ProjectUserRole, w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	myRole := helpers.GetFromContext(r, "projectUserRole").(db.ProjectUserRole)

	event := audit.Event{
		Kind:      audit.IAMMembershipRemove,
		Target:    audit.UserTarget(targetUser.ID, targetUser.Username),
		ProjectID: project.ID,
		Metadata:  audit.MembershipMetadata{Role: audit.TruncateName(string(role), audit.MaxNameBytes), SelfRemoval: targetUser.ID == me.ID},
	}

	if !me.Admin && targetUser.ID == me.ID && myRole == db.ProjectOwner {
		event.Outcome = audit.OutcomeFailure
		event.Reason = audit.ReasonOwnerSelfChange
		helpers.Audit(r).Record(r.Context(), event)
		helpers.WriteError(w, fmt.Errorf("owner can not left the project"))
		return
	}

	err := helpers.Store(r).DeleteProjectUser(project.ID, targetUser.ID)
	if errors.Is(err, db.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventUser,
		ObjectID:    targetUser.ID,
		Description: fmt.Sprintf("User ID %d removed from team", targetUser.ID),
	})
	helpers.Audit(r).Record(r.Context(), event)

	w.WriteHeader(http.StatusNoContent)
}

// LeftProject removes a user from a project team
func LeftProject(w http.ResponseWriter, r *http.Request) {
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	myRole := helpers.GetFromContext(r, "projectUserRole").(db.ProjectUserRole)
	removeUser(*me, myRole, w, r)
}

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User) // target user
	membership := helpers.GetFromContext(r, "projectMembership").(db.ProjectUser)
	removeUser(targetUser, membership.Role, w, r)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User)
	targetUserRole := helpers.GetFromContext(r, "projectUserRole").(db.ProjectUserRole)
	membership := helpers.GetFromContext(r, "projectMembership").(db.ProjectUser)

	event := audit.Event{
		Kind:      audit.IAMProjectRoleChange,
		Target:    audit.UserTarget(targetUser.ID, targetUser.Username),
		ProjectID: project.ID,
		Metadata:  audit.ProjectRoleMetadata{OldRole: audit.TruncateName(string(membership.Role), audit.MaxNameBytes)},
	}

	if !me.Admin && targetUser.ID == me.ID && targetUserRole == db.ProjectOwner {
		event.Outcome = audit.OutcomeFailure
		event.Reason = audit.ReasonOwnerSelfChange
		helpers.Audit(r).Record(r.Context(), event)
		helpers.WriteError(w, fmt.Errorf("owner can not change his role in the project"))
		return
	}

	var projectUser struct {
		Role db.ProjectUserRole `json:"role"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	if !projectUser.Role.IsValid() {
		_, err := helpers.Store(r).GetProjectOrGlobalRoleBySlug(project.ID, string(projectUser.Role))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	err := helpers.Store(r).UpdateProjectUser(db.ProjectUser{
		UserID:    targetUser.ID,
		ProjectID: project.ID,
		Role:      projectUser.Role,
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventUser,
		ObjectID:    targetUser.ID,
		Description: fmt.Sprintf("Changed role for User ID %d", targetUser.ID),
	})
	event.Metadata = audit.ProjectRoleMetadata{
		OldRole: audit.TruncateName(string(membership.Role), audit.MaxNameBytes),
		NewRole: audit.TruncateName(string(projectUser.Role), audit.MaxNameBytes),
	}
	helpers.Audit(r).Record(r.Context(), event)

	w.WriteHeader(http.StatusNoContent)
}
