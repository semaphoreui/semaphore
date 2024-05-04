package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) ListProjects(ctx context.Context, params model.ProjectParams) ([]*model.Project, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowProject(ctx context.Context, params model.ProjectParams) (*model.Project, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteProject(ctx context.Context, params model.ProjectParams) error {
	return nil
}

func (db *BoltdbStore) CreateProject(ctx context.Context, record *model.Project) error {
	return nil
}

func (db *BoltdbStore) UpdateProject(ctx context.Context, record *model.Project) error {
	return nil
}
