package projects

import (
	"database/sql"
	"strconv"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func UserMiddleware(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID, err := util.GetIntParam("user_id", c)
	if err != nil {
		return
	}

	var user models.User
	if err := database.Mysql.SelectOne(&user, "select u.* from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=?", userID, project.ID); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("projectUser", user)
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

	objType := "user"
	desc := "User ID " + strconv.Itoa(user.UserID) + " added to team"
	if err := (models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &user.UserID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	user := c.MustGet("projectUser").(models.User)

	if _, err := database.Mysql.Exec("delete from project__user where user_id=? and project_id=?", user.ID, project.ID); err != nil {
		panic(err)
	}

	objType := "user"
	desc := "User ID " + strconv.Itoa(user.ID) + " removed from team"
	if err := (models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &user.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func MakeUserAdmin(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	user := c.MustGet("projectUser").(models.User)
	admin := 1

	if c.Request.Method == "DELETE" {
		// strip admin
		admin = 0
	}

	if _, err := database.Mysql.Exec("update project__user set admin=? where user_id=? and project_id=?", admin, user.ID, project.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
