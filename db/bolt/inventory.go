package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	err = d.getObject(projectID, db.InventoryProps, intObjectID(inventoryID), &inventory)

	if err != nil {
		return
	}

	err = db.FillInventory(d, &inventory)
	return
}

func (d *BoltDb) GetInventories(projectID int, params db.RetrieveQueryParams) (inventories []db.Inventory, err error) {
	err = d.getObjects(projectID, db.InventoryProps, params, nil, &inventories)
	return
}

func (d *BoltDb) GetInventoryRefs(projectID int, inventoryID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.InventoryProps, inventoryID)
}

func (d *BoltDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, db.InventoryProps, intObjectID(inventoryID), nil)
}

func (d *BoltDb) UpdateInventory(inventory db.Inventory) error {
	return d.updateObject(inventory.ProjectID, db.InventoryProps, inventory)
}

func (d *BoltDb) CreateInventory(inventory db.Inventory) (db.Inventory, error) {
	newInventory, err := d.createObject(inventory.ProjectID, db.InventoryProps, inventory)
	return newInventory.(db.Inventory), err
}
