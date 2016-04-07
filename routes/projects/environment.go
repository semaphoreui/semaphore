package projects

import (
	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func EnvironmentMiddleware(c *gin.Context) {
	c.AbortWithStatus(501)
}

func GetEnvironment(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var env []models.Environment

	q := squirrel.Select("*").
		From("project__environment").
		Where("project_id=?", project.ID)

	query, args, _ := q.ToSql()

	if _, err := database.Mysql.Select(&env, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, env)
}

func AddEnvironment(c *gin.Context) {
	c.AbortWithStatus(501)
}

func RemoveEnvironment(c *gin.Context) {
	c.AbortWithStatus(501)
}
