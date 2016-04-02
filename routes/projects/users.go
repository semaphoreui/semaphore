package projects

import (
	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func GetProjectUsers(c *gin.Context) {
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

func AddProjectUser(c *gin.Context) {
}

func RemoveProjectUser(c *gin.Context) {
}
