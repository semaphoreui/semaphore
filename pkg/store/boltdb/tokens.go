package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) ListTokens(ctx context.Context, params model.TokenParams) ([]*model.APIToken, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowToken(ctx context.Context, params model.TokenParams) (*model.APIToken, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteToken(ctx context.Context, params model.TokenParams) error {
	return nil
}

func (db *BoltdbStore) ExpireToken(ctx context.Context, params model.TokenParams) error {
	return nil
}

func (db *BoltdbStore) CreateToken(ctx context.Context, record *model.APIToken) error {
	return nil
}
