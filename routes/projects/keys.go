package projects

import (
	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func KeyMiddleware(c *gin.Context) {
	c.Next()
}

func GetKeys(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var keys []models.AccessKey

	q := squirrel.Select("*").
		From("access_key").
		Where("project_id=?", project.ID)

	if len(c.Query("type")) > 0 {
		q = q.Where("type=?", c.Query("type"))
	}

	query, args, _ := q.ToSql()

	if _, err := database.Mysql.Select(&keys, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, keys)
}

func AddKey(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var key models.AccessKey

	if err := c.Bind(&key); err != nil {
		return
	}

	switch key.Type {
	case "aws", "gcloud", "do", "ssh":
		break
	default:
		c.AbortWithStatus(400)
		return
	}

	if _, err := database.Mysql.Exec("insert into access_key set name=?, type=?, project_id=?, `key`=?, secret=?", key.Name, key.Type, project.ID, key.Key, key.Secret); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func UpdateKey(c *gin.Context) {
	c.AbortWithStatus(501)
}

func RemoveKey(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	keyID, err := util.GetIntParam("key_id", c)
	if err != nil {
		return
	}

	if _, err := database.Mysql.Exec("delete from access_key where project_id=? and id=?", project.ID, keyID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
