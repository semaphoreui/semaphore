package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)


func (d *BoltDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	err = d.getObject(projectID, db.InventoryObject, intObjectID(inventoryID), &inventory)

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

func (d *BoltDb) GetInventories(projectID int, params db.RetrieveQueryParams) (inventories []db.Inventory, err error) {
	err = d.getObjects(projectID, db.AccessKeyObject, params, &inventories)
	return
}

func (d *BoltDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, db.InventoryObject, intObjectID(inventoryID))
}

func (d *BoltDb) DeleteInventorySoft(projectID int, inventoryID int) error {
	return d.deleteObjectSoft(projectID, db.InventoryObject, intObjectID(inventoryID))
}

func (d *BoltDb) UpdateInventory(inventory db.Inventory) error {
	return d.updateObject(inventory.ProjectID, db.InventoryObject, inventory)
}

func (d *BoltDb) CreateInventory(inventory db.Inventory) (db.Inventory, error) {
	newInventory, err := d.createObject(inventory.ProjectID, db.InventoryObject, inventory)
	return newInventory.(db.Inventory), err
}



