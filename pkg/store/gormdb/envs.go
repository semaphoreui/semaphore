package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) EnvRefs(ctx context.Context, params model.EnvParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) ListEnvs(ctx context.Context, params model.EnvParams) ([]*model.Environment, error) {
	return nil, nil
}

func (db *GormdbStore) ShowEnv(ctx context.Context, params model.EnvParams) (*model.Environment, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteEnv(ctx context.Context, params model.EnvParams) error {
	return nil
}

func (db *GormdbStore) CreateEnv(ctx context.Context, record *model.Environment) error {
	return nil
}

func (db *GormdbStore) UpdateEnv(ctx context.Context, record *model.Environment) error {
	return nil
}
