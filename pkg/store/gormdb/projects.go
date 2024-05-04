package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) ListProjects(ctx context.Context, params model.ProjectParams) ([]*model.Project, error) {
	return nil, nil
}

func (db *GormdbStore) ShowProject(ctx context.Context, params model.ProjectParams) (*model.Project, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteProject(ctx context.Context, params model.ProjectParams) error {
	return nil
}

func (db *GormdbStore) CreateProject(ctx context.Context, record *model.Project) error {
	return nil
}

func (db *GormdbStore) UpdateProject(ctx context.Context, record *model.Project) error {
	return nil
}
