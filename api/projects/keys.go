package projects

import (
	"database/sql"

	database "github.com/ansible-semaphore/semaphore/db"
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

	q := squirrel.Select("id, name, type, project_id, `key`, removed").
		From("access_key").
		Where("project_id=?", project.ID)

	if len(c.Query("type")) > 0 {
		switch c.Query("type") {
		case "static":
			q = q.Where("type=? or type=?", "ssh", "ssh-cert")
		case "ssh":
			q = q.Where("type=?", "ssh")
		default:
			panic("Unknown type query")
		}
//		q = q.Where("type=?", c.Query("type"))
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
	case "aws", "gcloud", "do":
		break
	case "ssh", "ssh-cert":
		if key.Secret == nil || len(*key.Secret) == 0 {
			c.JSON(400, map[string]string{
				"error": "SSH Secret empty",
			})
			return
		}
		if key.Type == "ssh-cert" {
			if key.Key == nil || len(*key.Key) == 0 {
				c.JSON(400, map[string]string{
					"error": "SSH Public key empty",
				})
				return
			}
		}
	default:
		c.JSON(400, map[string]string{
			"error": "Invalid key type",
		})
		return
	}

	secret := *key.Secret + "\n"

	res, err := database.Mysql.Exec("insert into access_key set name=?, type=?, project_id=?, `key`=?, secret=?", key.Name, key.Type, project.ID, key.Key, secret)
	if err != nil {
		panic(err)
	}

	insertID, _ := res.LastInsertId()
	insertIDInt := int(insertID)
	objType := "key"

	desc := "Access Key " + key.Name + " created"
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

func UpdateKey(c *gin.Context) {
	var key models.AccessKey
	oldKey := c.MustGet("accessKey").(models.AccessKey)

	if err := c.Bind(&key); err != nil {
		return
	}

	switch key.Type {
	case "aws", "gcloud", "do":
		break
	case "ssh", "ssh-cert":
		if key.Secret == nil || len(*key.Secret) == 0 {
			c.JSON(400, map[string]string{
				"error": "SSH Secret empty",
			})
			return
		}
		if key.Type == "ssh-cert" {
			if key.Key == nil || len(*key.Key) == 0 {
				c.JSON(400, map[string]string{
					"error": "SSH Public key empty",
				})
				return
			}
		}
	default:
		c.JSON(400, map[string]string{
			"error": "Invalid key type",
		})
		return
	}

	if key.Secret == nil || len(*key.Secret) == 0 {
		// override secret
		key.Secret = oldKey.Secret
	} else {
		secret := *key.Secret + "\n"
		key.Secret = &secret
	}

	if _, err := database.Mysql.Exec("update access_key set name=?, type=?, `key`=?, secret=? where id=?", key.Name, key.Type, key.Key, key.Secret, oldKey.ID); err != nil {
		panic(err)
	}

	desc := "Access Key " + key.Name + " updated"
	objType := "key"
	if err := (models.Event{
		ProjectID:   oldKey.ProjectID,
		Description: &desc,
		ObjectID:    &oldKey.ID,
		ObjectType:  &objType,
	}.Insert()); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveKey(c *gin.Context) {
	key := c.MustGet("accessKey").(models.AccessKey)

	templatesC, err := database.Mysql.SelectInt("select count(1) from project__template where project_id=? and ssh_key_id=?", *key.ProjectID, key.ID)
	if err != nil {
		panic(err)
	}

	inventoryC, err := database.Mysql.SelectInt("select count(1) from project__inventory where project_id=? and ssh_key_id=?", *key.ProjectID, key.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 || inventoryC > 0 {
		if len(c.Query("setRemoved")) == 0 {
			c.JSON(400, map[string]interface{}{
				"error": "Key is in use by one or more templates / inventory",
				"inUse": true,
			})

			return
		}

		if _, err := database.Mysql.Exec("update access_key set removed=1 where id=?", key.ID); err != nil {
			panic(err)
		}

		c.AbortWithStatus(204)
		return
	}

	if _, err := database.Mysql.Exec("delete from access_key where id=?", key.ID); err != nil {
		panic(err)
	}

	desc := "Access Key " + key.Name + " deleted"
	if err := (models.Event{
		ProjectID:   key.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
