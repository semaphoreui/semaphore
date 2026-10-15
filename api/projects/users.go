package projects

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

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

		memberRoles, err := helpers.Store(r).GetProjectUserRoles(project.ID, userID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
		if len(memberRoles) == 0 {
			helpers.WriteError(w, db.ErrNotFound)
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "projectUser", user)
		r = helpers.SetContextValue(r, "projectMemberRoles", memberRoles)
		next.ServeHTTP(w, r)
	})
}

type projUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func callerIsProjectOwner(r *http.Request) bool {
	roles, _ := helpers.GetFromContext(r, "projectRoles").([]db.Role)
	return slices.ContainsFunc(roles, func(role db.Role) bool {
		return role.Builtin && role.Slug == string(db.ProjectOwner)
	})
}

func roleRefs(roles []db.Role) []audit.RoleRef {
	refs := make([]audit.RoleRef, 0, len(roles))
	for _, role := range roles {
		refs = append(refs, audit.NewRoleRef(role))
	}
	return refs
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

	result := make([]projUser, 0)

	for _, user := range users {
		role := ""
		if len(user.Roles) > 0 {
			role = user.Roles[0].Slug
		}
		result = append(result, projUser{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Role:     role,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

// AddUser adds a user to a projects team in the database
func AddUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)
	var projectUser struct {
		UserID int    `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	role, err := store.GetProjectOrGlobalRoleBySlug(project.ID, projectUser.Role)
	if errors.Is(err, db.ErrNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = store.CreateProjectUser(project.ID, projectUser.UserID, []int{role.ID})

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
		Metadata:  audit.MembershipMetadata{Roles: []audit.RoleRef{audit.NewRoleRef(role)}},
	})

	w.WriteHeader(http.StatusNoContent)
}

// removeUser removes a user from a project team
func removeUser(targetUser db.User, roles []db.Role, w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user

	event := audit.Event{
		Kind:      audit.IAMMembershipRemove,
		Target:    audit.UserTarget(targetUser.ID, targetUser.Username),
		ProjectID: project.ID,
		Metadata:  audit.MembershipMetadata{Roles: roleRefs(roles), SelfRemoval: targetUser.ID == me.ID},
	}

	if !me.Admin && targetUser.ID == me.ID && callerIsProjectOwner(r) {
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
	roles, _ := helpers.GetFromContext(r, "projectRoles").([]db.Role)
	removeUser(*me, roles, w, r)
}

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User) // target user
	roles, _ := helpers.GetFromContext(r, "projectMemberRoles").([]db.Role)
	removeUser(targetUser, roles, w, r)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User)
	oldRoles, _ := helpers.GetFromContext(r, "projectMemberRoles").([]db.Role)

	event := audit.Event{
		Kind:      audit.IAMProjectRoleChange,
		Target:    audit.UserTarget(targetUser.ID, targetUser.Username),
		ProjectID: project.ID,
		Metadata:  audit.ProjectRoleMetadata{OldRoles: roleRefs(oldRoles)},
	}

	if !me.Admin && targetUser.ID == me.ID && callerIsProjectOwner(r) {
		event.Outcome = audit.OutcomeFailure
		event.Reason = audit.ReasonOwnerSelfChange
		helpers.Audit(r).Record(r.Context(), event)
		helpers.WriteError(w, fmt.Errorf("owner can not change his role in the project"))
		return
	}

	var projectUser struct {
		Role string `json:"role"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	newRole, err := store.GetProjectOrGlobalRoleBySlug(project.ID, projectUser.Role)
	if errors.Is(err, db.ErrNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = store.SetProjectUserRoles(project.ID, targetUser.ID, []int{newRole.ID})

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
		OldRoles: roleRefs(oldRoles),
		NewRoles: []audit.RoleRef{audit.NewRoleRef(newRole)},
	}
	helpers.Audit(r).Record(r.Context(), event)

	w.WriteHeader(http.StatusNoContent)
}
