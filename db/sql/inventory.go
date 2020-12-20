package sql

import "github.com/ansible-semaphore/semaphore/db"

var inventoryObject = objectProperties{
	TableName: "project__inventory",
	SortableColumns: []string{"name"},
	TemplateColumnName: "inventory_id",
}

func (d *SqlDb) GetInventory(projectID int, inventoryID int) (db.Inventory, error) {
	var inventory db.Inventory
	err := d.getObject(projectID, inventoryObject, inventoryID, &inventory)
	return inventory, err
}

func (d *SqlDb) GetInventories(projectID int, params db.RetrieveQueryParams) ([]db.Inventory, error) {
	var inventories []db.Inventory
	err := d.getObjects(projectID, inventoryObject, params, &inventories)
	return inventories, err
}

func (d *SqlDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, inventoryObject, inventoryID)
}

func (d *SqlDb) DeleteInventorySoft(projectID int, inventoryID int) error {
	return d.deleteObjectSoft(projectID, inventoryObject,  inventoryID)
}

func (d *SqlDb) UpdateInventory(inventory db.Inventory) error {
	_, err := d.sql.Exec(
		"update project__inventory set name=?, type=?, key_id=?, ssh_key_id=?, inventory=? where id=?",
		inventory.Name,
		inventory.Type,
		inventory.KeyID,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.ID)

	return err
}

func (d *SqlDb) CreateInventory(inventory db.Inventory) (newInventory db.Inventory, err error) {
	res, err := d.sql.Exec(
		"insert into project__inventory set project_id=?, name=?, type=?, key_id=?, ssh_key_id=?, inventory=?",
		inventory.ProjectID,
		inventory.Name,
		inventory.Type,
		inventory.KeyID,
		inventory.SSHKeyID,
		inventory.Inventory)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newInventory = inventory
	newInventory.ID = int(insertID)
	return
}



