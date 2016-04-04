package projects

import (
	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func UserMiddleware(c *gin.Context) {
	c.Next()
}

func GetUsers(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var users []models.User

	query, args, _ := squirrel.Select("u.*").
		From("project__user as pu").
		Join("user as u on pu.user_id=u.id").
		Where("pu.project_id=?", project.ID).
		ToSql()

	if _, err := database.Mysql.Select(&users, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, users)
}

func AddUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var user struct {
		UserID int  `json:"user_id" binding:"required"`
		Admin  bool `json:"admin"`
	}

	if err := c.Bind(&user); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("insert into project__user set user_id=?, project_id=?, admin=?", user.UserID, project.ID, user.Admin); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID, err := util.GetIntParam("user_id", c)
	if err != nil {
		return
	}

	if _, err := database.Mysql.Exec("delete from project__user where user_id=? and project_id=?", userID, project.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func MakeUserAdmin(c *gin.Context) {
	if c.Request.Method == "DELETE" {
		// strip admin
	}

	c.AbortWithStatus(501)
}
