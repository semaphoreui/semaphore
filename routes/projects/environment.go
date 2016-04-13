package projects

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func EnvironmentMiddleware(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	envID, err := util.GetIntParam("environment_id", c)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("*").
		From("project__environment").
		Where("project_id=?", project.ID).
		Where("id=?", envID).
		ToSql()

	var env models.Environment
	if err := database.Mysql.SelectOne(&env, query, args...); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("environment", env)
	c.Next()
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

func UpdateEnvironment(c *gin.Context) {
	oldEnv := c.MustGet("environment").(models.Environment)
	var env models.Environment
	if err := c.Bind(&env); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update project__environment set name=?, json=? where id=?", env.Name, env.JSON, oldEnv.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func AddEnvironment(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var env models.Environment

	if err := c.Bind(&env); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("insert into project__environment set project_id=?, name=?, json=?, password=?", project.ID, env.Name, env.JSON, env.Password); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveEnvironment(c *gin.Context) {
	env := c.MustGet("environment").(models.Environment)

	if _, err := database.Mysql.Exec("delete from project__environment where id=?", env.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
