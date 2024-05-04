package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) ListIntegrationAliases(ctx context.Context, params model.IntegrationAliasParams) ([]*model.IntegrationAlias, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteIntegrationAlias(ctx context.Context, params model.IntegrationAliasParams) error {
	return nil
}

func (db *GormdbStore) CreateIntegrationAlias(ctx context.Context, record *model.IntegrationAlias) error {
	return nil
}
