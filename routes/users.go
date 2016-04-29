package routes

import (
	"database/sql"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
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

	if _, err := database.Mysql.Exec("update user set name=?, username=?, email=? where id=?", user.Name, user.Username, user.Email, oldUser.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func updateUserPassword(c *gin.Context) {
	var pwd struct {
		Pwd string `json:"password"`
	}

	userID, err := util.GetIntParam("user_id", c)
	if err != nil {
		return
	}

	if err := c.Bind(&pwd); err != nil {
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(pwd.Pwd), 11)
	if _, err := database.Mysql.Exec("update user set password=? where id=?", string(password), userID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
