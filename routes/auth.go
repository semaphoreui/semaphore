package routes

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"gopkg.in/redis.v3"
)

func resetSessionExpiry(sessionID string, ttl time.Duration) {
	var cmd *redis.BoolCmd

	if ttl == 0 {
		cmd = database.Redis.Persist(sessionID)
	} else {
		cmd = database.Redis.Expire(sessionID, ttl)
	}

	if err := cmd.Err(); err != nil {
		fmt.Println("Cannot reset session expiry:", err)
	}
}

func authentication(c *gin.Context) {
	var redisKey string
	ttl := 7 * 24 * time.Hour

	if authHeader := strings.ToLower(c.Request.Header.Get("authorization")); len(authHeader) > 0 {
		redisKey = "token-session:" + strings.Replace(authHeader, "bearer ", "", 1)
		ttl = 0
	} else {
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

		redisKey = "session:" + cookie.Value
	}

	s, err := database.Redis.Get(redisKey).Result()
	if err == redis.Nil {
		// create a session
		temp_session := models.Session{}
		s = string(temp_session.Encode())

		if err := database.Redis.Set(redisKey, s, 0).Err(); err != nil {
			panic(err)
		}
	} else if err != nil {
		fmt.Println("Cannot get session from redis:", err)
		c.AbortWithStatus(500)

		return
	}

	sess, err := models.DecodeSession(redisKey, s)
	if err != nil {
		fmt.Println("Cannot decode session:", err)
		util.AuthFailed(c)
		return
	}

	if sess.UserID != nil {
		user, err := models.FetchUser(*sess.UserID)
		if err != nil {
			fmt.Println("Can't find user", err)
			c.AbortWithStatus(403)
			return
		}

		c.Set("user", user)
	}

	// reset session expiry
	go resetSessionExpiry(redisKey, ttl)

	c.Set("session", sess)
	c.Next()
}

func MustAuthenticate(c *gin.Context) {
	if _, exists := c.Get("user"); !exists {
		util.AuthFailed(c)
		return
	}

	c.Next()
}
