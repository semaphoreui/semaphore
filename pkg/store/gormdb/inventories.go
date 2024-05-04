package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) InventoryRefs(ctx context.Context, params model.InventoryParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) ListInventories(ctx context.Context, params model.InventoryParams) ([]*model.Inventory, error) {
	return nil, nil
}

func (db *GormdbStore) ShowInventory(ctx context.Context, params model.InventoryParams) (*model.Inventory, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteInventory(ctx context.Context, params model.InventoryParams) error {
	return nil
}

func (db *GormdbStore) CreateInventory(ctx context.Context, record *model.Inventory) error {
	return nil
}

func (db *GormdbStore) UpdateInventory(ctx context.Context, record *model.Inventory) error {
	return nil
}
