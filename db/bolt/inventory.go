package bolt

import (
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
	bolt "go.etcd.io/bbolt"
	"strconv"
)


func (d *BoltDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	id, err := makeObjectId("inventory", projectID)

	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(id)
		if b == nil {
			return db.ErrNotFound
		}

		id := []byte(strconv.Itoa(inventoryID))
		str := b.Get(id)
		if str == nil {
			return db.ErrNotFound
		}

		return json.Unmarshal(str, &inventory)
	})

	if err != nil {
		return
	}

	//if inventory.KeyID != nil {
	//	inventory.Key, err = d.GetAccessKey(projectID, *inventory.KeyID)
	//	if err != nil {
	//		return
	//	}
	//}
	//
	//if inventory.SSHKeyID != nil {
	//	inventory.SSHKey, err = d.GetAccessKey(projectID, *inventory.SSHKeyID)
	//}

	return
}

func (d *BoltDb) GetInventories(projectID int, params db.RetrieveQueryParams) (inventories []db.Inventory, err error) {
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("inventory_" + strconv.Itoa(projectID)))
		if b == nil {
			return db.ErrNotFound
		}

		return nil
	})

	return inventories, err
}

func (d *BoltDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.db.Update(func (tx *bolt.Tx) error {
		b := tx.Bucket([]byte("inventory_" + strconv.Itoa(projectID)))
		if b == nil {
			return db.ErrNotFound
		}
		return b.Delete([]byte(strconv.Itoa(inventoryID)))
	})
}

func (d *BoltDb) DeleteInventorySoft(projectID int, inventoryID int) error {
	return d.DeleteInventory(projectID, inventoryID)
}

func (d *BoltDb) UpdateInventory(inventory db.Inventory) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("inventory_" + strconv.Itoa(inventory.ProjectID)))
		if b == nil {
			return db.ErrNotFound
		}

		id := []byte(strconv.Itoa(inventory.ID))
		if b.Get(id) == nil {
			return db.ErrNotFound
		}

		str, err2 := json.Marshal(inventory)
		if err2 != nil {
			return err2
		}

		return b.Put(id, str)
	})

	return err
}

func (d *BoltDb) CreateInventory(inventory db.Inventory) (newInventory db.Inventory, err error) {
	err = d.db.Update(func(tx *bolt.Tx) error {
		b, err2 := tx.CreateBucketIfNotExists([]byte("inventory_" + strconv.Itoa(inventory.ProjectID)))
		if err2 != nil {
			return err2
		}

		id, _ := b.NextSequence()
		newInventory = inventory
		newInventory.ID = int(id)
		str, err2 := json.Marshal(newInventory)
		if err2 != nil {
			return err2
		}

		return b.Put([]byte(strconv.Itoa(newInventory.ID)), str)
	})

	return
}



