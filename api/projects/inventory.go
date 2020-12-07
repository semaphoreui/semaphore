package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/context"
)

const (
	asc  = "asc"
	desc = "desc"
)

// InventoryMiddleware ensures an inventory exists and loads it to the context
func InventoryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		inventoryID, err := helpers.GetIntParam("inventory_id", w, r)
		if err != nil {
			return
		}

		inventory, err := helpers.Store(r).GetInventory(project.ID, inventoryID)

		if err != nil {
			helpers.WriteError(w, err)
		}

		context.Set(r, "inventory", inventory)
		next.ServeHTTP(w, r)
	})
}

// GetInventory returns an inventory from the database
func GetInventory(w http.ResponseWriter, r *http.Request) {
	if inventory := context.Get(r, "inventory"); inventory != nil {
		helpers.WriteJSON(w, http.StatusOK, inventory.(db.Inventory))
		return
	}

	project := context.Get(r, "project").(db.Project)


	params := db.RetrieveQueryParams{
		SortBy: r.URL.Query().Get("sort"),
		SortInverted: r.URL.Query().Get("order") == desc,
	}

	inventories, err := helpers.Store(r).GetInventories(project.ID, params)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, inventories)
}

// AddInventory creates an inventory in the database
func AddInventory(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var inventory db.Inventory

	if !helpers.Bind(w, r, &inventory) {
		return
	}

	if inventory.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	switch inventory.Type {
	case "static", "file":
		break
	default:
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Not supported inventory type",
		})
		return
	}

	newInventory, err := helpers.Store(r).CreateInventory(inventory)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	objType := "inventory"
	desc := "Inventory " + inventory.Name + " created"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &newInventory.ID,
		Description: &desc,
	})

	if err != nil {
		// Write error to log but return ok to user, because inventory created
		log.Error(err)
	}

	helpers.WriteJSON(w, http.StatusCreated, newInventory)
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

	var inventory db.Inventory

	if !helpers.Bind(w, r, &inventory) {
		return
	}

	if inventory.ID != oldInventory.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Inventory ID in body and URL must be the same",
		})
		return
	}

	if inventory.ProjectID != oldInventory.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
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

	err := helpers.Store(r).UpdateInventory(inventory)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveInventory deletes an inventory from the database
func RemoveInventory(w http.ResponseWriter, r *http.Request) {
	inventory := context.Get(r, "inventory").(db.Inventory)

	templatesC, err := helpers.Store(r).Sql().SelectInt("select count(1) from project__template where project_id=? and inventory_id=?", inventory.ProjectID, inventory.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": "Inventory is in use by one or more templates",
				"inUse": true,
			})

			return
		}

		if _, err := helpers.Store(r).Sql().Exec("update project__inventory set removed=1 where id=?", inventory.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := helpers.Store(r).Sql().Exec("delete from project__inventory where id=?", inventory.ID); err != nil {
		panic(err)
	}

	desc := "Inventory " + inventory.Name + " deleted"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		ProjectID:   &inventory.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
