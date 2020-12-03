package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/models"
	"net/http"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// GetProjects returns all projects in this users context
func GetProjects(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*models.User)

	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("pu.user_id=?", user.ID).
		OrderBy("p.name").
		ToSql()

	util.LogWarning(err)
	var projects []models.Project
	if _, err := helpers.Store(r).Sql().Select(&projects, query, args...); err != nil {
		panic(err)
	}

	helpers.WriteJSON(w, http.StatusOK, projects)
}

// AddProject adds a new project to the database
func AddProject(w http.ResponseWriter, r *http.Request) {
	var body models.Project

	user := context.Get(r, "user").(*models.User)


	if !helpers.Bind(w, r, &body) {
		return
	}

	body, err := helpers.Store(r).CreateProject(body)
	if err != nil {
		panic(err)
	}

	_, err = helpers.Store(r).CreateProjectUser(models.ProjectUser{ProjectID: body.ID, UserID: user.ID, Admin: true})

	if err != nil {
		panic(err)
	}

	desc := "Project Created"
	oType := "Project"
	_, err = helpers.Store(r).CreateEvent(models.Event{
		ProjectID:   &body.ID,
		Description: &desc,
		ObjectType:  &oType,
		ObjectID:    &body.ID,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, body)
}
