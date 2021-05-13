package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	err = d.getObject(projectID, db.InventoryProps, inventoryID, &inventory)
	if err != nil {
		return
	}

	if inventory.KeyID != nil {
		inventory.Key, err = d.GetAccessKey(projectID, *inventory.KeyID)
		if err != nil {
			return
		}
	}

	if inventory.SSHKeyID != nil {
		inventory.SSHKey, err = d.GetAccessKey(projectID, *inventory.SSHKeyID)
	}

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
	return d.deleteObjectSoft(projectID, db.InventoryProps,  inventoryID)
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



