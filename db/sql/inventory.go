package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	err = d.getObject(projectID, db.InventoryProps, inventoryID, &inventory)
	if err != nil {
		return
	}

	err = db.FillInventory(d, &inventory)
	return
}

func (d *SqlDb) GetInventories(projectID int, params db.RetrieveQueryParams, types []db.InventoryType) ([]db.Inventory, error) {
	var inventories []db.Inventory
	err := d.getObjects(projectID, db.InventoryProps, params, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		if len(types) == 0 {
			return builder
		}

		return builder.Where(squirrel.Eq{"type": types})
	}, &inventories)
	return inventories, err
}

func (d *SqlDb) GetInventoryRefs(projectID int, inventoryID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.InventoryProps, inventoryID)
}

func (d *SqlDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, db.InventoryProps, inventoryID)
}

func (d *SqlDb) UpdateInventory(inventory db.Inventory) error {

	_, err := d.exec(
		"update project__inventory set name=?, type=?, ssh_key_id=?, inventory=?, become_key_id=?, template_id=?, repository_id=? where id=?",
		inventory.Name,
		inventory.Type,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.BecomeKeyID,
		inventory.TemplateID,
		inventory.RepositoryID,
		inventory.ID)

	return err
}

func (d *SqlDb) CreateInventory(inventory db.Inventory) (newInventory db.Inventory, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__inventory (project_id, name, type, ssh_key_id, inventory, become_key_id, template_id, repository_id) values "+
			"(?, ?, ?, ?, ?, ?, ?, ?)",
		inventory.ProjectID,
		inventory.Name,
		inventory.Type,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.BecomeKeyID,
		inventory.TemplateID,
		inventory.RepositoryID)

	if err != nil {
		return
	}

	newInventory = inventory
	newInventory.ID = insertID
	return
}
