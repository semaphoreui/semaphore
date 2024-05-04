package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) ListViews(ctx context.Context, params model.ViewParams) ([]*model.View, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowView(ctx context.Context, params model.ViewParams) (*model.View, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteView(ctx context.Context, params model.ViewParams) error {
	return nil
}

func (db *BoltdbStore) PositionView(ctx context.Context, params model.ViewParams) error {
	return nil
}

func (db *BoltdbStore) CreateView(ctx context.Context, record *model.View) error {
	return nil
}

func (db *BoltdbStore) UpdateView(ctx context.Context, record *model.View) error {
	return nil
}
