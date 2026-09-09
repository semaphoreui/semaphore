package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// allowedUserOptionKeys lists the suffixes a user is allowed to store via the
// per-user options API. The handler always namespaces them with the user ID.
var allowedUserOptionKeys = map[string]bool{
	"lang":              true,
	"nav.unpinnedItems": true,
}

func userOptionKey(userID int, suffix string) string {
	return fmt.Sprintf("user%d.%s", userID, suffix)
}

func getUserOptions(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)
	prefix := fmt.Sprintf("user%d.", user.ID)

	all, err := helpers.Store(r).GetOptions(db.RetrieveQueryParams{Filter: fmt.Sprintf("user%d", user.ID)})
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Can not get user options",
		})
		return
	}

	res := map[string]string{}
	for k, v := range all {
		res[strings.TrimPrefix(k, prefix)] = v
	}

	helpers.WriteJSON(w, http.StatusOK, res)
}

func setUserOption(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	var opt db.Option
	if !helpers.Bind(w, r, &opt) {
		return
	}

	if !allowedUserOptionKeys[opt.Key] {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unknown user option key",
		})
		return
	}

	if !json.Valid([]byte(opt.Value)) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "user option value must be valid JSON",
		})
		return
	}

	err := helpers.Store(r).SetOption(userOptionKey(user.ID, opt.Key), opt.Value)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Can not set user option",
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, opt)
}
