package projects

import (
	"database/sql"
	"net/http"

	"os"
	"path/filepath"
	"strings"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

const (
	asc  = "asc"
	desc = "desc"
)

// InventoryMiddleware ensures an inventory exists and loads it to the context
func InventoryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		inventoryID, err := util.GetIntParam("inventory_id", w, r)
		if err != nil {
			return
		}

		query, args, err := squirrel.Select("*").
			From("project__inventory").
			Where("project_id=?", project.ID).
			Where("id=?", inventoryID).
			ToSql()
		util.LogWarning(err)

		var inventory db.Inventory
		if err := db.Sql.SelectOne(&inventory, query, args...); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		context.Set(r, "inventory", inventory)
		next.ServeHTTP(w, r)
	})
}

// GetInventory returns an inventory from the database
func GetInventory(w http.ResponseWriter, r *http.Request) {
	if inventory := context.Get(r, "inventory"); inventory != nil {
		util.WriteJSON(w, http.StatusOK, inventory.(db.Inventory))
		return
	}

	project := context.Get(r, "project").(db.Project)

	var inv []db.Inventory

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != asc && order != desc {
		order = asc
	}

	q := squirrel.Select("*").
		From("project__inventory pi")

	switch sort {
	case "name", "type":
		q = q.Where("pi.project_id=?", project.ID).
			OrderBy("pi." + sort + " " + order)
	default:
		q = q.Where("pi.project_id=?", project.ID).
			OrderBy("pi.name " + order)
	}

	query, args, err := q.ToSql()
	util.LogWarning(err)

	if _, err := db.Sql.Select(&inv, query, args...); err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusOK, inv)
}

// AddInventory creates an inventory in the database
func AddInventory(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var inventory struct {
		Name      string `json:"name" binding:"required"`
		KeyID     *int   `json:"key_id"`
		SSHKeyID  int    `json:"ssh_key_id"`
		Type      string `json:"type"`
		Inventory string `json:"inventory"`
	}

	if err := util.Bind(w, r, &inventory); err != nil {
		return
	}

	switch inventory.Type {
	case "static", "file":
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := db.Sql.Exec("insert into project__inventory set project_id=?, name=?, type=?, key_id=?, ssh_key_id=?, inventory=?", project.ID, inventory.Name, inventory.Type, inventory.KeyID, inventory.SSHKeyID, inventory.Inventory)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	util.LogWarning(err)
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
		ID:        insertIDInt,
		Name:      inventory.Name,
		ProjectID: project.ID,
		Inventory: inventory.Inventory,
		KeyID:     inventory.KeyID,
		SSHKeyID:  &inventory.SSHKeyID,
		Type:      inventory.Type,
	}

	util.WriteJSON(w, http.StatusCreated, inv)
}

// IsValidInventoryPath tests a path to ensure it is below the cwd
func IsValidInventoryPath(path string) bool {

	currentPath, err := os.Getwd()
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(currentPath, absPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relPath, "..")
}

// UpdateInventory writes updated values to an existing inventory item in the database
func UpdateInventory(w http.ResponseWriter, r *http.Request) {
	oldInventory := context.Get(r, "inventory").(db.Inventory)

	var inventory struct {
		Name      string `json:"name" binding:"required"`
		KeyID     *int   `json:"key_id"`
		SSHKeyID  int    `json:"ssh_key_id"`
		Type      string `json:"type"`
		Inventory string `json:"inventory"`
	}

	if err := util.Bind(w, r, &inventory); err != nil {
		return
	}

	switch inventory.Type {
	case "static":
		break
	case "file":
		if !IsValidInventoryPath(inventory.Inventory) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := db.Sql.Exec("update project__inventory set name=?, type=?, key_id=?, ssh_key_id=?, inventory=? where id=?", inventory.Name, inventory.Type, inventory.KeyID, inventory.SSHKeyID, inventory.Inventory, oldInventory.ID); err != nil {
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

// RemoveInventory deletes an inventory from the database
func RemoveInventory(w http.ResponseWriter, r *http.Request) {
	inventory := context.Get(r, "inventory").(db.Inventory)

	templatesC, err := db.Sql.SelectInt("select count(1) from project__template where project_id=? and inventory_id=?", inventory.ProjectID, inventory.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Inventory is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := db.Sql.Exec("update project__inventory set removed=1 where id=?", inventory.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.Sql.Exec("delete from project__inventory where id=?", inventory.ID); err != nil {
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
