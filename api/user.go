package api

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strings"
)

func getUser(w http.ResponseWriter, r *http.Request) {
	if u, exists := context.GetOk(r, "_user"); exists {
		helpers.WriteJSON(w, http.StatusOK, u)
		return
	}

	var user struct {
		db.User
		CanCreateProject bool `json:"can_create_project"`
	}

	user.User = *context.Get(r, "user").(*db.User)
	user.CanCreateProject = user.Admin || util.Config.NonAdminCanCreateProject

	helpers.WriteJSON(w, http.StatusOK, user)
}

func getAPITokens(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	tokens, err := helpers.Store(r).GetAPITokens(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tokens)
}

func createAPIToken(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)
	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		panic(err)
	}

	token, err := helpers.Store(r).CreateAPIToken(db.APIToken{
		ID:      strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		UserID:  user.ID,
		Expired: false,
	})
	if err != nil {
		panic(err)
	}

	helpers.WriteJSON(w, http.StatusCreated, token)
}

func expireAPIToken(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	tokenID := mux.Vars(r)["token_id"]

	err := helpers.Store(r).ExpireAPIToken(user.ID, tokenID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
