package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) RepoRefs(ctx context.Context, params model.RepoParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) ListRepos(ctx context.Context, params model.RepoParams) ([]*model.Repository, error) {
	return nil, nil
}

func (db *GormdbStore) ShowRepo(ctx context.Context, params model.RepoParams) (*model.Repository, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteRepo(ctx context.Context, params model.RepoParams) error {
	return nil
}

func (db *GormdbStore) CreateRepo(ctx context.Context, record *model.Repository) error {
	return nil
}

func (db *GormdbStore) UpdateRepo(ctx context.Context, record *model.Repository) error {
	return nil
}
