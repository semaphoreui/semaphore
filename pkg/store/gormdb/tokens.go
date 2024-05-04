package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) ListTokens(ctx context.Context, params model.TokenParams) ([]*model.APIToken, error) {
	return nil, nil
}

func (db *GormdbStore) ShowToken(ctx context.Context, params model.TokenParams) (*model.APIToken, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteToken(ctx context.Context, params model.TokenParams) error {
	return nil
}

func (db *GormdbStore) ExpireToken(ctx context.Context, params model.TokenParams) error {
	return nil
}

func (db *GormdbStore) CreateToken(ctx context.Context, record *model.APIToken) error {
	return nil
}
