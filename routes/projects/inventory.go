package projects

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func InventoryMiddleware(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	inventoryID, err := util.GetIntParam("inventory_id", c)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("*").
		From("project__inventory").
		Where("project_id=?", project.ID).
		Where("id=?", inventoryID).
		ToSql()

	var inventory models.Inventory
	if err := database.Mysql.SelectOne(&inventory, query, args...); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("inventory", inventory)
	c.Next()
}

func GetInventory(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var inv []models.Inventory

	query, args, _ := squirrel.Select("*").
		From("project__inventory").
		Where("project_id=?", project.ID).
		ToSql()

	if _, err := database.Mysql.Select(&inv, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, inv)
}

func AddInventory(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var inventory models.Inventory

	if err := c.Bind(&inventory); err != nil {
		return
	}

	switch inventory.Type {
	case "static", "aws", "do", "gcloud":
		break
	default:
		c.AbortWithStatus(400)
		return
	}

	if _, err := database.Mysql.Exec("insert into project__inventory set project_id=?, type=?, key_id=?, inventory=?", project.ID, inventory.Type, inventory.KeyID, inventory.Inventory); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func UpdateInventory(c *gin.Context) {
	c.AbortWithStatus(501)
}

func RemoveInventory(c *gin.Context) {
	inventory := c.MustGet("inventory").(models.Inventory)

	if _, err := database.Mysql.Exec("delete from project__inventory where id=?", inventory.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
