package projects

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func KeyMiddleware(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	keyID, err := util.GetIntParam("key_id", c)
	if err != nil {
		return
	}

	var key models.AccessKey
	if err := database.Mysql.SelectOne(&key, "select * from access_key where project_id=? and id=?", project.ID, keyID); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("accessKey", key)
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
	var key models.AccessKey
	oldKey := c.MustGet("accessKey").(models.AccessKey)

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

	if _, err := database.Mysql.Exec("update access_key set name=?, type=?, `key`=?, secret=?", key.Name, key.Type, key.Key, key.Secret, oldKey.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveKey(c *gin.Context) {
	key := c.MustGet("accessKey").(models.AccessKey)

	if _, err := database.Mysql.Exec("delete from access_key where id=?", key.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
