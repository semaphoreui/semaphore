package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) IntegrationMatcherRefs(ctx context.Context, params model.IntegrationMatcherParams) (*model.IntegrationExtractorChildReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) ListIntegrationMatchers(ctx context.Context, params model.IntegrationMatcherParams) ([]*model.IntegrationMatcher, error) {
	return nil, nil
}

func (db *GormdbStore) ShowIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams) (*model.IntegrationMatcher, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams) error {
	return nil
}

func (db *GormdbStore) CreateIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams, record *model.IntegrationMatcher) error {
	return nil
}

func (db *GormdbStore) UpdateIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams, record *model.IntegrationMatcher) error {
	return nil
}
