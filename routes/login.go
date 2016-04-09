package routes

import (
	"database/sql"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	sq "github.com/masterminds/squirrel"
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

	session := c.MustGet("session").(models.Session)
	session.UserID = &user.ID

	status := database.Redis.Set(session.ID, string(session.Encode()), 7*24*time.Hour)
	if err := status.Err(); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func logout(c *gin.Context) {
	session := c.MustGet("session").(models.Session)
	if err := database.Redis.Del(session.ID).Err(); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
