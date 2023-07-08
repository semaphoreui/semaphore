package projects

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
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

	helpers.WriteJSON(w, http.StatusOK, users)
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

// RemoveUser removes a user from a project team
func RemoveUser(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	projectUser := context.Get(r, "projectUser").(db.User)

	err := helpers.Store(r).DeleteProjectUser(project.ID, projectUser.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)
	objType := db.EventUser
	desc := "User ID " + strconv.Itoa(projectUser.ID) + " removed from team"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &projectUser.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// MakeUserAdmin writes the admin flag to the users account
func MakeUserAdmin(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "projectUser").(db.User)
	role := db.ProjectOwner

	if r.Method == "DELETE" {
		// strip admin
		role = db.ProjectTaskRunner
	}

	err := helpers.Store(r).UpdateProjectUser(db.ProjectUser{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      role,
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
