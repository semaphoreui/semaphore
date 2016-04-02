package projects

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func ProjectMiddleware(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	projectID, err := util.GetIntParam("project_id", c)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("p.id=?", projectID).
		Where("pu.user_id=?", user.ID).
		ToSql()

	var project models.Project
	if err := database.Mysql.SelectOne(&project, query, args...); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("project", project)
	c.Next()
}

func GetProject(c *gin.Context) {
	c.JSON(200, c.MustGet("project"))
}
