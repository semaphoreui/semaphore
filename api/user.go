package api

import (
	"crypto/rand"
	"encoding/base64"
	util2 "github.com/ansible-semaphore/semaphore/api/util"
	"github.com/ansible-semaphore/semaphore/models"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func getUser(w http.ResponseWriter, r *http.Request) {
	if u, exists := context.GetOk(r, "_user"); exists {
		util2.WriteJSON(w, http.StatusOK, u)
		return
	}

	util2.WriteJSON(w, http.StatusOK, context.Get(r, "user"))
}

func getAPITokens(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*models.User)

	var tokens []models.APIToken
	if _, err := util2.GetStore(r).Sql().Select(&tokens, "select * from user__token where user_id=?", user.ID); err != nil {
		panic(err)
	}

	util2.WriteJSON(w, http.StatusOK, tokens)
}

func createAPIToken(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*models.User)
	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		panic(err)
	}

	token := models.APIToken{
		ID:      strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		Created: time.Now(), // TODO: use GetParsedTime
		UserID:  user.ID,
		Expired: false,
	}

	createdToken, err := util2.GetStore(r).CreateAPIToken(token)
	if err != nil {
		panic(err)
	}

	util2.WriteJSON(w, http.StatusCreated, createdToken)
}

func expireAPIToken(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*models.User)

	tokenID := mux.Vars(r)["token_id"]
	res, err := util2.GetStore(r).Sql().Exec("update user__token set expired=1 where id=? and user_id=?", tokenID, user.ID)
	if err != nil {
		panic(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	if affected == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
