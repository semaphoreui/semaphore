package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) EnvRefs(ctx context.Context, params model.EnvParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) ListEnvs(ctx context.Context, params model.EnvParams) ([]*model.Environment, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowEnv(ctx context.Context, params model.EnvParams) (*model.Environment, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteEnv(ctx context.Context, params model.EnvParams) error {
	return nil
}

func (db *BoltdbStore) CreateEnv(ctx context.Context, record *model.Environment) error {
	return nil
}

func (db *BoltdbStore) UpdateEnv(ctx context.Context, record *model.Environment) error {
	return nil
}
