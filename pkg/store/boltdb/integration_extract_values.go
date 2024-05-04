package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) IntegrationExtractValueRefs(ctx context.Context, params model.IntegrationExtractValueParams) (*model.IntegrationExtractorChildReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) ListIntegrationExtractValues(ctx context.Context, params model.IntegrationExtractValueParams) ([]*model.IntegrationExtractValue, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams) (*model.IntegrationExtractValue, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams) error {
	return nil
}

func (db *BoltdbStore) CreateIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams, record *model.IntegrationExtractValue) error {
	return nil
}

func (db *BoltdbStore) UpdateIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams, record *model.IntegrationExtractValue) error {
	return nil
}
