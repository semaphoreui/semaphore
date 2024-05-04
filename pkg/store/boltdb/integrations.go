package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) IntegrationRefs(ctx context.Context, params model.IntegrationParams) (*model.IntegrationReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) SearchableIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error) {
	return nil, nil
}

func (db *BoltdbStore) AliasedIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error) {
	return nil, nil
}

func (db *BoltdbStore) ListIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowIntegration(ctx context.Context, params model.IntegrationParams) (*model.Integration, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteIntegration(ctx context.Context, params model.IntegrationParams) error {
	return nil
}

func (db *BoltdbStore) CreateIntegration(ctx context.Context, record *model.Integration) error {
	return nil
}

func (db *BoltdbStore) UpdateIntegration(ctx context.Context, record *model.Integration) error {
	return nil
}
