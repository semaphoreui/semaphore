package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) IntegrationMatcherRefs(ctx context.Context, params model.IntegrationMatcherParams) (*model.IntegrationExtractorChildReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) ListIntegrationMatchers(ctx context.Context, params model.IntegrationMatcherParams) ([]*model.IntegrationMatcher, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams) (*model.IntegrationMatcher, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams) error {
	return nil
}

func (db *BoltdbStore) CreateIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams, record *model.IntegrationMatcher) error {
	return nil
}

func (db *BoltdbStore) UpdateIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams, record *model.IntegrationMatcher) error {
	return nil
}
