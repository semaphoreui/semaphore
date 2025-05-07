package api

import (
	"github.com/gorilla/context"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func clearCache(w http.ResponseWriter, r *http.Request) {
	currentUser := context.Get(r, "user").(*db.User)

	if !currentUser.Admin {
		helpers.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error": "User must be admin",
		})
		return
	}

	err := util.Config.ClearTmpDir()
	if err != nil {
		log.Error(err)
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Can not clear cache",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
