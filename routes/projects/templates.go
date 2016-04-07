package projects

import (
	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func TemplatesMiddleware(c *gin.Context) {
	c.AbortWithStatus(501)
}

func GetTemplates(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var templates []models.Template

	q := squirrel.Select("*").
		From("project__template").
		Where("project_id=?", project.ID)

	query, args, _ := q.ToSql()

	if _, err := database.Mysql.Select(&templates, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, templates)
}

func AddTemplate(c *gin.Context) {
	c.AbortWithStatus(501)
}

func UpdateTemplate(c *gin.Context) {
	c.AbortWithStatus(501)
}

func RemoveTemplate(c *gin.Context) {
	c.AbortWithStatus(501)
}
