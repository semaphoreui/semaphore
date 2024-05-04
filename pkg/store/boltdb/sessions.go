package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) ShowSession(ctx context.Context, params model.SessionParams) (*model.Session, error) {
	return nil, nil
}

func (db *BoltdbStore) ExpireSession(ctx context.Context, params model.SessionParams) error {
	return nil
}

func (db *BoltdbStore) TouchSession(ctx context.Context, params model.SessionParams) error {
	return nil
}

func (db *BoltdbStore) CreateSession(ctx context.Context, record *model.Session) error {
	return nil
}
