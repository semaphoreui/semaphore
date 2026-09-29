package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// UserMiddleware ensures a user exists and loads it to the context
func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		userID, ok := helpers.GetIntParamOrAbort("user_id", w, r)
		if !ok {
			return
		}

		_, err := helpers.Store(r).GetProjectUser(project.ID, userID)

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
		next.ServeHTTP(w, r)
	})
}

type projUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	RoleID   int    `json:"role_id"`
}

func callerIsProjectOwner(r *http.Request) bool {
	role, _ := helpers.GetFromContext(r, "projectRole").(*db.Role)
	return role != nil && role.BuiltinKey != nil && *role.BuiltinKey == db.BuiltinRoleOwner
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
		result = append(result, projUser{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			RoleID:   user.RoleID,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, result)
}

// AddUser adds a user to a projects team in the database
func AddUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)
	var projectUser struct {
		UserID int `json:"user_id" binding:"required"`
		RoleID int `json:"role_id" binding:"required"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	_, err := db.ResolveRoleForProject(store, projectUser.RoleID, project.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    projectUser.UserID,
		RoleID:    projectUser.RoleID,
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

	w.WriteHeader(http.StatusNoContent)
}

// removeUser removes a user from a project team
func removeUser(targetUser db.User, w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user

	if !me.Admin && targetUser.ID == me.ID && callerIsProjectOwner(r) {
		helpers.WriteError(w, fmt.Errorf("owner can not left the project"))
		return
	}

	err := helpers.Store(r).DeleteProjectUser(project.ID, targetUser.ID)

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

	w.WriteHeader(http.StatusNoContent)
}

// LeftProject removes a user from a project team
func LeftProject(w http.ResponseWriter, r *http.Request) {
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	removeUser(*me, w, r)
}

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User) // target user
	removeUser(targetUser, w, r)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	store := helpers.Store(r)
	me := helpers.GetFromContext(r, "user").(*db.User) // logged in user
	targetUser := helpers.GetFromContext(r, "projectUser").(db.User)

	if !me.Admin && targetUser.ID == me.ID && callerIsProjectOwner(r) {
		helpers.WriteError(w, fmt.Errorf("owner can not change his role in the project"))
		return
	}

	var projectUser struct {
		RoleID int `json:"role_id" binding:"required"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	_, err := db.ResolveRoleForProject(store, projectUser.RoleID, project.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = store.UpdateProjectUser(db.ProjectUser{
		UserID:    targetUser.ID,
		ProjectID: project.ID,
		RoleID:    projectUser.RoleID,
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

	w.WriteHeader(http.StatusNoContent)
}
