package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) ListViews(ctx context.Context, params model.ViewParams) ([]*model.View, error) {
	return nil, nil
}

func (db *GormdbStore) ShowView(ctx context.Context, params model.ViewParams) (*model.View, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteView(ctx context.Context, params model.ViewParams) error {
	return nil
}

func (db *GormdbStore) PositionView(ctx context.Context, params model.ViewParams) error {
	return nil
}

func (db *GormdbStore) CreateView(ctx context.Context, record *model.View) error {
	return nil
}

func (db *GormdbStore) UpdateView(ctx context.Context, record *model.View) error {
	return nil
}
