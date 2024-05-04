package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) AdminUsers(ctxy context.Context, params model.UserParams) ([]*model.User, error) {
	return nil, nil
}

func (db *BoltdbStore) ListUsers(ctx context.Context, params model.UserParams) ([]*model.User, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowUser(ctx context.Context, params model.UserParams) (*model.User, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteUser(ctx context.Context, params model.UserParams) error {
	return nil
}

func (db *BoltdbStore) CreateUser(ctx context.Context, record *model.User) error {
	return nil
}

func (db *BoltdbStore) UpdateUser(ctx context.Context, record *model.User) error {
	return nil
}

func (db *BoltdbStore) UpdatePassword(ctx context.Context, params model.UserParams) error {
	return nil
}
