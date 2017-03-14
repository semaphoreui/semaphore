package projects

import (
	"database/sql"
	"net/http"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func InventoryMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	inventoryID, err := util.GetIntParam("inventory_id", w, r)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("*").
		From("project__inventory").
		Where("project_id=?", project.ID).
		Where("id=?", inventoryID).
		ToSql()

	var inventory db.Inventory
	if err := db.Mysql.SelectOne(&inventory, query, args...); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "inventory", inventory)
}

func GetInventory(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var inv []db.Inventory

	query, args, _ := squirrel.Select("*").
		From("project__inventory").
		Where("project_id=?", project.ID).
		OrderBy("name asc").
		ToSql()

	if _, err := db.Mysql.Select(&inv, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, inv)
}

func AddInventory(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var inventory struct {
		Name      string `json:"name" binding:"required"`
		KeyID     *int   `json:"key_id"`
		SshKeyID  int    `json:"ssh_key_id"`
		Type      string `json:"type"`
		Inventory string `json:"inventory"`
	}

	if err := mulekick.Bind(w, r, &inventory); err != nil {
		return
	}

	switch inventory.Type {
	case "static", "aws", "do", "gcloud":
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := db.Mysql.Exec("insert into project__inventory set project_id=?, name=?, type=?, key_id=?, ssh_key_id=?, inventory=?", project.ID, inventory.Name, inventory.Type, inventory.KeyID, inventory.SshKeyID, inventory.Inventory)
	if err != nil {
		panic(err)
	}

	insertID, _ := res.LastInsertId()
	insertIDInt := int(insertID)
	objType := "inventory"

	desc := "Inventory " + inventory.Name + " created"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &insertIDInt,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	inv := db.Inventory{
		ID: insertIDInt,
		Name: inventory.Name,
		ProjectID: project.ID,
		Inventory: inventory.Inventory,
		KeyID: inventory.KeyID,
		SshKeyID: &inventory.SshKeyID,
		Type: inventory.Type,
	}

	mulekick.WriteJSON(w, http.StatusCreated, inv)
}

func UpdateInventory(w http.ResponseWriter, r *http.Request) {
	oldInventory := context.Get(r, "inventory").(db.Inventory)

	var inventory struct {
		Name      string `json:"name" binding:"required"`
		KeyID     *int   `json:"key_id"`
		SshKeyID  int    `json:"ssh_key_id"`
		Type      string `json:"type"`
		Inventory string `json:"inventory"`
	}

	if err := mulekick.Bind(w, r, &inventory); err != nil {
		return
	}

	switch inventory.Type {
	case "static", "aws", "do", "gcloud":
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := db.Mysql.Exec("update project__inventory set name=?, type=?, key_id=?, ssh_key_id=?, inventory=? where id=?", inventory.Name, inventory.Type, inventory.KeyID, inventory.SshKeyID, inventory.Inventory, oldInventory.ID); err != nil {
		panic(err)
	}

	desc := "Inventory " + inventory.Name + " updated"
	objType := "inventory"
	if err := (db.Event{
		ProjectID:   &oldInventory.ProjectID,
		Description: &desc,
		ObjectID:    &oldInventory.ID,
		ObjectType:  &objType,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveInventory(w http.ResponseWriter, r *http.Request) {
	inventory := context.Get(r, "inventory").(db.Inventory)

	templatesC, err := db.Mysql.SelectInt("select count(1) from project__template where project_id=? and inventory_id=?", inventory.ProjectID, inventory.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			mulekick.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Inventory is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := db.Mysql.Exec("update project__inventory set removed=1 where id=?", inventory.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.Mysql.Exec("delete from project__inventory where id=?", inventory.ID); err != nil {
		panic(err)
	}

	desc := "Inventory " + inventory.Name + " deleted"
	if err := (db.Event{
		ProjectID:   &inventory.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
