package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) IntegrationExtractValueRefs(ctx context.Context, params model.IntegrationExtractValueParams) (*model.IntegrationExtractorChildReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) ListIntegrationExtractValues(ctx context.Context, params model.IntegrationExtractValueParams) ([]*model.IntegrationExtractValue, error) {
	return nil, nil
}

func (db *GormdbStore) ShowIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams) (*model.IntegrationExtractValue, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams) error {
	return nil
}

func (db *GormdbStore) CreateIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams, record *model.IntegrationExtractValue) error {
	return nil
}

func (db *GormdbStore) UpdateIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams, record *model.IntegrationExtractValue) error {
	return nil
}
