package projects

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	"net/http"
	"strconv"
)

// UserMiddleware ensures a user exists and loads it to the context
func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		userID, err := helpers.GetIntParam("user_id", w, r)
		if err != nil {
			return
		}

		_, err = helpers.Store(r).GetProjectUser(project.ID, userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "projectUser", user)
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
	if user := context.Get(r, "projectUser"); user != nil {
		helpers.WriteJSON(w, http.StatusOK, user.(db.User))
		return
	}

	project := context.Get(r, "project").(db.Project)
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
	project := context.Get(r, "project").(db.Project)
	var projectUser struct {
		UserID int                `json:"user_id" binding:"required"`
		Role   db.ProjectUserRole `json:"role"`
	}

	if !helpers.Bind(w, r, &projectUser) {
		return
	}

	if !projectUser.Role.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
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

	user := context.Get(r, "user").(*db.User)
	objType := db.EventUser
	desc := "User ID " + strconv.Itoa(projectUser.UserID) + " added to team"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &projectUser.UserID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// removeUser removes a user from a project team
func removeUser(targetUser db.User, w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	me := context.Get(r, "user").(*db.User) // logged in user
	myRole := context.Get(r, "projectUserRole").(db.ProjectUserRole)

	if !me.Admin && targetUser.ID == me.ID && myRole == db.ProjectOwner {
		helpers.WriteError(w, fmt.Errorf("owner can not left the project"))
		return
	}

	err := helpers.Store(r).DeleteProjectUser(project.ID, targetUser.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	objType := db.EventUser
	desc := "User ID " + strconv.Itoa(targetUser.ID) + " removed from team"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &me.ID,
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &targetUser.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// LeftProject removes a user from a project team
func LeftProject(w http.ResponseWriter, r *http.Request) {
	me := context.Get(r, "user").(*db.User) // logged in user
	removeUser(*me, w, r)
}

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	targetUser := context.Get(r, "projectUser").(db.User) // target user
	removeUser(targetUser, w, r)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	me := context.Get(r, "user").(*db.User) // logged in user
	targetUser := context.Get(r, "projectUser").(db.User)
	targetUserRole := context.Get(r, "projectUserRole").(db.ProjectUserRole)

	if !me.Admin && targetUser.ID == me.ID && targetUserRole == db.ProjectOwner {
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
		w.WriteHeader(http.StatusBadRequest)
		return
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

	w.WriteHeader(http.StatusNoContent)
}
