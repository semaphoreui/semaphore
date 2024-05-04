package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) InventoryRefs(ctx context.Context, params model.InventoryParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) ListInventories(ctx context.Context, params model.InventoryParams) ([]*model.Inventory, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowInventory(ctx context.Context, params model.InventoryParams) (*model.Inventory, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteInventory(ctx context.Context, params model.InventoryParams) error {
	return nil
}

func (db *BoltdbStore) CreateInventory(ctx context.Context, record *model.Inventory) error {
	return nil
}

func (db *BoltdbStore) UpdateInventory(ctx context.Context, record *model.Inventory) error {
	return nil
}
