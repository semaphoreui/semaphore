package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

// GetProjects returns all projects in this users context
func GetProjects(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	projects, err := helpers.Store(r).GetProjects(user.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, projects)
}

// AddProject adds a new project to the database
func AddProject(w http.ResponseWriter, r *http.Request) {
	var body db.Project

	user := context.Get(r, "user").(*db.User)


	if !helpers.Bind(w, r, &body) {
		return
	}

	body, err := helpers.Store(r).CreateProject(body)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	_, err = helpers.Store(r).CreateProjectUser(db.ProjectUser{ProjectID: body.ID, UserID: user.ID, Admin: true})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	desc := "Project Created"
	oType := "Project"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &body.ID,
		Description: &desc,
		ObjectType:  &oType,
		ObjectID:    &body.ID,
	})

	if err != nil {
		log.Error(err)
	}

	helpers.WriteJSON(w, http.StatusCreated, body)
}
