package routes

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
)

func authentication(c *gin.Context) {
	var userID int

	if authHeader := strings.ToLower(c.Request.Header.Get("authorization")); len(authHeader) > 0 {
		var token models.APIToken
		if err := database.Mysql.SelectOne(&token, "select * from user__token where id=? and expired=0", strings.Replace(authHeader, "bearer ", "", 1)); err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatus(403)
				return
			}

			panic(err)
		}

		userID = token.UserID
	} else {
		// fetch session from cookie
		cookie, err := c.Request.Cookie("semaphore")
		if err != nil {
			c.AbortWithStatus(403)
			return
		}

		value := make(map[string]interface{})
		if err = util.Cookie.Decode("semaphore", cookie.Value, &value); err != nil {
			c.AbortWithStatus(403)
			panic(err)
		}

		user, ok := value["user"]
		sessionVal, okSession := value["session"]
		if !ok || !okSession {
			c.AbortWithStatus(403)
			return
		}

		userID = user.(int)
		sessionID := sessionVal.(int)

		// fetch session
		var session models.Session
		if err := database.Mysql.SelectOne(&session, "select * from session where id=? and user_id=? and expired=0", sessionID, userID); err != nil {
			c.AbortWithStatus(403)
			return
		}

		if time.Now().Sub(session.LastActive).Hours() > 7*24 {
			// more than week old unused session
			// destroy.
			if _, err := database.Mysql.Exec("update session set expired=1 where id=?", sessionID); err != nil {
				panic(err)
			}

			c.AbortWithStatus(403)
			return
		}

		if _, err := database.Mysql.Exec("update session set last_active=NOW() where id=?", sessionID); err != nil {
			panic(err)
		}
	}

	user, err := models.FetchUser(userID)
	if err != nil {
		fmt.Println("Can't find user", err)
		c.AbortWithStatus(403)
		return
	}

	c.Set("user", user)
}
