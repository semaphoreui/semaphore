package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) IntegrationRefs(ctx context.Context, params model.IntegrationParams) (*model.IntegrationReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) SearchableIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error) {
	return nil, nil
}

func (db *GormdbStore) AliasedIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error) {
	return nil, nil
}

func (db *GormdbStore) ListIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error) {
	return nil, nil
}

func (db *GormdbStore) ShowIntegration(ctx context.Context, params model.IntegrationParams) (*model.Integration, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteIntegration(ctx context.Context, params model.IntegrationParams) error {
	return nil
}

func (db *GormdbStore) CreateIntegration(ctx context.Context, record *model.Integration) error {
	return nil
}

func (db *GormdbStore) UpdateIntegration(ctx context.Context, record *model.Integration) error {
	return nil
}
