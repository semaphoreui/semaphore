package routes

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"gopkg.in/redis.v3"
)

func resetSessionExpiry(sessionID string) {
	if err := database.Redis.Expire(sessionID, 7*24*time.Hour).Err(); err != nil {
		fmt.Println("Cannot reset session expiry:", err)
	}
}

func authentication(c *gin.Context) {
	cookie, err := c.Request.Cookie("semaphore")
	if err != nil {
		// create cookie
		new_cookie := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, new_cookie); err != nil {
			panic(err)
		}

		cookie_value := url.QueryEscape(base64.URLEncoding.EncodeToString(new_cookie))
		cookie = &http.Cookie{Name: "semaphore", Value: cookie_value, Path: "/", HttpOnly: true}
		http.SetCookie(c.Writer, cookie)
	}

	redis_key := "session:" + cookie.Value
	s, err := database.Redis.Get(redis_key).Result()
	if err == redis.Nil {
		// create a session
		temp_session := models.Session{}
		s = string(temp_session.Encode())

		if err := database.Redis.Set(redis_key, s, 0).Err(); err != nil {
			panic(err)
		}
	} else if err != nil {
		fmt.Println("Cannot get cookie from redis:", err)
		c.AbortWithStatus(500)

		return
	}

	// reset session expiry
	go resetSessionExpiry(redis_key)

	sess, err := models.DecodeSession(cookie.Value, s)
	if err != nil {
		fmt.Println("Cannot decode session:", err)
		util.AuthFailed(c)
		return
	}

	sess.ID = cookie.Value
	c.Set("session", sess)

	if sess.Login != nil {
		user := &models.User{}
		// user.Parse(*sess.Login)

		// if err := user.GetUser(); err != nil {
		// 	panic(err)
		// }

		c.Set("user", user)
	}

	c.Next()
}

func MustAuthenticate(c *gin.Context) {
	if _, exists := c.Get("user"); !exists {
		util.AuthFailed(c)
		return
	}

	c.Next()
}
