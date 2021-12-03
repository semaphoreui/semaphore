package sql

import "github.com/neo1908/semaphore/db"

func (d *SqlDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	err = d.getObject(projectID, db.InventoryProps, inventoryID, &inventory)
	if err != nil {
		return
	}

	err = db.FillInventory(d, &inventory)
	return
}

func (d *SqlDb) GetInventories(projectID int, params db.RetrieveQueryParams) ([]db.Inventory, error) {
	var inventories []db.Inventory
	err := d.getObjects(projectID, db.InventoryProps, params, &inventories)
	return inventories, err
}

func (d *SqlDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, db.InventoryProps, inventoryID)
}

func (d *SqlDb) DeleteInventorySoft(projectID int, inventoryID int) error {
	return d.deleteObjectSoft(projectID, db.InventoryProps, inventoryID)
}

func (d *SqlDb) UpdateInventory(inventory db.Inventory) error {
	_, err := d.exec(
		"update project__inventory set name=?, type=?, ssh_key_id=?, inventory=?, become_key_id=? where id=?",
		inventory.Name,
		inventory.Type,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.BecomeKeyID,
		inventory.ID)

	return err
}

func (d *SqlDb) CreateInventory(inventory db.Inventory) (newInventory db.Inventory, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__inventory (project_id, name, type, ssh_key_id, inventory, become_key_id) values (?, ?, ?, ?, ?, ?)",
		inventory.ProjectID,
		inventory.Name,
		inventory.Type,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.BecomeKeyID)

	if err != nil {
		return
	}

	newInventory = inventory
	newInventory.ID = insertID
	return
}



