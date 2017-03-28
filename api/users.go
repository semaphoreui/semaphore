package api

import (
	"database/sql"
	"time"

	log "github.com/Sirupsen/logrus"
	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func getUsers(c *gin.Context) {
	var users []models.User
	if _, err := database.Mysql.Select(&users, "select * from user"); err != nil {
		panic(err)
	}

	c.JSON(200, users)
}

func addUser(c *gin.Context) {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return
	}

	user.Created = time.Now()

	if err := database.Mysql.Insert(&user); err != nil {
		panic(err)
	}

	c.JSON(201, user)
}

func getUserMiddleware(c *gin.Context) {
	userID, err := util.GetIntParam("user_id", c)
	if err != nil {
		return
	}

	var user models.User
	if err := database.Mysql.SelectOne(&user, "select * from user where id=?", userID); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("_user", user)
	c.Next()
}

func updateUser(c *gin.Context) {
	oldUser := c.MustGet("_user").(models.User)

	var user models.User
	if err := c.Bind(&user); err != nil {
		return
	}

	if oldUser.External == true && oldUser.Username != user.Username {
		log.Warn("Username is not editable for external LDAP users")
		c.AbortWithStatus(400)
		return
	}

	if _, err := database.Mysql.Exec("update user set name=?, username=?, email=?, alert=? where id=?", user.Name, user.Username, user.Email, user.Alert, oldUser.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func updateUserPassword(c *gin.Context) {
	user := c.MustGet("_user").(models.User)
	var pwd struct {
		Pwd string `json:"password"`
	}

	if user.External == true {
		log.Warn("Password is not editable for external LDAP users")
		c.AbortWithStatus(400)
		return
	}

	if err := c.Bind(&pwd); err != nil {
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(pwd.Pwd), 11)
	if _, err := database.Mysql.Exec("update user set password=? where id=?", string(password), user.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func deleteUser(c *gin.Context) {
	user := c.MustGet("_user").(models.User)

	if _, err := database.Mysql.Exec("delete from project__user where user_id=?", user.ID); err != nil {
		panic(err)
	}
	if _, err := database.Mysql.Exec("delete from user where id=?", user.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
