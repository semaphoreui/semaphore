package routes

import (
	"database/sql"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	sq "github.com/masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
)

func login(c *gin.Context) {
	var login struct {
		Auth     string `json:"auth" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&login); err != nil {
		return
	}

	login.Auth = strings.ToLower(login.Auth)

	q := sq.Select("*").
		From("user")

	_, err := mail.ParseAddress(login.Auth)
	if err == nil {
		q = q.Where("email=?", login.Auth)
	} else {
		q = q.Where("username=?", login.Auth)
	}

	query, args, _ := q.ToSql()

	var user models.User
	if err := database.Mysql.SelectOne(&user, query, args...); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(400)
			return
		}

		panic(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		c.AbortWithStatus(400)
		return
	}

	session := models.Session{
		UserID:     user.ID,
		Created:    time.Now(),
		LastActive: time.Now(),
		IP:         c.ClientIP(),
		UserAgent:  c.Request.Header.Get("user-agent"),
		Expired:    false,
	}
	if err := database.Mysql.Insert(&session); err != nil {
		panic(err)
	}

	encoded, err := util.Cookie.Encode("semaphore", map[string]interface{}{
		"user":    user.ID,
		"session": session.ID,
	})
	if err != nil {
		panic(err)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:  "semaphore",
		Value: encoded,
		Path:  "/",
	})

	c.AbortWithStatus(204)
}

func logout(c *gin.Context) {
	c.SetCookie("semaphore", "", -1, "/", "", false, true)
	c.AbortWithStatus(204)
}
