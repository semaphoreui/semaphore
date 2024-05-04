package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) AdminUsers(ctxy context.Context, params model.UserParams) ([]*model.User, error) {
	return nil, nil
}

func (db *GormdbStore) ListUsers(ctx context.Context, params model.UserParams) ([]*model.User, error) {
	return nil, nil
}

func (db *GormdbStore) ShowUser(ctx context.Context, params model.UserParams) (*model.User, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteUser(ctx context.Context, params model.UserParams) error {
	return nil
}

func (db *GormdbStore) CreateUser(ctx context.Context, record *model.User) error {
	return nil
}

func (db *GormdbStore) UpdateUser(ctx context.Context, record *model.User) error {
	return nil
}

func (db *GormdbStore) UpdatePassword(ctx context.Context, params model.UserParams) error {
	return nil
}
