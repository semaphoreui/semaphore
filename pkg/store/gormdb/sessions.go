package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) ShowSession(ctx context.Context, params model.SessionParams) (*model.Session, error) {
	return nil, nil
}

func (db *GormdbStore) ExpireSession(ctx context.Context, params model.SessionParams) error {
	return nil
}

func (db *GormdbStore) TouchSession(ctx context.Context, params model.SessionParams) error {
	return nil
}

func (db *GormdbStore) CreateSession(ctx context.Context, record *model.Session) error {
	return nil
}
