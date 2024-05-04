package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) UserEvents(ctx context.Context, params model.EventParams) ([]*model.Event, error) {
	return nil, nil
}

func (db *GormdbStore) ProjectEvents(ctx context.Context, params model.EventParams) ([]*model.Event, error) {
	return nil, nil
}

func (db *GormdbStore) CreateEvent(ctx context.Context, record *model.Event) error {
	return nil
}
