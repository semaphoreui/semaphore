package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	"net/http"
)

func setOption(w http.ResponseWriter, r *http.Request) {
	currentUser := context.Get(r, "user").(*db.User)

	if !currentUser.Admin {
		helpers.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error": "User must be admin",
		})
		return
	}

	var option db.Option
	if !helpers.Bind(w, r, &option) {
		return
	}

	err := helpers.Store(r).SetOption(option.Key, option.Value)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Can not set option",
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, option)
}

func getOptions(w http.ResponseWriter, r *http.Request) {
	currentUser := context.Get(r, "user").(*db.User)

	if !currentUser.Admin {
		helpers.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error": "User must be admin",
		})
		return
	}

	options, err := helpers.Store(r).GetOptions(db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Can not get options",
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, options)
}
