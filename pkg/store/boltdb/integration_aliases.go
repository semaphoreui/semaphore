package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) ListIntegrationAliases(ctx context.Context, params model.IntegrationAliasParams) ([]*model.IntegrationAlias, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteIntegrationAlias(ctx context.Context, params model.IntegrationAliasParams) error {
	return nil
}

func (db *BoltdbStore) CreateIntegrationAlias(ctx context.Context, record *model.IntegrationAlias) error {
	return nil
}
