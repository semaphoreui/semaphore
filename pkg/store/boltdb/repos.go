package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) RepoRefs(ctx context.Context, params model.RepoParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) ListRepos(ctx context.Context, params model.RepoParams) ([]*model.Repository, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowRepo(ctx context.Context, params model.RepoParams) (*model.Repository, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteRepo(ctx context.Context, params model.RepoParams) error {
	return nil
}

func (db *BoltdbStore) CreateRepo(ctx context.Context, record *model.Repository) error {
	return nil
}

func (db *BoltdbStore) UpdateRepo(ctx context.Context, record *model.Repository) error {
	return nil
}
