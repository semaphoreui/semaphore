package projects

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"

	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/context"
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
			return
		}

		context.Set(r, "inventory", inventory)
		next.ServeHTTP(w, r)
	})
}

func GetInventoryRefs(w http.ResponseWriter, r *http.Request) {
	inventory := context.Get(r, "inventory").(db.Inventory)
	refs, err := helpers.Store(r).GetInventoryRefs(inventory.ProjectID, inventory.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// GetInventory returns an inventory from the database
func GetInventory(w http.ResponseWriter, r *http.Request) {
	if inventory := context.Get(r, "inventory"); inventory != nil {
		helpers.WriteJSON(w, http.StatusOK, inventory.(db.Inventory))
		return
	}

	project := context.Get(r, "project").(db.Project)

	inventories, err := helpers.Store(r).GetInventories(project.ID, helpers.QueryParams(r.URL))

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
	case db.InventoryStatic, db.InventoryStaticYaml, db.InventoryFile, db.InventoryTerraformWorkspace:
		break
	default:
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Not supported inventory type",
		})
		return
	}

	err := db.ValidateInventory(helpers.Store(r), &inventory)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	newInventory, err := helpers.Store(r).CreateInventory(inventory)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventInventory,
		ObjectID:    newInventory.ID,
		Description: fmt.Sprintf("Inventory %s created", inventory.Name),
	})

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
		helpers.WriteErrorStatus(w,
			"Inventory ID in body and URL must be the same",
			http.StatusBadRequest)
		return
	}

	if inventory.ProjectID != oldInventory.ProjectID {
		helpers.WriteErrorStatus(w,
			"project ID in body and URL must be the same",
			http.StatusBadRequest)
		return
	}

	switch inventory.Type {
	case db.InventoryStatic, db.InventoryStaticYaml:
		break
	case db.InventoryFile:
		if !IsValidInventoryPath(inventory.Inventory) {
			helpers.WriteErrorStatus(w, "Invalid inventory file pathname. Must be: path/to/inventory.", http.StatusBadRequest)
			return
		}
	case db.InventoryTerraformWorkspace:
		break
	default:
		helpers.WriteErrorStatus(w,
			"unknown inventory type: "+string(inventory.Type),
			http.StatusBadRequest)
		return
	}

	if err := db.ValidateInventory(helpers.Store(r), &inventory); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).UpdateInventory(inventory); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldInventory.ProjectID,
		ObjectType:  db.EventInventory,
		ObjectID:    oldInventory.ID,
		Description: fmt.Sprintf("Inventory %s updated", inventory.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveInventory deletes an inventory from the database
func RemoveInventory(w http.ResponseWriter, r *http.Request) {
	inventory := context.Get(r, "inventory").(db.Inventory)
	var err error

	err = helpers.Store(r).DeleteInventory(inventory.ProjectID, inventory.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Inventory is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   inventory.ProjectID,
		ObjectType:  db.EventInventory,
		ObjectID:    inventory.ID,
		Description: fmt.Sprintf("Inventory %s deleted", inventory.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}
