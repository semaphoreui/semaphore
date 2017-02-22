package api

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"time"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
)

func getUser(w http.ResponseWriter, r *http.Request) {
	if u, exists := c.Get("_user"); exists {
		c.JSON(200, u)
		return
	}

	c.JSON(200, c.MustGet("user"))
}

func getAPITokens(w http.ResponseWriter, r *http.Request) {
	user := c.MustGet("user").(*models.User)

	var tokens []models.APIToken
	if _, err := database.Mysql.Select(&tokens, "select * from user__token where user_id=?", user.ID); err != nil {
		panic(err)
	}

	c.JSON(200, tokens)
}

func createAPIToken(w http.ResponseWriter, r *http.Request) {
	user := c.MustGet("user").(*models.User)
	tokenID := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenID); err != nil {
		panic(err)
	}

	token := models.APIToken{
		ID:      strings.ToLower(base64.URLEncoding.EncodeToString(tokenID)),
		Created: time.Now(),
		UserID:  user.ID,
		Expired: false,
	}

	if err := database.Mysql.Insert(&token); err != nil {
		panic(err)
	}

	c.JSON(201, token)
}

func expireAPIToken(w http.ResponseWriter, r *http.Request) {
	user := c.MustGet("user").(*models.User)

	tokenID := c.Param("token_id")
	res, err := database.Mysql.Exec("update user__token set expired=1 where id=? and user_id=?", tokenID, user.ID)
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
