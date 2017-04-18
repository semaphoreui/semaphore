package projects

import (
	"database/sql"

	database "github.com/ansible-semaphore/semaphore/db"
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

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		c.JSON(400, map[string]string{
			"error": "JSON is not valid",
		})
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

	var js map[string]interface{}
	if json.Unmarshal([]byte(env.JSON), &js) != nil {
		c.JSON(400, map[string]string{
			"error": "JSON is not valid",
		})
		return
	}

	res, err := database.Mysql.Exec("insert into project__environment set project_id=?, name=?, json=?, password=?", project.ID, env.Name, env.JSON, env.Password)
	if err != nil {
		panic(err)
	}

	insertID, _ := res.LastInsertId()
	insertIDInt := int(insertID)
	objType := "environment"

	desc := "Environment " + env.Name + " created"
	if err := (models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &insertIDInt,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveEnvironment(c *gin.Context) {
	env := c.MustGet("environment").(models.Environment)

	templatesC, err := database.Mysql.SelectInt("select count(1) from project__template where project_id=? and environment_id=?", env.ProjectID, env.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(c.Query("setRemoved")) == 0 {
			c.JSON(400, map[string]interface{}{
				"error": "Environment is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := database.Mysql.Exec("update project__environment set removed=1 where id=?", env.ID); err != nil {
			panic(err)
		}

		c.AbortWithStatus(204)
		return
	}

	if _, err := database.Mysql.Exec("delete from project__environment where id=?", env.ID); err != nil {
		panic(err)
	}

	desc := "Environment " + env.Name + " deleted"
	if err := (models.Event{
		ProjectID:   &env.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
