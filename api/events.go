package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

// nolint: gocyclo
func getEvents(w http.ResponseWriter, r *http.Request, limit int) {
	user := context.Get(r, "user").(*db.User)
	projectObj, exists := context.GetOk(r, "project")

	var err error
	var events []db.Event

	if exists {
		project := projectObj.(db.Project)

		if !user.Admin { // check permissions to view events
			_, err = helpers.Store(r).GetProjectUser(project.ID, user.ID)
		}

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		events, err = helpers.Store(r).GetEvents(project.ID, db.RetrieveQueryParams{Count: limit})
	} else {
		events, err = helpers.Store(r).GetUserEvents(user.ID, db.RetrieveQueryParams{Count: limit})
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, events)
}

func getLastEvents(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, 200)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, 0)
}
