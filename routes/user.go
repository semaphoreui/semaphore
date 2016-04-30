package routes

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
)

func getUser(c *gin.Context) {
	c.JSON(200, c.MustGet("user"))
}

func getAPITokens(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var tokens []models.APIToken
	if _, err := database.Mysql.Select(&tokens, "select * from user__token where user_id=?", user.ID); err != nil {
		panic(err)
	}

	c.JSON(200, tokens)
}

func createAPIToken(c *gin.Context) {
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

func expireAPIToken(c *gin.Context) {
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
		c.AbortWithStatus(400)
		return
	}

	c.AbortWithStatus(204)
}
