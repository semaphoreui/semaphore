package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) UserEvents(ctx context.Context, params model.EventParams) ([]*model.Event, error) {
	return nil, nil
}

func (db *BoltdbStore) ProjectEvents(ctx context.Context, params model.EventParams) ([]*model.Event, error) {
	return nil, nil
}

func (db *BoltdbStore) CreateEvent(ctx context.Context, record *model.Event) error {
	return nil
}
