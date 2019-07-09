package api

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/mulekick"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func getUser(w http.ResponseWriter, r *http.Request) {
	if u, exists := context.GetOk(r, "_user"); exists {
		mulekick.WriteJSON(w, http.StatusOK, u)
		return
	}

	mulekick.WriteJSON(w, http.StatusOK, context.Get(r, "user"))
}

func getAPITokens(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	var tokens []db.APIToken
	if _, err := db.Mysql.Select(&tokens, "select * from user__token where user_id=?", user.ID); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, tokens)
}

func createAPIToken(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)
	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		panic(err)
	}

	token := db.APIToken{
		ID:      strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		Created: db.GetParsedTime(time.Now()),
		UserID:  user.ID,
		Expired: false,
	}

	if err := db.Mysql.Insert(&token); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, token)
}

func expireAPIToken(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	tokenID := mux.Vars(r)["token_id"]
	res, err := db.Mysql.Exec("update user__token set expired=1 where id=? and user_id=?", tokenID, user.ID)
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
